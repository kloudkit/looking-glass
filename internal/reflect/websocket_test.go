package reflect

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

func wsServer(t *testing.T, ctx context.Context, o Options) string {
	t.Helper()

	srv := httptest.NewServer(Handler(ctx, o))
	t.Cleanup(srv.Close)

	return "ws" + strings.TrimPrefix(srv.URL, "http")
}

func wsOpts() Options {
	o := testOpts
	o.MaxWSDuration = 5 * time.Second

	return o
}

func dial(t *testing.T, url string, header http.Header) (*websocket.Conn, reflection) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, url, &websocket.DialOptions{HTTPHeader: header})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.CloseNow() })

	typ, data, err := conn.Read(ctx)
	if err != nil || typ != websocket.MessageText {
		t.Fatalf("first message: %v %v", typ, err)
	}

	var ref reflection
	if err := json.Unmarshal(data, &ref); err != nil {
		t.Fatalf("first message is not the reflection: %v\n%s", err, data)
	}

	return conn, ref
}

func TestWebSocketReflectsThenEchoes(t *testing.T) {
	conn, ref := dial(t, wsServer(t, context.Background(), wsOpts())+"/live?room=1", nil)

	if ref.Method != "GET" || ref.URI != "/live?room=1" || ref.Headers["Upgrade"][0] != "websocket" {
		t.Errorf("reflection = %+v", ref)
	}

	ctx := context.Background()
	for _, msg := range []struct {
		typ  websocket.MessageType
		data string
	}{{websocket.MessageText, "hello"}, {websocket.MessageBinary, "\x00\xff"}} {
		if err := conn.Write(ctx, msg.typ, []byte(msg.data)); err != nil {
			t.Fatal(err)
		}

		typ, data, err := conn.Read(ctx)
		if err != nil || typ != msg.typ || string(data) != msg.data {
			t.Errorf("echo = %v %q %v, want %v %q", typ, data, err, msg.typ, msg.data)
		}
	}
}

func TestWebSocketStripsCrossOriginCookies(t *testing.T) {
	url := wsServer(t, context.Background(), wsOpts())

	header := http.Header{"Origin": {"https://evil.example"}, "Cookie": {"session=secret"}}
	if _, ref := dial(t, url, header); ref.Headers["Cookie"] != nil {
		t.Errorf("a cross-origin page was handed the cookies: %v", ref.Headers["Cookie"])
	}

	o := wsOpts()
	o.Credentials = true
	if _, ref := dial(t, wsServer(t, context.Background(), o), header); ref.Headers["Cookie"] == nil {
		t.Error("with credentials on, the cookies must be reflected")
	}
}

func TestWebSocketKeepsSameOriginCookies(t *testing.T) {
	url := wsServer(t, context.Background(), wsOpts())
	host := strings.TrimPrefix(url, "ws://")

	header := http.Header{"Origin": {"http://" + host}, "Cookie": {"session=mine"}}
	if _, ref := dial(t, url, header); ref.Headers["Cookie"] == nil {
		t.Error("a same-origin page must see its own cookies")
	}
}

func TestWebSocketMessageTooBig(t *testing.T) {
	o := wsOpts()
	o.MaxBody = 16
	conn, _ := dial(t, wsServer(t, context.Background(), o), nil)

	ctx := context.Background()
	_ = conn.Write(ctx, websocket.MessageText, []byte(strings.Repeat("x", 64)))

	if _, _, err := conn.Read(ctx); websocket.CloseStatus(err) != websocket.StatusMessageTooBig {
		t.Errorf("close = %v, want 1009", err)
	}
}

func TestWebSocketLifetime(t *testing.T) {
	o := wsOpts()
	o.MaxWSDuration = 50 * time.Millisecond
	conn, _ := dial(t, wsServer(t, context.Background(), o), nil)

	if _, _, err := conn.Read(context.Background()); websocket.CloseStatus(err) != websocket.StatusGoingAway {
		t.Errorf("close = %v, want 1001", err)
	}
}

func TestWebSocketClosedOnShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	conn, _ := dial(t, wsServer(t, ctx, wsOpts()), nil)

	cancel()

	_, _, err := conn.Read(context.Background())
	if websocket.CloseStatus(err) != websocket.StatusGoingAway || !strings.Contains(err.Error(), "shutting down") {
		t.Errorf("close = %v, want 1001 server shutting down", err)
	}
}

func TestWebSocketStatusRefusesUpgrade(t *testing.T) {
	url := wsServer(t, context.Background(), wsOpts())

	_, resp, err := websocket.Dial(context.Background(), url+"/?glass-status=503", nil)
	if err == nil || resp == nil || resp.StatusCode != 503 {
		t.Errorf("dial = %v, want a refused handshake with 503", err)
	}
}

func TestIsWebSocket(t *testing.T) {
	req := httptest.NewRequest("GET", "/x", nil)
	req.Header.Set("Connection", "keep-alive, Upgrade")
	req.Header.Set("Upgrade", "WebSocket")

	if !isWebSocket(req) {
		t.Error("expected a WebSocket upgrade")
	}

	req.Header.Set("Upgrade", "h2c")
	if isWebSocket(req) {
		t.Error("an h2c upgrade is not a WebSocket")
	}
}
