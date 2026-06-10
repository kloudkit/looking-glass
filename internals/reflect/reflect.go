package reflect

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

const divider = "──────────────────────────────────────────────"

func Render(r *http.Request, maxBody int64) (string, bool) {
	var b strings.Builder

	fmt.Fprintf(&b, "%s\n", divider)
	fmt.Fprintf(&b, "%s %s  %s\n", r.Method, r.RequestURI, r.Proto)
	fmt.Fprintf(&b, "%s\n", divider)
	fmt.Fprintf(&b, "Time:    %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Fprintf(&b, "Remote:  %s\n", r.RemoteAddr)
	fmt.Fprintf(&b, "Host:    %s\n", r.Host)

	writeSection(&b, "Query", r.URL.Query(), "", " = ")
	writeSection(&b, "Headers", r.Header, ":", " ")
	truncated := writeBody(&b, r, maxBody)

	return b.String(), truncated
}

func Handler(maxBody int64) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		out, _ := Render(r, maxBody)

		log.Print("\n" + out)

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		io.WriteString(w, out)
	}
}

func writeSection(b *strings.Builder, title string, values map[string][]string, suffix, joiner string) {
	if len(values) == 0 {
		return
	}

	keys := make([]string, 0, len(values))
	width := 0
	for k := range values {
		keys = append(keys, k)
		if l := len(k) + len(suffix); l > width {
			width = l
		}
	}
	sort.Strings(keys)

	fmt.Fprintf(b, "\n%s:\n", title)
	for _, k := range keys {
		for _, v := range values[k] {
			fmt.Fprintf(b, "  %-*s%s%s\n", width, k+suffix, joiner, v)
		}
	}
}

func writeBody(b *strings.Builder, r *http.Request, maxBody int64) bool {
	if r.Body == nil {
		return false
	}

	body, _ := io.ReadAll(io.LimitReader(r.Body, maxBody+1))
	truncated := int64(len(body)) > maxBody
	if truncated {
		body = body[:maxBody]
	}

	if len(body) == 0 {
		return false
	}

	note := ""
	if truncated {
		note = ", truncated"
	}

	fmt.Fprintf(b, "\nBody (%d bytes%s):\n", len(body), note)
	for line := range strings.SplitSeq(strings.TrimRight(string(body), "\n"), "\n") {
		fmt.Fprintf(b, "  %s\n", line)
	}

	return truncated
}
