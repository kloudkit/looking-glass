package reflect

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORSWithoutOrigin(t *testing.T) {
	req := httptest.NewRequest("GET", "/x", nil)

	rec := serve(req, testOpts)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("allow-origin = %q", got)
	}

	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Errorf("credentials must not be allowed with *, got %q", got)
	}
}

func TestCORSEchoesOrigin(t *testing.T) {
	req := httptest.NewRequest("POST", "/x", nil)
	req.Header.Set("Origin", "https://app.example.com")

	rec := serve(req, testOpts)

	h := rec.Header()
	if h.Get("Access-Control-Allow-Origin") != "https://app.example.com" || h.Get("Vary") != "Origin" {
		t.Errorf("headers = %v", h)
	}

	if got := h.Get("Access-Control-Allow-Credentials"); got != "" {
		t.Errorf("credentials must be opt-in, got %q", got)
	}

	assertHTML(t, rec)
}

func TestCORSCredentials(t *testing.T) {
	req := httptest.NewRequest("POST", "/x", nil)
	req.Header.Set("Origin", "https://app.example.com")

	o := testOpts
	o.Credentials = true
	rec := serve(req, o)

	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("allow-credentials = %q", got)
	}
}

func TestCORSPreflight(t *testing.T) {
	req := httptest.NewRequest("OPTIONS", "/x", nil)
	req.Header.Set("Origin", "https://app.example.com")
	req.Header.Set("Access-Control-Request-Method", "PUT")
	req.Header.Set("Access-Control-Request-Headers", "content-type, x-glass-format")

	rec := serve(req, testOpts)

	h := rec.Header()
	if rec.Code != 204 || rec.Body.Len() != 0 {
		t.Errorf("preflight answered %d with %q", rec.Code, rec.Body.String())
	}

	if h.Get("Access-Control-Allow-Methods") != "PUT" ||
		h.Get("Access-Control-Allow-Headers") != "content-type, x-glass-format" {
		t.Errorf("headers = %v", h)
	}
}

func TestPlainOptionsIsReflected(t *testing.T) {
	req := httptest.NewRequest("OPTIONS", "/x", nil)

	rec := serve(req, testOpts)

	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "OPTIONS /x") {
		t.Errorf("OPTIONS without a preflight was not reflected: %d", rec.Code)
	}
}
