package reflect

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

func build_(method, target, body string) reflection {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.Header.Set("X-Trace", "abc")

	return build(req, 1024)
}

func TestBuild(t *testing.T) {
	ref := build_("POST", "/api?page=2&q=hi", `{"name":"ada"}`)

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

func TestBuildTruncates(t *testing.T) {
	req := httptest.NewRequest("POST", "/big", strings.NewReader(strings.Repeat("x", 50)))

	ref := build(req, 10)

	if !ref.Truncated || ref.BodyBytes != 10 {
		t.Errorf("expected truncated at 10, got %d (trunc=%v)", ref.BodyBytes, ref.Truncated)
	}
}

func TestJSON(t *testing.T) {
	out := build_("GET", "/x?a=1", "").json()

	var parsed map[string]any
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}

	if parsed["method"] != "GET" {
		t.Errorf("method = %v", parsed["method"])
	}
}

func TestHTMLEscapesInput(t *testing.T) {
	out := build_("POST", "/x", "<script>alert(1)</script>").html()

	if strings.Contains(out, "<script>alert(1)</script>") {
		t.Errorf("body was not escaped:\n%s", out)
	}

	if !strings.Contains(out, "&lt;script&gt;") {
		t.Errorf("expected escaped body:\n%s", out)
	}
}

func TestTerminalIsColored(t *testing.T) {
	out := build_("DELETE", "/gone", "").terminal()

	if !strings.Contains(out, "\x1b[") {
		t.Errorf("expected ANSI color codes in terminal output:\n%q", out)
	}

	if !strings.Contains(out, "DELETE") {
		t.Errorf("expected method in output:\n%s", out)
	}
}

func TestHandlerDefaultsToJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health?ping=1", nil)

	Handler(1024)(rec, req)

	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type = %q", ct)
	}

	var parsed map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
}

func TestHandlerHTMLFormat(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/page", nil)
	req.Header.Set(FormatHeader, "html")

	Handler(1024)(rec, req)

	if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("content-type = %q", ct)
	}

	if !strings.Contains(rec.Body.String(), "<!doctype html>") {
		t.Errorf("expected HTML document:\n%s", rec.Body.String())
	}
}
