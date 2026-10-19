package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/netip"
	"os"
	"os/signal"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
	"time"

	"charm.land/fang/v2"
	"github.com/charmbracelet/x/term"
	"github.com/kloudkit/looking-glass/internal/reflect"
	"github.com/kloudkit/looking-glass/internal/ui"
	"github.com/spf13/cobra"
)

const (
	defaultPort     = "8080"
	defaultMaxBody  = 1 << 20 // 1 MiB
	defaultMaxDelay = 10 * time.Second
)

// version is stamped at build time with -ldflags "-X .../internal/cli.version=v1.2.3".
var version = "dev"

var releaseTag = regexp.MustCompile(`^v?\d+\.\d+\.\d+$`)

// resolveVersion prefers the stamped version, then a `go install …@vX.Y.Z` module version.
// Anything else (a pseudo-version, a dirty tree) is a development build.
func resolveVersion() string {
	if version != "dev" {
		return strings.TrimPrefix(version, "v")
	}
	if info, ok := debug.ReadBuildInfo(); ok && releaseTag.MatchString(info.Main.Version) {
		return strings.TrimPrefix(info.Main.Version, "v")
	}

	return "dev"
}

var rootCmd = &cobra.Command{
	Use:   "looking-glass",
	Short: "See exactly what is hitting your service",
	Long: "See exactly what is hitting your service.\n\n" +
		"Looking Glass answers every method on every path with the request it received.",
	Example: "# reflect on :9000, cutting bodies off at 64 KiB\n" +
		"looking-glass --port 9000 --max-body 65536\n\n" +
		"# let requests ask for up to a minute of X-Glass-Delay\n" +
		"looking-glass --max-delay 1m",
	Args: cobra.NoArgs,
	RunE: serve,
}

func init() {
	f := rootCmd.Flags()
	f.StringP("port", "p", envOr("PORT", defaultPort), "port to listen on (env: PORT)")
	f.Int64("max-body", envInt("MAX_BODY_BYTES", defaultMaxBody), "max body bytes reflected before cut-off (env: MAX_BODY_BYTES)")
	f.Duration("max-delay", envDuration("MAX_DELAY", defaultMaxDelay), "longest X-Glass-Delay a request may ask for (env: MAX_DELAY)")
	f.Bool("cors-credentials", envBool("CORS_CREDENTIALS"), "let cross-origin pages send cookies and read them back (env: CORS_CREDENTIALS)")
	f.StringSlice("trusted-proxies", envList("TRUSTED_PROXIES", defaultTrustedProxies()), "proxies whose X-Forwarded-For is believed, as IPs or CIDRs (env: TRUSTED_PROXIES)")
}

func Execute() {
	opts := []fang.Option{
		fang.WithColorSchemeFunc(ui.ColorScheme),
		fang.WithVersion(resolveVersion()),
		fang.WithoutManpage(),
		fang.WithoutCompletions(),
		fang.WithErrorHandler(printError),
	}
	if err := fang.Execute(context.Background(), rootCmd, opts...); err != nil {
		os.Exit(1)
	}
}

// printError is fang's default handler without title-casing, which mangles flag names.
func printError(w io.Writer, styles fang.Styles, err error) {
	if f, ok := w.(term.File); ok && !term.IsTerminal(f.Fd()) {
		fmt.Fprintln(w, err)
		return
	}

	fmt.Fprintln(w, styles.ErrorHeader.String())
	fmt.Fprintln(w, styles.ErrorText.UnsetTransform().Render(err.Error()))
	fmt.Fprintln(w)
}

func serve(cmd *cobra.Command, _ []string) error {
	port, _ := cmd.Flags().GetString("port")
	maxBody, _ := cmd.Flags().GetInt64("max-body")
	if maxBody <= 0 {
		return fmt.Errorf("--max-body must be positive, got %d", maxBody)
	}
	maxDelay, _ := cmd.Flags().GetDuration("max-delay")
	if maxDelay < 0 {
		return fmt.Errorf("--max-delay must not be negative, got %s", maxDelay)
	}
	credentials, _ := cmd.Flags().GetBool("cors-credentials")
	proxies, _ := cmd.Flags().GetStringSlice("trusted-proxies")
	trusted, err := parsePrefixes(proxies)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	mux := http.NewServeMux()
	mux.HandleFunc("/", reflect.Handler(ctx, reflect.Options{
		MaxBody:        maxBody,
		MaxDelay:       maxDelay,
		Credentials:    credentials,
		TrustedProxies: trusted,
	}))

	// HTTP/2 without TLS, for clients and proxies that speak it to a plain backend.
	var protocols http.Protocols
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)

	srv := &http.Server{
		Addr:      ":" + port,
		Handler:   mux,
		Protocols: &protocols,
	}

	failed := make(chan error, 1)
	go func() {
		log.Print(strings.Join([]string{
			ui.Left.Render("looking-glass"),
			ui.Muted.Render("listening on"),
			ui.Right.Render(":" + port),
			ui.Muted.Render(fmt.Sprintf("(max body %d bytes)", maxBody)),
		}, " "))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			failed <- err
		}
	}()

	select {
	case err := <-failed:
		return err
	case <-ctx.Done():
	}

	log.Print(ui.Muted.Render("shutting down"))

	shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return srv.Shutdown(shutdown)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

func envInt(key string, fallback int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			return n
		}
	}

	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d >= 0 {
			return d
		}
	}

	return fallback
}

func envBool(key string) bool {
	v, _ := strconv.ParseBool(os.Getenv(key))

	return v
}

// envList splits a comma-separated variable; set but empty, it is an empty list.
func envList(key string, fallback []string) []string {
	v, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	var list []string
	for item := range strings.SplitSeq(v, ",") {
		if item = strings.TrimSpace(item); item != "" {
			list = append(list, item)
		}
	}

	return list
}

func defaultTrustedProxies() []string {
	list := make([]string, len(reflect.PrivateNetworks))
	for i, prefix := range reflect.PrivateNetworks {
		list[i] = prefix.String()
	}

	return list
}

// parsePrefixes takes CIDRs, or bare IPs as single-address ranges.
func parsePrefixes(list []string) ([]netip.Prefix, error) {
	prefixes := make([]netip.Prefix, 0, len(list))
	for _, item := range list {
		if ip, err := netip.ParseAddr(item); err == nil {
			prefixes = append(prefixes, netip.PrefixFrom(ip, ip.BitLen()))
			continue
		}

		prefix, err := netip.ParsePrefix(item)
		if err != nil {
			return nil, fmt.Errorf("--trusted-proxies takes IPs or CIDRs, got %q", item)
		}
		prefixes = append(prefixes, prefix)
	}

	return prefixes, nil
}
