package reflect

import (
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	FormatHeader = "X-Glass-Format"
	FormatQuery  = "format"
)

type reflection struct {
	Method    string              `json:"method"`
	URI       string              `json:"uri"`
	Proto     string              `json:"proto"`
	Time      string              `json:"time"`
	Remote    string              `json:"remote"`
	Host      string              `json:"host"`
	Query     map[string][]string `json:"query"`
	Headers   map[string][]string `json:"headers"`
	Body      string              `json:"body"`
	BodyBytes int                 `json:"bodyBytes"`
	Truncated bool                `json:"truncated"`
}

func build(r *http.Request, maxBody int64) reflection {
	body, n, truncated := readBody(r, maxBody)

	return reflection{
		Method:    r.Method,
		URI:       r.RequestURI,
		Proto:     r.Proto,
		Time:      time.Now().UTC().Format(time.RFC3339),
		Remote:    r.RemoteAddr,
		Host:      r.Host,
		Query:     r.URL.Query(),
		Headers:   r.Header,
		Body:      body,
		BodyBytes: n,
		Truncated: truncated,
	}
}

func Handler(maxBody int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ref := build(r, maxBody)

		format := "json"
		if wantsHTML(r) {
			format = "html"
		}

		log.Print(ref.activity(format))

		if format == "html" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			io.WriteString(w, ref.html())

			return
		}

		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, ref.json())
	}
}

func wantsHTML(r *http.Request) bool {
	format := r.Header.Get(FormatHeader)
	if format == "" {
		format = r.URL.Query().Get(FormatQuery)
	}

	return strings.EqualFold(format, "html")
}

func readBody(r *http.Request, maxBody int64) (string, int, bool) {
	if r.Body == nil {
		return "", 0, false
	}

	data, _ := io.ReadAll(io.LimitReader(r.Body, maxBody+1))
	truncated := int64(len(data)) > maxBody
	if truncated {
		data = data[:maxBody]
	}

	return string(data), len(data), truncated
}
