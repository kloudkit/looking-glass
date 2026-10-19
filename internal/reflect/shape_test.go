package reflect

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHandlerStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set(statusOption.header, "503")

	Handler(context.Background(), testOpts)(rec, req)

	if rec.Code != 503 {
		t.Errorf("status = %d, want 503", rec.Code)
	}

	assertHTML(t, rec)
}

func TestHandlerDelay(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set(delayOption.header, "50ms")

	start := time.Now()
	Handler(context.Background(), testOpts)(rec, req)

	if took := time.Since(start); took < 50*time.Millisecond {
		t.Errorf("answered after %s, want at least 50ms", took)
	}

	if rec.Code != 200 {
		t.Errorf("status = %d", rec.Code)
	}
}

func TestHandlerRejectsBadShape(t *testing.T) {
	for header, value := range map[string]string{
		statusOption.header: "100",
		delayOption.header:  "5s",
	} {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/x", nil)
		req.Header.Set(header, value)

		Handler(context.Background(), testOpts)(rec, req)

		if rec.Code != 400 || !strings.Contains(rec.Body.String(), header) {
			t.Errorf("%s: %s: got %d %q", header, value, rec.Code, rec.Body.String())
		}
	}
}

func TestHandlerShapeViaQuery(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health?glass-status=503&glass-delay=10ms", nil)

	Handler(context.Background(), testOpts)(rec, req)

	if rec.Code != 503 {
		t.Errorf("status = %d, want 503", rec.Code)
	}
}

func TestHandlerShapeHeaderWins(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health?glass-status=503", nil)
	req.Header.Set(statusOption.header, "500")

	Handler(context.Background(), testOpts)(rec, req)

	if rec.Code != 500 {
		t.Errorf("status = %d, want the header's 500", rec.Code)
	}
}

func TestHandlerDelayEndsOnShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set(delayOption.header, "1s")

	start := time.Now()
	Handler(ctx, testOpts)(rec, req)

	if took := time.Since(start); took > 500*time.Millisecond || rec.Code != 200 {
		t.Errorf("answered %d after %s, want 200 at once", rec.Code, took)
	}
}

func TestHandlerClientGoneDuringDelay(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/x", nil).WithContext(ctx)
	req.Header.Set(delayOption.header, "1s")

	Handler(context.Background(), testOpts)(rec, req)

	if rec.Body.Len() != 0 {
		t.Errorf("answered a client that hung up: %q", rec.Body.String())
	}
}

func TestStatusQueryParamIsJustReflected(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/orders?status=open&delay=soon", nil)

	Handler(context.Background(), testOpts)(rec, req)

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
