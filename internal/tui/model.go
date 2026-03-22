package tui

import (
	"strings"
	"time"

	"github.com/heywinit/alout/internal/history"
	"github.com/heywinit/alout/internal/testrunner"
)

type Model struct {
	Packages        []testrunner.Package
	SelectedPackage int
	SelectedFile    int
	SelectedTest    int

	ExpandedPackages map[int]bool
	ExpandedFiles    map[string]bool

	TestResults   map[string]TestResult
	CurrentOutput string
	StatusMessage string

	FilterQuery string
	FilterMode  bool
	ShowHistory bool
	HistoryRuns []history.TestRun
	HistoryPage int

	IsRunning    bool
	RunAllMode   bool
	RunningTests map[string]bool

	OutputFormat   string
	ShowTestOutput bool

	Width  int
	Height int
}

type TestResult struct {
	Status   string
	Duration time.Duration
	Output   string
}

type FilterMode int

const (
	FilterNone FilterMode = iota
	FilterPassing
	FilterFailing
	FilterAll
)

func NewModel() Model {
	return Model{
		ExpandedPackages: make(map[int]bool),
		ExpandedFiles:    make(map[string]bool),
		TestResults:      make(map[string]TestResult),
		RunningTests:     make(map[string]bool),
		OutputFormat:     "summary",
		ShowTestOutput:   true,
		FilterMode:       false,
		HistoryPage:      0,
	}
}

func (m *Model) GetSelectedTest() (pkg testrunner.Package, tf testrunner.TestFile, testName string) {
	if m.SelectedPackage < 0 || m.SelectedPackage >= len(m.Packages) {
		return
	}

	pkg = m.Packages[m.SelectedPackage]

	if m.SelectedFile < 0 || m.SelectedFile >= len(pkg.TestFiles) {
		return
	}

	tf = pkg.TestFiles[m.SelectedFile]

	if m.SelectedTest < 0 || m.SelectedTest >= len(tf.Tests) {
		return
	}

	testName = tf.Tests[m.SelectedTest]
	return
}

func (m *Model) GetTestKey(pkg, testName string) string {
	return pkg + "::" + testName
}

func (m *Model) SetTestResult(pkg, testName, status string, duration time.Duration, output string) {
	key := m.GetTestKey(pkg, testName)
	m.TestResults[key] = TestResult{
		Status:   status,
		Duration: duration,
		Output:   output,
	}
}

func (m *Model) GetTestResult(pkg, testName string) TestResult {
	key := m.GetTestKey(pkg, testName)
	return m.TestResults[key]
}

func (m *Model) GetVisiblePackages() []testrunner.Package {
	if m.FilterQuery == "" {
		return m.Packages
	}

	query := strings.ToLower(m.FilterQuery)
	var filtered []testrunner.Package

	for _, pkg := range m.Packages {
		if !strings.Contains(strings.ToLower(pkg.ImportPath), query) {
			continue
		}

		var files []testrunner.TestFile
		for _, tf := range pkg.TestFiles {
			var tests []string
			for _, t := range tf.Tests {
				if strings.Contains(strings.ToLower(t), query) {
					tests = append(tests, t)
				}
			}
			if len(tests) > 0 {
				files = append(files, testrunner.TestFile{
					Path:  tf.Path,
					Tests: tests,
				})
			}
		}

		if len(files) > 0 {
			pkg.TestFiles = files
			filtered = append(filtered, pkg)
		}
	}

	return filtered
}

func (m *Model) GetStats() (total, pass, fail, skip, running int) {
	total = len(m.TestResults)
	for _, result := range m.TestResults {
		switch result.Status {
		case "pass":
			pass++
		case "fail":
			fail++
		case "skip":
			skip++
		default:
			if _, ok := m.RunningTests[result.Status]; ok {
				running++
			}
		}
	}
	return
}

func (m *Model) TotalTests() int {
	count := 0
	for _, pkg := range m.Packages {
		for _, tf := range pkg.TestFiles {
			count += len(tf.Tests)
		}
	}
	return count
}

func (m *Model) TogglePackageExpand(idx int) {
	m.ExpandedPackages[idx] = !m.ExpandedPackages[idx]
}

func (m *Model) ToggleFileExpand(pkgIdx int, fileIdx int) {
	pkg := m.Packages[pkgIdx]
	tf := pkg.TestFiles[fileIdx]
	key := pkg.ImportPath + "::" + tf.Path
	m.ExpandedFiles[key] = !m.ExpandedFiles[key]
}

func (m *Model) IsPackageExpanded(idx int) bool {
	return m.ExpandedPackages[idx]
}

func (m *Model) IsFileExpanded(pkgIdx, fileIdx int) bool {
	pkg := m.Packages[pkgIdx]
	if fileIdx >= len(pkg.TestFiles) {
		return false
	}
	tf := pkg.TestFiles[fileIdx]
	key := pkg.ImportPath + "::" + tf.Path
	return m.ExpandedFiles[key]
}

func (m *Model) ClearResults() {
	m.TestResults = make(map[string]TestResult)
	m.RunningTests = make(map[string]bool)
	m.IsRunning = false
	m.RunAllMode = false
	m.CurrentOutput = ""
}
