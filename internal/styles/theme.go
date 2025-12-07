package styles

import "github.com/charmbracelet/lipgloss"

var (
	ColRed      = lipgloss.Color("#D80000")
	ColOrange   = lipgloss.Color("#FF4500")
	ColBlue     = lipgloss.Color("#00FFFF")
	ColBlack    = lipgloss.Color("#080808")
	ColGray     = lipgloss.Color("#333333")
	ColGreen    = lipgloss.Color("#00FF00")
	ColWhite    = lipgloss.Color("#FFFFFF")
	ColNeonPink = lipgloss.Color("#FF00FF")

	TxtRed    = lipgloss.NewStyle().Foreground(ColRed)
	TxtOrange = lipgloss.NewStyle().Foreground(ColOrange)
	TxtBlue   = lipgloss.NewStyle().Foreground(ColBlue)
	TxtGray   = lipgloss.NewStyle().Foreground(ColGray)
	TxtGreen  = lipgloss.NewStyle().Foreground(ColGreen)
	TxtWhite  = lipgloss.NewStyle().Foreground(ColWhite)
	TxtPink   = lipgloss.NewStyle().Foreground(ColNeonPink)

	Bars = []string{" ", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
)
