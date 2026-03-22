package tui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	_ "github.com/heywinit/alout/internal/testrunner"
)

func (m Model) View() tea.View {
	var content string

	if m.ShowHistory {
		content = m.viewHistory()
	} else {
		var b strings.Builder

		b.WriteString(m.viewHeader())
		b.WriteString("\n")
		b.WriteString(m.viewTestList())
		b.WriteString("\n")
		b.WriteString(m.viewFooter())

		content = b.String()
	}

	return tea.View{Content: content}
}

func (m Model) viewHeader() string {
	var b strings.Builder

	title := " alout "
	if m.IsRunning {
		title += "[running]"
	}

	total, pass, fail, skip, running := m.GetStats()
	stats := fmt.Sprintf(" total: %d | pass: %d | fail: %d | skip: %d | running: %d ",
		total, pass, fail, skip, running)

	b.WriteString(fmt.Sprintf("%s%s\n", TitleStyle.Render(title), StatsStyle.Render(stats)))

	if m.FilterQuery != "" {
		b.WriteString(fmt.Sprintf(" filter: %s ", FilterInputStyle.Render(m.FilterQuery)))
	}

	return b.String()
}

func (m Model) viewTestList() string {
	var b strings.Builder

	packages := m.GetVisiblePackages()
	if len(packages) == 0 {
		return "  No tests found. Press 'r' to refresh."
	}

	for pkgIdx, pkg := range packages {
		isExpanded := m.IsPackageExpanded(pkgIdx)
		arrow := ">"
		if !isExpanded {
			arrow = "v"
		}

		testCount := 0
		for _, tf := range pkg.TestFiles {
			testCount += len(tf.Tests)
		}

		pkgLine := fmt.Sprintf(" %s %s (%d tests)", arrow, pkg.ImportPath, testCount)
		if pkgIdx == m.SelectedPackage {
			pkgLine = PackageStyle.Inline(true).Render(pkgLine)
		} else {
			pkgLine = PackageStyle.Render(pkgLine)
		}

		b.WriteString(pkgLine)
		b.WriteString("\n")

		if !isExpanded {
			continue
		}

		for fileIdx, tf := range pkg.TestFiles {
			isFileExpanded := m.IsFileExpanded(pkgIdx, fileIdx)
			fileArrow := "+"
			if !isFileExpanded {
				fileArrow = "-"
			}

			fileLine := fmt.Sprintf("   %s %s (%d tests)", fileArrow, tf.Path, len(tf.Tests))
			if pkgIdx == m.SelectedPackage && m.SelectedFile == fileIdx {
				fileLine = FileStyle.Inline(true).Render(fileLine)
			} else {
				fileLine = FileStyle.Render(fileLine)
			}
			b.WriteString(fileLine)
			b.WriteString("\n")

			if !isFileExpanded {
				continue
			}

			for testIdx, testName := range tf.Tests {
				result := m.GetTestResult(pkg.ImportPath, testName)
				statusStr := ""
				statusStyle := TestStyle

				switch result.Status {
				case "pass":
					statusStr = " ✓"
					statusStyle = StatusPassStyle
				case "fail":
					statusStr = " ✗"
					statusStyle = StatusFailStyle
				case "skip":
					statusStr = " -"
					statusStyle = StatusSkipStyle
				default:
					if _, running := m.RunningTests[m.GetTestKey(pkg.ImportPath, testName)]; running {
						statusStr = " ..."
						statusStyle = StatusRunningStyle
					}
				}

				if result.Duration > 0 {
					statusStr += fmt.Sprintf(" (%.2fs)", result.Duration.Seconds())
				}

				testLine := fmt.Sprintf("     %s%s", testName, statusStr)
				if pkgIdx == m.SelectedPackage && m.SelectedFile == fileIdx && m.SelectedTest == testIdx {
					testLine = TestSelectedStyle.Render(testLine)
				} else {
					testLine = statusStyle.Inline(true).Render(testLine)
				}
				b.WriteString(testLine)
				b.WriteString("\n")
			}
		}
	}

	return b.String()
}

func (m Model) viewFooter() string {
	var b strings.Builder

	b.WriteString(HelpStyle.Render("\n"))
	b.WriteString(HelpStyle.Render(" "))
	b.WriteString(HelpKeyStyle.Render("[↑/↓]"))
	b.WriteString(HelpStyle.Render(" navigate "))
	b.WriteString(HelpKeyStyle.Render("[←/→]"))
	b.WriteString(HelpStyle.Render(" expand "))
	b.WriteString(HelpKeyStyle.Render("[enter]"))
	b.WriteString(HelpStyle.Render(" run "))
	b.WriteString(HelpKeyStyle.Render("[r]"))
	b.WriteString(HelpStyle.Render(" run all "))
	b.WriteString(HelpKeyStyle.Render("[f]"))
	b.WriteString(HelpStyle.Render(" filter "))
	b.WriteString(HelpKeyStyle.Render("[h]"))
	b.WriteString(HelpStyle.Render(" history "))
	b.WriteString(HelpKeyStyle.Render("[q]"))
	b.WriteString(HelpStyle.Render(" quit"))

	if m.StatusMessage != "" {
		b.WriteString("\n")
		b.WriteString(HelpStyle.Render(" " + m.StatusMessage))
	}

	return b.String()
}

func (m Model) viewHistory() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("%s\n", TitleStyle.Render(" alout - test history ")))
	b.WriteString(HelpStyle.Render("\n"))
	b.WriteString(HelpStyle.Render(" "))
	b.WriteString(HelpKeyStyle.Render("[↑/↓]"))
	b.WriteString(HelpStyle.Render(" navigate "))
	b.WriteString(HelpKeyStyle.Render("[b]"))
	b.WriteString(HelpStyle.Render(" back "))
	b.WriteString(HelpKeyStyle.Render("[q]"))
	b.WriteString(HelpStyle.Render(" quit"))

	b.WriteString("\n\n")

	if len(m.HistoryRuns) == 0 {
		b.WriteString("  No test history yet.\n")
		return b.String()
	}

	for i, run := range m.HistoryRuns {
		statusStr := ""
		statusStyle := StatusPassStyle
		switch run.Status {
		case "pass":
			statusStr = "✓"
		case "fail":
			statusStr = "✗"
			statusStyle = StatusFailStyle
		case "skip":
			statusStr = "-"
			statusStyle = StatusSkipStyle
		}

		timeStr := run.Timestamp.Format("2006-01-02 15:04:05")
		durationStr := time.Duration(run.Duration * 1e6).String()

		line := fmt.Sprintf(" %s  %s  %s  %s (%s)",
			statusStyle.Render(statusStr),
			run.TestName,
			run.Package,
			timeStr,
			durationStr,
		)

		if i == m.HistoryPage {
			line = HistoryItemStyle.Inline(true).Background(ColorHighlight).Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	return b.String()
}
