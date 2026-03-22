package tui

import "charm.land/lipgloss/v2"

var (
	ColorForeground = lipgloss.Color("#FAFAFA")
	ColorBackground = lipgloss.Color("#1E1E2E")
	ColorPrimary    = lipgloss.Color("#CBA6F7")
	ColorSecondary  = lipgloss.Color("#94E2D5")
	ColorSuccess    = lipgloss.Color("#A6E3A1")
	ColorError      = lipgloss.Color("#F38BA8")
	ColorWarning    = lipgloss.Color("#FAB387")
	ColorMuted      = lipgloss.Color("#6C7086")
	ColorBorder     = lipgloss.Color("#313244")
	ColorHighlight  = lipgloss.Color("#45475A")

	TitleStyle = lipgloss.NewStyle().
			Foreground(ColorForeground).
			Bold(true).
			Padding(0, 1)

	HeaderStyle = lipgloss.NewStyle().
			Foreground(ColorForeground).
			Bold(true).
			Background(ColorBorder).
			Padding(0, 1).
			Width(100)

	PackageStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true).
			PaddingLeft(0)

	PackageExpandedStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true).
				PaddingLeft(0)

	FileStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			PaddingLeft(2)

	TestStyle = lipgloss.NewStyle().
			Foreground(ColorForeground).
			PaddingLeft(4)

	TestSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorBackground).
				Background(ColorPrimary).
				PaddingLeft(4).
				Width(100)

	StatusPassStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess)

	StatusFailStyle = lipgloss.NewStyle().
			Foreground(ColorError)

	StatusSkipStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)

	StatusRunningStyle = lipgloss.NewStyle().
				Foreground(ColorWarning)

	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(1, 0)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Bold(true)

	FilterInputStyle = lipgloss.NewStyle().
				Foreground(ColorForeground).
				Background(ColorHighlight).
				Padding(0, 1)

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorBorder)

	OutputStyle = lipgloss.NewStyle().
			Foreground(ColorForeground).
			Background(ColorBackground).
			Padding(1, 1)

	HistoryItemStyle = lipgloss.NewStyle().
				Foreground(ColorForeground).
				PaddingLeft(2)

	StatsStyle = lipgloss.NewStyle().
			Foreground(ColorMuted)
)
