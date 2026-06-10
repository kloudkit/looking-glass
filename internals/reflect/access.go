package reflect

import (
	"fmt"
	"image/color"
	"strings"

	"charm.land/lipgloss/v2"
)

var (
	mutedStyle  = lipgloss.NewStyle().Foreground(color.RGBA{115, 121, 148, 255})
	methodStyle = lipgloss.NewStyle().Foreground(color.RGBA{140, 170, 238, 255}).Bold(true)
	formatStyle = lipgloss.NewStyle().Foreground(color.RGBA{129, 200, 190, 255})
	pathStyle   = lipgloss.NewStyle().Foreground(color.RGBA{198, 208, 245, 255})
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
