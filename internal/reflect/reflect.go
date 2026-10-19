package reflect

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strings"
	"time"
)

const (
	FormatHeader = "X-Glass-Format"
	FormatQuery  = "format"

	// nginx's status for a client that hung up before the answer, as go-httpbin logs it.
	statusClientClosed = 499
)

// PrivateNetworks are trusted as proxies by default: a reverse proxy in front of Looking
// Glass sits on one, and a caller from the internet can't.
var PrivateNetworks = []netip.Prefix{
	netip.MustParsePrefix("10.0.0.0/8"),
	netip.MustParsePrefix("172.16.0.0/12"),
	netip.MustParsePrefix("192.168.0.0/16"),
	netip.MustParsePrefix("127.0.0.0/8"),
	netip.MustParsePrefix("fc00::/7"),
	netip.MustParsePrefix("::1/128"),
}

// hostname names the replica that answered, for a Looking Glass behind a load balancer.
var hostname, _ = os.Hostname()

type reflection struct {
	Method          string              `json:"method"`
	URI             string              `json:"uri"`
	Proto           string              `json:"proto"`
	Time            string              `json:"time"`
	Remote          string              `json:"remote"`
	Client          string              `json:"client"`
	Host            string              `json:"host"`
	Server          string              `json:"server"`
	Query           map[string][]string `json:"query"`
	Headers         map[string][]string `json:"headers"`
	Body            string              `json:"body"`
	BodyEncoding    string              `json:"bodyEncoding"`
	ContentEncoding string              `json:"contentEncoding"`
	BodyBytes       int64               `json:"bodyBytes"`
	Truncated       bool                `json:"truncated"`
	JSON            json.RawMessage     `json:"json"`
	Form            map[string][]string `json:"form"`
	Files           map[string][]upload `json:"files"`
}

// Options configure the handler.
type Options struct {
	MaxBody        int64          // body bytes reflected before the cut-off
	MaxDelay       time.Duration  // longest delay a request may ask for
	MaxWSDuration  time.Duration  // longest a WebSocket stays open
	Credentials    bool           // let cross-origin pages send cookies and read the reflection
	TrustedProxies []netip.Prefix // peers whose X-Forwarded-For is believed
}

func build(r *http.Request, opts Options) reflection {
	ref := reflection{
		Method:  r.Method,
		URI:     r.RequestURI,
		Proto:   r.Proto,
		Time:    time.Now().UTC().Format(time.RFC3339),
		Remote:  r.RemoteAddr,
		Client:  client(r, opts.TrustedProxies),
		Host:    r.Host,
		Server:  hostname,
		Query:   r.URL.Query(),
		Headers: r.Header,
	}
	ref.readBody(r, opts.MaxBody)

	return ref
}

// Handler reflects every request. Once ctx is done, requests waiting on a delay are answered at once.
func Handler(ctx context.Context, opts Options) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ref := build(r, opts)

		if cors(w, r, opts.Credentials) {
			log.Print(ref.activity("preflight", http.StatusNoContent))
			w.WriteHeader(http.StatusNoContent)

			return
		}

		answer, err := shape(r, opts.MaxDelay)
		if err != nil {
			log.Print(ref.activity("error", http.StatusBadRequest))
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		format := "html"
		if wantsJSON(r) {
			format = "json"
		}

		timer := time.NewTimer(answer.delay)
		defer timer.Stop()

		select {
		case <-r.Context().Done():
			log.Print(ref.activity(format, statusClientClosed))

			return
		case <-ctx.Done():
		case <-timer.C:
		}

		// Any other status refuses the upgrade, so a client can test a failed connect.
		if isWebSocket(r) && answer.status == http.StatusOK {
			serveWebSocket(ctx, w, r, ref, opts)

			return
		}

		log.Print(ref.activity(format, answer.status))

		body, contentType := ref.html(), "text/html; charset=utf-8"
		if format == "json" {
			body, contentType = ref.json(), "application/json"
		}

		w.Header().Set("Content-Type", contentType)
		write(w, answer.status, answer.encoding, body)
	}
}

func wantsJSON(r *http.Request) bool {
	return strings.EqualFold(option(r, FormatHeader, FormatQuery), "json")
}

// option reads a setting from its header, or else from its query param.
func option(r *http.Request, header, query string) string {
	if v := r.Header.Get(header); v != "" {
		return v
	}

	return r.URL.Query().Get(query)
}

// client walks X-Forwarded-For from the right, past trusted proxies, to the first address
// one of them vouches for. An untrusted peer is the client itself, whatever it forwards.
func client(r *http.Request, trusted []netip.Prefix) string {
	peer := r.RemoteAddr
	if host, _, err := net.SplitHostPort(peer); err == nil {
		peer = host
	}

	if !isTrusted(peer, trusted) {
		return peer
	}

	var hops []string
	for _, line := range r.Header.Values("X-Forwarded-For") {
		for hop := range strings.SplitSeq(line, ",") {
			if hop = strings.TrimSpace(hop); hop != "" {
				hops = append(hops, hop)
			}
		}
	}

	for i := len(hops) - 1; i >= 0; i-- {
		if !isTrusted(hops[i], trusted) || i == 0 {
			return hops[i]
		}
	}

	return peer
}

func isTrusted(addr string, trusted []netip.Prefix) bool {
	ip, err := netip.ParseAddr(addr)
	if err != nil {
		return false
	}

	ip = ip.Unmap()
	for _, prefix := range trusted {
		if prefix.Contains(ip) {
			return true
		}
	}

	return false
}
