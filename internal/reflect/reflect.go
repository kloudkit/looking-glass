package reflect

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	FormatHeader = "X-Glass-Format"
	FormatQuery  = "format"
)

type reflection struct {
	Method       string              `json:"method"`
	URI          string              `json:"uri"`
	Proto        string              `json:"proto"`
	Time         string              `json:"time"`
	Remote       string              `json:"remote"`
	Client       string              `json:"client"`
	Host         string              `json:"host"`
	Query        map[string][]string `json:"query"`
	Headers      map[string][]string `json:"headers"`
	Body         string              `json:"body"`
	BodyEncoding string              `json:"bodyEncoding"`
	BodyBytes    int64               `json:"bodyBytes"`
	Truncated    bool                `json:"truncated"`
	JSON         json.RawMessage     `json:"json"`
	Form         map[string][]string `json:"form"`
	Files        map[string][]upload `json:"files"`
}

func build(r *http.Request, maxBody int64) reflection {
	ref := reflection{
		Method:  r.Method,
		URI:     r.RequestURI,
		Proto:   r.Proto,
		Time:    time.Now().UTC().Format(time.RFC3339),
		Remote:  r.RemoteAddr,
		Client:  client(r),
		Host:    r.Host,
		Query:   r.URL.Query(),
		Headers: r.Header,
	}
	ref.readBody(r, maxBody)

	return ref
}

func Handler(maxBody int64, maxDelay time.Duration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ref := build(r, maxBody)

		if cors(w, r) {
			log.Print(ref.activity("preflight", http.StatusNoContent))
			w.WriteHeader(http.StatusNoContent)

			return
		}

		status, delay, err := shape(r, maxDelay)
		if err != nil {
			log.Print(ref.activity("error", http.StatusBadRequest))
			http.Error(w, err.Error(), http.StatusBadRequest)

			return
		}

		format := "html"
		if wantsJSON(r) {
			format = "json"
		}

		log.Print(ref.activity(format, status))

		select {
		case <-r.Context().Done():
			return
		case <-time.After(delay):
		}

		if format == "html" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(status)
			io.WriteString(w, ref.html())

			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		io.WriteString(w, ref.json())
	}
}

func wantsJSON(r *http.Request) bool {
	format := r.Header.Get(FormatHeader)
	if format == "" {
		format = r.URL.Query().Get(FormatQuery)
	}

	return strings.EqualFold(format, "json")
}

// client is the first X-Forwarded-For entry, so a request behind a proxy shows who sent it.
func client(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		first, _, _ := strings.Cut(forwarded, ",")

		return strings.TrimSpace(first)
	}

	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		return host
	}

	return r.RemoteAddr
}
