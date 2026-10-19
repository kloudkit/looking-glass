package reflect

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCORSWithoutOrigin(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/x", nil)

	Handler(1024, time.Second)(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("allow-origin = %q", got)
	}

	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "" {
		t.Errorf("credentials must not be allowed with *, got %q", got)
	}
}

func TestCORSEchoesOrigin(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/x", nil)
	req.Header.Set("Origin", "https://app.example.com")

	Handler(1024, time.Second)(rec, req)

	h := rec.Header()
	if h.Get("Access-Control-Allow-Origin") != "https://app.example.com" ||
		h.Get("Access-Control-Allow-Credentials") != "true" || h.Get("Vary") != "Origin" {
		t.Errorf("headers = %v", h)
	}

	assertHTML(t, rec)
}

func TestCORSPreflight(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("OPTIONS", "/x", nil)
	req.Header.Set("Origin", "https://app.example.com")
	req.Header.Set("Access-Control-Request-Method", "PUT")
	req.Header.Set("Access-Control-Request-Headers", "content-type, x-glass-format")

	Handler(1024, time.Second)(rec, req)

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
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("OPTIONS", "/x", nil)

	Handler(1024, time.Second)(rec, req)

	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "OPTIONS /x") {
		t.Errorf("OPTIONS without a preflight was not reflected: %d", rec.Code)
	}
}
