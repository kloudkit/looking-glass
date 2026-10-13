package reflect

import (
	"fmt"
	"strings"

	"github.com/kloudkit/looking-glass/internal/ui"
)

// The request side is lavender, what it is reflected as is yellow, as in the logo.
var (
	mutedStyle  = ui.Muted
	methodStyle = ui.Left
	formatStyle = ui.Right
	pathStyle   = ui.Plain
)

func (ref reflection) activity(format string) string {
	size := fmt.Sprintf("%db", ref.BodyBytes)
	if ref.Truncated {
		size += "+"
	}

	return strings.Join([]string{
		mutedStyle.Render(ref.Remote),
		methodStyle.Render(ref.Method),
		pathStyle.Render(sanitize(ref.URI)),
		mutedStyle.Render("→"),
		formatStyle.Render(format),
		mutedStyle.Render(size),
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
