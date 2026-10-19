package reflect

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHandlerStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set(StatusHeader, "503")

	Handler(1024, time.Second)(rec, req)

	if rec.Code != 503 {
		t.Errorf("status = %d, want 503", rec.Code)
	}

	assertHTML(t, rec)
}

func TestHandlerDelay(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set(DelayHeader, "50ms")

	start := time.Now()
	Handler(1024, time.Second)(rec, req)

	if took := time.Since(start); took < 50*time.Millisecond {
		t.Errorf("answered after %s, want at least 50ms", took)
	}

	if rec.Code != 200 {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestHandlerRejectsBadShape(t *testing.T) {
	for header, value := range map[string]string{
		StatusHeader: "100",
		DelayHeader:  "5s",
	} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/x", nil)
		req.Header.Set(header, value)

		Handler(1024, time.Second)(rec, req)

		if rec.Code != 400 || !strings.Contains(rec.Body.String(), header) {
			t.Errorf("%s: %s: got %d %q", header, value, rec.Code, rec.Body.String())
		}
	}
}

func TestStatusQueryParamIsJustReflected(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/orders?status=open&delay=soon", nil)

	Handler(1024, time.Second)(rec, req)

	if rec.Code != 200 {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestParseDelay(t *testing.T) {
	for in, want := range map[string]time.Duration{
		"300ms": 300 * time.Millisecond,
		"1.5":   1500 * time.Millisecond,
		"2":     2 * time.Second,
	} {
		if got, err := parseDelay(in); err != nil || got != want {
			t.Errorf("%q: got %s (%v), want %s", in, got, err, want)
		}
	}

	if _, err := parseDelay("soon"); err == nil {
		t.Error("expected an error for a non-duration")
	}
}
