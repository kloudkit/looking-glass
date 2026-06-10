package reflect

import (
	"fmt"
	"image/color"

	"charm.land/lipgloss/v2"
	"charm.land/lipgloss/v2/table"
)

var (
	mauve   = color.RGBA{202, 158, 230, 255}
	blue    = color.RGBA{140, 170, 238, 255}
	teal    = color.RGBA{129, 200, 190, 255}
	text    = color.RGBA{198, 208, 245, 255}
	overlay = color.RGBA{115, 121, 148, 255}
	surface = color.RGBA{98, 104, 128, 255}
	yellow  = color.RGBA{229, 200, 144, 255}

	titleStyle   = lipgloss.NewStyle().Foreground(mauve).Bold(true)
	metaStyle    = lipgloss.NewStyle().Foreground(overlay)
	sectionStyle = lipgloss.NewStyle().Foreground(blue).Bold(true)
	keyStyle     = lipgloss.NewStyle().Foreground(teal).Bold(true).Padding(0, 1)
	valueStyle   = lipgloss.NewStyle().Foreground(text).Padding(0, 1).Width(60)
	borderStyle  = lipgloss.NewStyle().Foreground(surface)
	bodyStyle    = lipgloss.NewStyle().
			Foreground(text).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(surface).
			Padding(0, 1)
)

func (ref reflection) terminal() string {
	blocks := []string{
		titleStyle.Render(fmt.Sprintf("%s %s  %s", ref.Method, ref.URI, ref.Proto)),
		metaStyle.Render(fmt.Sprintf("%s  ·  %s  ·  %s", ref.Time, ref.Remote, ref.Host)),
	}

	if len(ref.Query) > 0 {
		blocks = append(blocks, section("Query", ref.Query))
	}

	if len(ref.Headers) > 0 {
		blocks = append(blocks, section("Headers", ref.Headers))
	}

	if ref.BodyBytes > 0 {
		blocks = append(blocks, bodyBlock(ref))
	}

	return lipgloss.JoinVertical(lipgloss.Left, blocks...)
}

func section(title string, values map[string][]string) string {
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(borderStyle).
		StyleFunc(func(_, col int) lipgloss.Style {
			if col == 0 {
				return keyStyle
			}

			return valueStyle
		})

	for _, k := range sortedKeys(values) {
		for _, v := range values[k] {
			t.Row(k, v)
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, "", sectionStyle.Render(title), t.Render())
}

func bodyBlock(ref reflection) string {
	label := fmt.Sprintf("Body (%d bytes)", ref.BodyBytes)
	style := sectionStyle

	if ref.Truncated {
		label = fmt.Sprintf("Body (%d bytes, truncated)", ref.BodyBytes)
		style = lipgloss.NewStyle().Foreground(yellow).Bold(true)
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		"",
		style.Render(label),
		bodyStyle.Render(ref.Body),
	)
}
