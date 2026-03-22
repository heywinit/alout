package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
)

func (m Model) View() tea.View {
	var s string

	s += Title.Render("alout")
	if m.IsRunning {
		s += " [running]"
	}
	s += "\n"

	total, pass, fail, skip, _ := m.GetStats()
	s += Stats.Render(fmt.Sprintf(" total: %d | pass: %d | fail: %d | skip: %d\n", total, pass, fail, skip))

	if m.FilterQuery != "" {
		s += fmt.Sprintf(" filter: %s\n", m.FilterQuery)
	}

	s += "\n"

	packages := m.GetVisiblePackages()
	if len(packages) == 0 {
		s += "  No tests found.\n"
		return tea.View{Content: s}
	}

	for pkgIdx, pkg := range packages {
		arrow := ">"
		if !m.IsPackageExpanded(pkgIdx) {
			arrow = "v"
		}

		testCount := 0
		for _, tf := range pkg.TestFiles {
			testCount += len(tf.Tests)
		}

		prefix := " "
		if pkgIdx == m.SelectedPackage {
			prefix = "*"
		}

		s += fmt.Sprintf("%s%s %s (%d)\n", prefix, arrow, pkg.ImportPath, testCount)

		if !m.IsPackageExpanded(pkgIdx) {
			continue
		}

		for fileIdx, tf := range pkg.TestFiles {
			fileArrow := "+"
			if !m.IsFileExpanded(pkgIdx, fileIdx) {
				fileArrow = "-"
			}

			filePrefix := "   "
			if pkgIdx == m.SelectedPackage && m.SelectedFile == fileIdx {
				filePrefix = " * "
			}

			s += fmt.Sprintf("%s%s %s\n", filePrefix, fileArrow, tf.Path)

			if !m.IsFileExpanded(pkgIdx, fileIdx) {
				continue
			}

			for testIdx, testName := range tf.Tests {
				result := m.GetTestResult(pkg.ImportPath, testName)

				testPrefix := "     "
				if pkgIdx == m.SelectedPackage && m.SelectedFile == fileIdx && m.SelectedTest == testIdx {
					testPrefix = ">>> "
				}

				statusStr := ""
				style := TestStyle

				switch result.Status {
				case "pass":
					statusStr = " [PASS]"
					style = TestPass
				case "fail":
					statusStr = " [FAIL]"
					style = TestFail
				case "skip":
					statusStr = " [SKIP]"
					style = TestSkip
				}

				s += style.Render(fmt.Sprintf("%s%s%s\n", testPrefix, testName, statusStr))
			}
		}
	}

	s += "\n"
	s += HelpKey.Render("[r]") + HelpStyle.Render(" run  ")
	s += HelpKey.Render("[f]") + HelpStyle.Render(" filter  ")
	s += HelpKey.Render("[h]") + HelpStyle.Render(" history  ")
	s += HelpKey.Render("[q]") + HelpStyle.Render(" quit\n")

	if m.StatusMessage != "" {
		s += "\n" + m.StatusMessage
	}

	return tea.View{Content: s}
}
