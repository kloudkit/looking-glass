package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
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
	defaultPort    = "8080"
	defaultMaxBody = 1 << 20 // 1 MiB
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
		"looking-glass --port 9000 --max-body 65536",
	Args: cobra.NoArgs,
	RunE: serve,
}

func init() {
	f := rootCmd.Flags()
	f.StringP("port", "p", envOr("PORT", defaultPort), "port to listen on (env: PORT)")
	f.Int64("max-body", envInt("MAX_BODY_BYTES", defaultMaxBody), "max body bytes reflected before cut-off (env: MAX_BODY_BYTES)")
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

	mux := http.NewServeMux()
	mux.HandleFunc("/", reflect.Handler(maxBody))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

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
