package reflect

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/coder/websocket"
)

var errLifetime = errors.New("max websocket duration reached")

func isWebSocket(r *http.Request) bool {
	return r.Method == http.MethodGet &&
		headerHas(r, "Connection", "upgrade") && headerHas(r, "Upgrade", "websocket")
}

func headerHas(r *http.Request, key, token string) bool {
	for _, v := range r.Header.Values(key) {
		for part := range strings.SplitSeq(v, ",") {
			if strings.EqualFold(strings.TrimSpace(part), token) {
				return true
			}
		}
	}

	return false
}

// crossOrigin is a page on another site. Browsers send it cookies on a WebSocket handshake,
// and CORS never applies, so the reflection must not hand them over.
func crossOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return false
	}

	u, err := url.Parse(origin)

	return err != nil || !strings.EqualFold(u.Host, r.Host)
}

// serveWebSocket sends the handshake's reflection as the first message, then echoes every
// message back as it came, until either side closes, the lifetime runs out or ctx is done.
func serveWebSocket(ctx context.Context, w http.ResponseWriter, r *http.Request, ref reflection, opts Options) {
	if crossOrigin(r) && !opts.Credentials {
		ref.Headers = r.Header.Clone()
		delete(ref.Headers, "Cookie")
	}

	// The origin policy is ours: every origin is accepted, without its cookies.
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		log.Print(ref.activity("error", http.StatusBadRequest))

		return
	}
	conn.SetReadLimit(opts.MaxBody)

	log.Print(ref.activity("ws", http.StatusSwitchingProtocols))

	ctx, cancel := context.WithTimeoutCause(ctx, opts.MaxWSDuration, errLifetime)
	defer cancel()

	// Closing from here sends a close frame; a cancelled read would just drop the connection.
	stop := context.AfterFunc(ctx, func() {
		reason := "server shutting down"
		if errors.Is(context.Cause(ctx), errLifetime) {
			reason = errLifetime.Error()
		}
		_ = conn.Close(websocket.StatusGoingAway, reason)
	})
	defer stop()

	start, in, out := time.Now(), 0, 0
	err = conn.Write(ctx, websocket.MessageText, []byte(ref.json()))
	for err == nil {
		out++

		var (
			typ  websocket.MessageType
			data []byte
		)
		if typ, data, err = conn.Read(context.Background()); err == nil {
			in++
			err = conn.Write(ctx, typ, data)
		}
	}

	code := websocket.CloseStatus(err)
	if code == -1 {
		code = websocket.StatusAbnormalClosure
	}
	_ = conn.CloseNow()

	log.Print(ref.closed(int(code), time.Since(start), in, out))
}
