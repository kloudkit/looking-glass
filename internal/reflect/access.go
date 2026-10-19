package reflect

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/kloudkit/looking-glass/internal/ui"
)

// The request side is lavender, what it is reflected as is yellow, as in the logo.
var (
	mutedStyle  = ui.Muted
	methodStyle = ui.Left
	formatStyle = ui.Right
	pathStyle   = ui.Plain
)

func (ref reflection) activity(format string, status int) string {
	size := fmt.Sprintf("%db", ref.BodyBytes)
	if ref.Truncated {
		size += "+"
	}

	return ref.line(format, status, size)
}

// closed is the line for a WebSocket that ended: its close code, how long it stayed open,
// and the messages it received and sent.
func (ref reflection) closed(code int, open time.Duration, in, out int) string {
	return ref.line("ws closed", code, fmt.Sprintf("%s  %d in  %d out", open.Round(time.Millisecond), in, out))
}

// line is who asked for what, then what it was answered with. Everything that came from the
// request is sanitized, so it can't write escape codes to the operator's terminal.
func (ref reflection) line(format string, status int, detail string) string {
	return strings.Join([]string{
		mutedStyle.Render(sanitize(ref.Client)),
		methodStyle.Render(ref.Method),
		pathStyle.Render(sanitize(ref.URI)),
		mutedStyle.Render("→"),
		formatStyle.Render(format),
		formatStyle.Render(strconv.Itoa(status)),
		mutedStyle.Render(detail),
	}, "  ")
}

func sanitize(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}

		return r
	}, s)
}
