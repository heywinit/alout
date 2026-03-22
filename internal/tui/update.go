package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/heywinit/alout/internal/history"
	"github.com/heywinit/alout/internal/testrunner"
)

type TestDiscoveredMsg struct {
	Packages []testrunner.Package
	Dir      string
}

type TestResultMsg struct {
	Result testrunner.RunResult
}

type TestCompletedMsg struct {
	Package  string
	TestName string
	Success  bool
}

type HistoryLoadedMsg struct {
	Runs []history.TestRun
}

type StatusMsg struct {
	Message string
}

type FilterSubmittedMsg struct {
	Query string
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		return StatusMsg{Message: "Discovering tests..."}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case TestDiscoveredMsg:
		m.Packages = msg.Packages
		m.StatusMessage = fmt.Sprintf("Found %d packages", len(msg.Packages))
		return m, nil

	case TestResultMsg:
		r := msg.Result
		m.SetTestResult(r.Package, r.TestName, r.Status, r.Duration, r.Output)
		delete(m.RunningTests, m.GetTestKey(r.Package, r.TestName))

		switch r.Status {
		case "pass":
			m.StatusMessage = fmt.Sprintf("✓ %s passed in %s", r.TestName, r.Duration)
		case "fail":
			m.StatusMessage = fmt.Sprintf("✗ %s failed", r.TestName)
			m.CurrentOutput = r.Output
		case "skip":
			m.StatusMessage = fmt.Sprintf("- %s skipped", r.TestName)
		}

		m.CheckAllComplete()
		return m, nil

	case TestCompletedMsg:
		m.StatusMessage = fmt.Sprintf("Test %s: %s", msg.TestName, boolStr(msg.Success))
		m.CheckAllComplete()
		return m, nil

	case HistoryLoadedMsg:
		m.HistoryRuns = msg.Runs
		return m, nil

	case StatusMsg:
		m.StatusMessage = msg.Message
		return m, nil

	case FilterSubmittedMsg:
		m.FilterQuery = msg.Query
		return m, nil

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.FilterMode {
			return m.handleFilterKey(msg)
		}
		return m.handleKey(msg)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) handleFilterKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.Key()
	switch key.Code {
	case tea.KeyEnter:
		m.FilterMode = false
		m.StatusMessage = fmt.Sprintf("Filter: %s", m.FilterQuery)
		return m, nil

	case tea.KeyEsc:
		m.FilterMode = false
		m.FilterQuery = ""
		return m, nil

	case tea.KeyBackspace:
		if len(m.FilterQuery) > 0 {
			m.FilterQuery = m.FilterQuery[:len(m.FilterQuery)-1]
		}
		return m, nil

	default:
		if key.Text != "" {
			m.FilterQuery += key.Text
		}
		return m, nil
	}
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg.String() {
	case "q", "ctrl+c":
		cmds = append(cmds, tea.Quit)

	case "up", "k":
		m.moveUp()

	case "down", "j":
		m.moveDown()

	case "left":
		m.collapse()

	case "right", "l":
		m.expand()

	case "enter":
		if !m.IsRunning {
			cmds = append(cmds, m.runSelectedCmd())
		}

	case "r":
		if !m.IsRunning {
			cmds = append(cmds, m.runAllCmd())
		}

	case "R":
		if !m.IsRunning {
			cmds = append(cmds, m.refreshCmd())
		}

	case "f", "/":
		m.FilterMode = true
		m.FilterQuery = ""
		m.StatusMessage = "Filter: "
		return m, nil

	case "c":
		m.ClearResults()
		m.StatusMessage = "Results cleared"

	case "h":
		m.ShowHistory = !m.ShowHistory
		if m.ShowHistory {
			cmds = append(cmds, m.loadHistoryCmd())
		}

	case "o":
		m.ShowTestOutput = !m.ShowTestOutput
		if !m.ShowTestOutput {
			m.CurrentOutput = ""
		}

	case "1":
		m.OutputFormat = "quiet"
		m.StatusMessage = "Output format: quiet"

	case "2":
		m.OutputFormat = "summary"
		m.StatusMessage = "Output format: summary"

	case "3":
		m.OutputFormat = "verbose"
		m.StatusMessage = "Output format: verbose"
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) moveUp() {
	if m.ShowHistory {
		if m.HistoryPage > 0 {
			m.HistoryPage--
		}
		return
	}

	if m.SelectedTest > 0 {
		m.SelectedTest--
	} else if m.SelectedFile > 0 {
		m.SelectedFile--
		m.SelectedTest = m.getCurrentFileTestCount() - 1
	} else if m.SelectedPackage > 0 {
		m.SelectedPackage--
		m.SelectedFile = len(m.Packages[m.SelectedPackage].TestFiles) - 1
		if m.SelectedFile >= 0 {
			m.SelectedTest = len(m.Packages[m.SelectedPackage].TestFiles[m.SelectedFile].Tests) - 1
		}
	}
}

func (m *Model) moveDown() {
	if m.ShowHistory {
		if m.HistoryPage < len(m.HistoryRuns)-1 {
			m.HistoryPage++
		}
		return
	}

	totalFiles := len(m.GetVisiblePackages())
	if m.SelectedPackage >= totalFiles {
		return
	}

	currentFileTestCount := m.getCurrentFileTestCount()

	if m.SelectedTest < currentFileTestCount-1 {
		m.SelectedTest++
	} else if m.SelectedFile < len(m.Packages[m.SelectedPackage].TestFiles)-1 {
		m.SelectedFile++
		m.SelectedTest = 0
	} else if m.SelectedPackage < totalFiles-1 {
		m.SelectedPackage++
		m.SelectedFile = 0
		m.SelectedTest = 0
	}
}

func (m *Model) expand() {
	if m.ShowHistory {
		return
	}

	packages := m.GetVisiblePackages()
	if m.SelectedPackage >= len(packages) {
		return
	}

	if m.SelectedTest >= 0 {
		m.ToggleFileExpand(m.SelectedPackage, m.SelectedFile)
	} else {
		m.TogglePackageExpand(m.SelectedPackage)
	}
}

func (m *Model) collapse() {
	if m.ShowHistory {
		return
	}

	packages := m.GetVisiblePackages()
	if m.SelectedPackage >= len(packages) {
		return
	}

	if m.SelectedTest >= 0 {
		m.ToggleFileExpand(m.SelectedPackage, m.SelectedFile)
	} else {
		m.TogglePackageExpand(m.SelectedPackage)
	}
}

func (m *Model) getCurrentFileTestCount() int {
	packages := m.GetVisiblePackages()
	if m.SelectedPackage >= len(packages) {
		return 0
	}

	pkg := packages[m.SelectedPackage]
	if m.SelectedFile >= len(pkg.TestFiles) {
		return 0
	}

	return len(pkg.TestFiles[m.SelectedFile].Tests)
}

func (m *Model) CheckAllComplete() {
	if len(m.RunningTests) == 0 && m.IsRunning {
		m.IsRunning = false
		total, pass, fail, _, _ := m.GetStats()
		m.StatusMessage = fmt.Sprintf("Complete: %d pass, %d fail (total: %d)", pass, fail, total)
	}
}

func boolStr(b bool) string {
	if b {
		return "success"
	}
	return "failed"
}

func (m *Model) runSelectedCmd() tea.Cmd {
	pkg, _, testName := m.GetSelectedTest()
	if testName == "" {
		return nil
	}

	m.IsRunning = true
	m.RunningTests[m.GetTestKey(pkg.ImportPath, testName)] = true
	m.StatusMessage = fmt.Sprintf("Running %s...", testName)

	return func() tea.Msg {
		config := testrunner.RunConfig{
			Verbose:      m.OutputFormat == "verbose",
			ShowOutput:   m.ShowTestOutput,
			OutputFormat: m.OutputFormat,
		}

		results, err := testrunner.Run(pkg, testName, "", config)
		if err != nil {
			return StatusMsg{Message: fmt.Sprintf("Error: %v", err)}
		}

		for result := range results {
			return TestResultMsg{Result: result}
		}

		return StatusMsg{Message: "Test completed"}
	}
}

func (m *Model) runAllCmd() tea.Cmd {
	m.ClearResults()
	m.IsRunning = true
	m.RunAllMode = true
	m.StatusMessage = "Running all tests..."

	packages := m.GetVisiblePackages()
	for _, pkg := range packages {
		for _, tf := range pkg.TestFiles {
			for _, testName := range tf.Tests {
				key := m.GetTestKey(pkg.ImportPath, testName)
				m.RunningTests[key] = true
			}
		}
	}

	return func() tea.Msg {
		config := testrunner.RunConfig{
			Verbose:      m.OutputFormat == "verbose",
			ShowOutput:   m.ShowTestOutput,
			OutputFormat: m.OutputFormat,
		}

		results, err := testrunner.RunAll(packages, "", config)
		if err != nil {
			return StatusMsg{Message: fmt.Sprintf("Error: %v", err)}
		}

		for result := range results {
			return TestResultMsg{Result: result}
		}

		return StatusMsg{Message: "All tests completed"}
	}
}

func (m *Model) refreshCmd() tea.Cmd {
	m.StatusMessage = "Refreshing test list..."

	return func() tea.Msg {
		packages, err := testrunner.Discover(".")
		if err != nil {
			return StatusMsg{Message: fmt.Sprintf("Error discovering tests: %v", err)}
		}
		return TestDiscoveredMsg{Packages: packages}
	}
}

func (m *Model) loadHistoryCmd() tea.Cmd {
	return func() tea.Msg {
		db, err := history.New("")
		if err != nil {
			return HistoryLoadedMsg{Runs: []history.TestRun{}}
		}
		defer db.Close()

		runs, err := db.GetTestRuns(100)
		if err != nil {
			return HistoryLoadedMsg{Runs: []history.TestRun{}}
		}

		return HistoryLoadedMsg{Runs: runs}
	}
}

func max(a, b int) int {
	return int(math.Max(float64(a), float64(b)))
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func highlightMatch(s, query string) string {
	if query == "" {
		return s
	}
	return strings.ReplaceAll(s, query, fmt.Sprintf("\033[7m%s\033[0m", query))
}
