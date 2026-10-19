package reflect

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

var testOpts = opts(1024)

func opts(maxBody int64) Options {
	return Options{MaxBody: maxBody, MaxDelay: time.Second, MaxWSDuration: 5 * time.Second, TrustedProxies: PrivateNetworks}
}

// serve answers req with a handler that never shuts down.
func serve(req *http.Request, o Options) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	Handler(context.Background(), o)(rec, req)

	return rec
}

func reflectOf(method, target, body string) reflection {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("X-Trace", "abc")

	return build(req, testOpts)
}

func TestBuild(t *testing.T) {
	ref := reflectOf("POST", "/api?page=2&q=hi", `{"name":"ada"}`)

	if ref.Method != "POST" || ref.URI != "/api?page=2&q=hi" {
		t.Errorf("method/uri = %q %q", ref.Method, ref.URI)
	}

	if got := ref.Query["page"]; len(got) != 1 || got[0] != "2" {
		t.Errorf("query page = %v", ref.Query["page"])
	}

	if got := ref.Headers["X-Trace"]; len(got) != 1 || got[0] != "abc" {
		t.Errorf("header X-Trace = %v", ref.Headers["X-Trace"])
	}

	if ref.Body != `{"name":"ada"}` || ref.BodyBytes != 14 || ref.Truncated {
		t.Errorf("body = %q (%d, trunc=%v)", ref.Body, ref.BodyBytes, ref.Truncated)
	}
}

func TestJSON(t *testing.T) {
	out := reflectOf("GET", "/x?a=1", "").json()

	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}

	if parsed["method"] != "GET" {
		t.Errorf("method = %v", parsed["method"])
	}
}

func TestHTMLEscapesInput(t *testing.T) {
	out := reflectOf("POST", "/x", "<script>alert(1)</script>").html()

	if strings.Contains(out, "<script>alert(1)</script>") {
		t.Errorf("body was not escaped:\n%s", out)
	}

	if !strings.Contains(out, "&lt;script&gt;") {
		t.Errorf("expected escaped body:\n%s", out)
	}
}

func TestActivityLine(t *testing.T) {
	line := reflectOf("DELETE", "/gone?id=7", "SENSITIVE-BODY").activity("json", 200)

	if !strings.Contains(line, "\x1b[") {
		t.Errorf("expected ANSI color codes:\n%q", line)
	}

	for _, want := range []string{"DELETE", "/gone?id=7", "json", "200"} {
		if !strings.Contains(line, want) {
			t.Errorf("activity line missing %q:\n%s", want, line)
		}
	}

	if strings.Contains(line, "SENSITIVE-BODY") {
		t.Errorf("request body must not appear in the activity log:\n%s", line)
	}

	if strings.Contains(line, "\n") {
		t.Errorf("activity log must be a single line:\n%q", line)
	}
}

func TestActivitySanitizesControlChars(t *testing.T) {
	ref := reflectOf("GET", "/x", "")
	ref.URI = "/x\x1b[2J\x07"

	line := ref.activity("json", 200)

	if strings.Contains(line, "\x1b[2J") || strings.Contains(line, "\x07") {
		t.Errorf("control chars from the request must be stripped:\n%q", line)
	}
}

func TestHandlerDefaultsToHTML(t *testing.T) {
	req := httptest.NewRequest("GET", "/page?ping=1", nil)

	rec := serve(req, testOpts)

	assertHTML(t, rec)
}

func TestHandlerJSONViaHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	req.Header.Set(formatOption.header, "json")

	rec := serve(req, testOpts)

	assertJSON(t, rec)
}

func TestHandlerJSONViaQuery(t *testing.T) {
	req := httptest.NewRequest("GET", "/health?format=json", nil)

	rec := serve(req, testOpts)

	assertJSON(t, rec)
}

func TestFlat(t *testing.T) {
	out := reflectOf("POST", "/api?page=2", `{"name":"ada"}`).flat()

	for _, want := range []string{
		"POST /api?page=2 HTTP/1.1",
		"Host: example.com",
		"X-Trace: abc",
		`{"name":"ada"}`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("flat output missing %q:\n%s", want, out)
		}
	}
}

func TestHTMLHasCopyControl(t *testing.T) {
	out := reflectOf("GET", "/x?a=1", "").html()

	for _, want := range []string{
		`id="copy-json"`,
		`id="copy-flat"`,
		"navigator.clipboard.writeText",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("HTML missing copy control %q:\n%s", want, out)
		}
	}
}

func TestHTMLHasLogo(t *testing.T) {
	out := reflectOf("GET", "/x", "").html()

	for _, want := range []string{
		`<link rel="icon" href="data:image/svg+xml,`,
		`<svg class="logo"`,
		`king glass`,
		`@media (prefers-reduced-motion: reduce)`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("HTML missing logo part %q", want)
		}
	}
	if strings.Contains(out, "ZgotmplZ") {
		t.Error("html/template rejected a logo URL or attribute")
	}
}

func assertHTML(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()

	if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("content-type = %q", ct)
	}

	if !strings.Contains(rec.Body.String(), "<!doctype html>") {
		t.Errorf("expected HTML document:\n%s", rec.Body.String())
	}
}

func assertJSON(t *testing.T, rec *httptest.ResponseRecorder) {
	t.Helper()

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type = %q", ct)
	}

	var parsed map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}

func TestClient(t *testing.T) {
	ref := reflectOf("GET", "/x", "")
	if ref.Client != "192.0.2.1" {
		t.Errorf("client = %q, want the remote address without its port", ref.Client)
	}

	for _, tc := range []struct {
		peer, forwarded, want string
	}{
		{"10.0.0.2:1234", "203.0.113.7", "203.0.113.7"},
		{"10.0.0.2:1234", "6.6.6.6, 203.0.113.7, 10.0.0.9", "203.0.113.7"},
		{"10.0.0.2:1234", "10.0.0.8, 10.0.0.9", "10.0.0.8"},
		{"10.0.0.2:1234", ", 203.0.113.7,", "203.0.113.7"},
		{"[::ffff:10.0.0.2]:1234", "203.0.113.7", "203.0.113.7"},
		{"192.0.2.1:1234", "203.0.113.7", "192.0.2.1"},
	} {
		req := httptest.NewRequest("GET", "/x", nil)
		req.RemoteAddr = tc.peer
		req.Header.Set("X-Forwarded-For", tc.forwarded)

		if got := build(req, testOpts).Client; got != tc.want {
			t.Errorf("peer %s, X-Forwarded-For %q: client = %q, want %q", tc.peer, tc.forwarded, got, tc.want)
		}
	}
}

func TestClientTrustsNoProxyWhenEmpty(t *testing.T) {
	req := httptest.NewRequest("GET", "/x", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.7")

	if got := build(req, Options{MaxBody: 1024}).Client; got != "10.0.0.2" {
		t.Errorf("client = %q, want the peer", got)
	}
}

func TestActivityShowsSanitizedClient(t *testing.T) {
	req := httptest.NewRequest("GET", "/x", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.7\x1b[2J")

	line := build(req, testOpts).activity("html", 200)

	if !strings.Contains(line, "203.0.113.7") || strings.Contains(line, "\x1b[2J") {
		t.Errorf("activity line = %q", line)
	}
}

func TestFlatSortsHeaders(t *testing.T) {
	req := httptest.NewRequest("GET", "/x", nil)
	for _, h := range []string{"X-C", "X-A", "X-B"} {
		req.Header.Set(h, "1")
	}

	out := build(req, testOpts).flat()

	if a, b, c := strings.Index(out, "X-A"), strings.Index(out, "X-B"), strings.Index(out, "X-C"); a >= b || b >= c {
		t.Errorf("headers out of order:\n%s", out)
	}
}

func TestHTMLShowsProxyAndPrettyJSON(t *testing.T) {
	req := httptest.NewRequest("POST", "/x", strings.NewReader(`{"name":"ada"}`))
	req.RemoteAddr = "10.0.0.2:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.7")

	out := build(req, testOpts).html()

	for _, want := range []string{"203.0.113.7 via 10.0.0.2:1234", "{\n  &#34;name&#34;: &#34;ada&#34;\n}"} {
		if !strings.Contains(out, want) {
			t.Errorf("HTML missing %q", want)
		}
	}
}

func TestServerIsHostname(t *testing.T) {
	want, _ := os.Hostname()
	ref := reflectOf("GET", "/x", "")

	if ref.Server != want || !strings.Contains(ref.html(), "served by "+want) {
		t.Errorf("server = %q, want %q", ref.Server, want)
	}
}
