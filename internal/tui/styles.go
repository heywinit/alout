package tui

import "github.com/charmbracelet/lipgloss"

var (
	ColorFg     = lipgloss.Color("#FAFAFA")
	ColorBg     = lipgloss.Color("#1E1E2E")
	ColorPurple = lipgloss.Color("#CBA6F7")
	ColorGreen  = lipgloss.Color("#A6E3A1")
	ColorRed    = lipgloss.Color("#F38BA8")
	ColorGray   = lipgloss.Color("#6C7086")
	ColorBorder = lipgloss.Color("#313244")

	Title = lipgloss.NewStyle().
		Foreground(ColorPurple).
		Bold(true)

	Stats = lipgloss.NewStyle().
		Foreground(ColorGray)

	PackageStyle = lipgloss.NewStyle().
			Foreground(ColorPurple)

	FileStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#94E2D5"))

	TestStyle = lipgloss.NewStyle().
			Foreground(ColorFg)

	TestPass = lipgloss.NewStyle().
			Foreground(ColorGreen)

	TestFail = lipgloss.NewStyle().
			Foreground(ColorRed)

	TestSkip = lipgloss.NewStyle().
			Foreground(ColorGray)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorGray)

	HelpKey = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#94E2D5")).
		Bold(true)

	StatusBar = lipgloss.NewStyle().
			Background(ColorBorder).
			Width(100).
			Foreground(ColorFg)
)
