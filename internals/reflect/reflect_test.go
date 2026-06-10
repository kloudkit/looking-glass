package reflect

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRender(t *testing.T) {
	cases := []struct {
		name     string
		method   string
		target   string
		body     string
		maxBody  int64
		contains []string
		want     bool
	}{
		{
			name:    "method and path",
			method:  "DELETE",
			target:  "/anything/else",
			maxBody: 1024,
			contains: []string{
				"DELETE /anything/else",
			},
		},
		{
			name:    "query params",
			method:  "GET",
			target:  "/api?page=2&q=hi",
			maxBody: 1024,
			contains: []string{
				"Query:",
				"page = 2",
				"q    = hi",
			},
		},
		{
			name:    "body reflected",
			method:  "POST",
			target:  "/users",
			body:    `{"name":"ada"}`,
			maxBody: 1024,
			contains: []string{
				"Body (14 bytes):",
				`{"name":"ada"}`,
			},
		},
		{
			name:    "body truncated",
			method:  "POST",
			target:  "/big",
			body:    strings.Repeat("x", 50),
			maxBody: 10,
			contains: []string{
				"Body (10 bytes, truncated):",
			},
			want: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.target, strings.NewReader(tc.body))
			req.Header.Set("X-Trace", "abc")

			out, truncated := Render(req, tc.maxBody)

			if truncated != tc.want {
				t.Errorf("truncated = %v, want %v", truncated, tc.want)
			}

			for _, want := range tc.contains {
				if !strings.Contains(out, want) {
					t.Errorf("output missing %q\n--- got ---\n%s", want, out)
				}
			}

			if !strings.Contains(out, "X-Trace: abc") {
				t.Errorf("output missing header X-Trace\n--- got ---\n%s", out)
			}
		})
	}
}

func TestHandler(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("PATCH", "/health?ping=1", nil)

	Handler(1024)(rec, req)

	if rec.Code != 200 {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	if ct := rec.Header().Get("Content-Type"); ct != "text/plain; charset=utf-8" {
		t.Errorf("content-type = %q", ct)
	}

	if !strings.Contains(rec.Body.String(), "PATCH /health?ping=1") {
		t.Errorf("body missing request line\n%s", rec.Body.String())
	}
}
