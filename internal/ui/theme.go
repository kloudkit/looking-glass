package ui

import (
	"image/color"

	"charm.land/fang/v2"
	"charm.land/lipgloss/v2"
)

// Catppuccin Frappé.
var (
	Base      = color.RGBA{48, 52, 70, 255}
	Surface0  = color.RGBA{65, 69, 89, 255}
	Surface2  = color.RGBA{98, 104, 128, 255}
	Overlay0  = color.RGBA{115, 121, 148, 255}
	Overlay1  = color.RGBA{131, 139, 167, 255}
	Text      = color.RGBA{198, 208, 245, 255}
	Subtext0  = color.RGBA{165, 173, 206, 255}
	Rosewater = color.RGBA{242, 213, 207, 255}
	Red       = color.RGBA{231, 130, 132, 255}
	Yellow    = color.RGBA{229, 200, 144, 255}
	Lavender  = color.RGBA{186, 187, 241, 255}
)

// Styles shared by the startup lines and the activity log.
var (
	Muted  = lipgloss.NewStyle().Foreground(Overlay0)
	Plain  = lipgloss.NewStyle().Foreground(Text)
	Left   = lipgloss.NewStyle().Foreground(Lavender).Bold(true)
	Right  = lipgloss.NewStyle().Foreground(Yellow)
	Danger = lipgloss.NewStyle().Foreground(Red)
)

// ColorScheme follows the logo: the request in lavender, its reflection in yellow.
func ColorScheme(lipgloss.LightDarkFunc) fang.ColorScheme {
	return fang.ColorScheme{
		Base:           Base,
		Title:          Lavender,
		Description:    Text,
		Codeblock:      Surface0,
		Program:        Yellow,
		DimmedArgument: Overlay0,
		Comment:        Overlay1,
		Flag:           Lavender,
		FlagDefault:    Subtext0,
		Command:        Yellow,
		QuotedString:   Yellow,
		Argument:       Rosewater,
		Help:           Subtext0,
		Dash:           Surface2,
		ErrorHeader:    [2]color.Color{Red, Base},
		ErrorDetails:   Text,
	}
}
