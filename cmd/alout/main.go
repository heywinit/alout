package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/heywinit/alout/internal/history"
	"github.com/heywinit/alout/internal/testrunner"
)

type TestItem struct {
	Package  string
	File     string
	TestName string
	Result   string
	Duration string
}

var (
	packages    []testrunner.Package
	allTests    []TestItem
	selected    int
	filterQuery string
	moduleRoot  string
	showHistory bool
)

func main() {
	rootDir := "."
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}

	fmt.Println("\n  alout - Go Test Runner")
	fmt.Println("  =========================\n")

	discoverTests(rootDir)
	if len(allTests) == 0 {
		fmt.Println("  No tests found.")
		os.Exit(0)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\n  Controls: j/k or arrows to navigate | enter to run | f to filter | h for history | q to quit\n")

	for {
		clearScreen()
		render()

		fmt.Print("\n  > ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		handleInput(input)
	}
}

func discoverTests(rootDir string) {
	fmt.Print("  Discovering tests... ")

	moduleRoot = rootDir
	modPath := findGoMod(rootDir)
	if modPath != "" {
		moduleRoot = filepath.Dir(modPath)
	}

	var err error
	packages, err = testrunner.Discover(rootDir)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	allTests = nil
	for _, pkg := range packages {
		for _, tf := range pkg.TestFiles {
			for _, testName := range tf.Tests {
				allTests = append(allTests, TestItem{
					Package:  pkg.ImportPath,
					File:     tf.Path,
					TestName: testName,
				})
			}
		}
	}

	fmt.Printf("Found %d tests in %d packages\n", len(allTests), len(packages))
}

func render() {
	filtered := getFilteredTests()

	title := color("alout", 13)
	if showHistory {
		title = color("alout - History", 13)
		fmt.Printf("  %s\n\n", title)
		renderHistory()
		return
	}

	fmt.Printf("  %s\n\n", title)

	if filterQuery != "" {
		fmt.Printf("  Filter: %s\n", color(filterQuery, 14))
	}

	fmt.Println("  ──────────────────────────────────────────────────────────────")

	if len(filtered) == 0 {
		fmt.Println("  No tests match filter")
		return
	}

	start := selected - 5
	if start < 0 {
		start = 0
	}
	end := start + 20
	if end > len(filtered) {
		end = len(filtered)
	}

	for i := start; i < end; i++ {
		test := filtered[i]
		prefix := "  "
		arrow := "  "

		if i == selected {
			prefix = color(" >", 10)
			arrow = color("▸", 10)
		}

		status := ""
		if test.Result != "" {
			switch test.Result {
			case "pass":
				status = color(" [PASS]", 10)
			case "fail":
				status = color(" [FAIL]", 9)
			case "skip":
				status = color(" [SKIP]", 8)
			}
		}

		shortPkg := shortenPackage(test.Package)
		line := prefix + arrow + " " + color(shortPkg, 12) + color(".", 8) + test.TestName + status
		fmt.Println("  " + line)
	}

	fmt.Println("  ──────────────────────────────────────────────────────────────")
	fmt.Printf("\n  %d/%d tests selected | [j/k] nav | [r] run | [f] filter | [h] history | [q] quit\n", selected+1, len(filtered))
}

func renderHistory() {
	db, err := history.New("")
	if err != nil {
		fmt.Println("  Failed to load history")
		return
	}
	defer db.Close()

	runs, err := db.GetTestRuns(50)
	if err != nil || len(runs) == 0 {
		fmt.Println("  No test history")
		return
	}

	for _, run := range runs {
		status := ""
		switch run.Status {
		case "pass":
			status = color("✓", 10)
		case "fail":
			status = color("✗", 9)
		case "skip":
			status = color("-", 8)
		}

		dur := time.Duration(run.Duration * 1e6).String()
		time := run.Timestamp.Format("15:04:05")
		shortPkg := shortenPackage(run.Package)

		fmt.Printf("  %s %s %s %s (%s)\n", status, run.TestName, color(shortPkg, 12), color(time, 8), dur)
	}
}

func getFilteredTests() []TestItem {
	if filterQuery == "" {
		return allTests
	}

	q := strings.ToLower(filterQuery)
	var filtered []TestItem
	for _, t := range allTests {
		if strings.Contains(strings.ToLower(t.Package), q) ||
			strings.Contains(strings.ToLower(t.TestName), q) {
			filtered = append(filtered, t)
		}
	}
	return filtered
}

func handleInput(input string) {
	cmd := strings.ToLower(input)

	switch cmd {
	case "q", "quit", "exit":
		fmt.Println("\n  Goodbye!")
		os.Exit(0)

	case "j", "down", "l", "right":
		filtered := getFilteredTests()
		if selected < len(filtered)-1 {
			selected++
		}

	case "k", "up", "h", "left":
		if selected > 0 {
			selected--
		}

	case "r", "run", "enter":
		runSelected()

	case "f", "filter":
		filter()

	case "history":
		showHistory = !showHistory

	case "c", "clear":
		clearResults()

	case "1", "run-all":
		runAll()

	case "2", "run-pkg":
		runPackage()
	}
}

func runSelected() {
	filtered := getFilteredTests()
	if selected >= len(filtered) {
		return
	}

	test := filtered[selected]
	fmt.Printf("\n  Running %s...\n", test.TestName)

	config := testrunner.RunConfig{
		Verbose:      true,
		ShowOutput:   true,
		OutputFormat: "verbose",
	}

	pkg := findPackage(test.Package)
	if pkg == nil {
		return
	}

	results, err := testrunner.Run(*pkg, test.TestName, moduleRoot, config)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}

	for result := range results {
		test.Result = result.Status
		test.Duration = result.Duration.String()

		saveToHistory(test.Package, test.TestName, result.Status, result.Duration)

		switch result.Status {
		case "pass":
			fmt.Printf("  %s %s\n", color("✓", 10), test.TestName)
		case "fail":
			fmt.Printf("  %s %s\n", color("✗", 9), test.TestName)
		case "skip":
			fmt.Printf("  %s %s\n", color("-", 8), test.TestName)
		}
	}
}

func runAll() {
	fmt.Println("\n  Running all tests...")
	start := time.Now()

	pass, fail, skip := 0, 0, 0

	config := testrunner.RunConfig{
		Verbose:      false,
		ShowOutput:   false,
		OutputFormat: "summary",
	}

	results, err := testrunner.RunAll(packages, moduleRoot, config)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}

	for result := range results {
		for i, t := range allTests {
			if t.Package == result.Package && t.TestName == result.TestName {
				allTests[i].Result = result.Status
				allTests[i].Duration = result.Duration.String()
				saveToHistory(t.Package, t.TestName, result.Status, result.Duration)

				switch result.Status {
				case "pass":
					pass++
				case "fail":
					fail++
				case "skip":
					skip++
				}
				break
			}
		}
	}

	elapsed := time.Since(start)
	fmt.Printf("\n  Results: %s %d  %s %d  %s %d  (%.2fs)\n",
		color("✓", 10), pass,
		color("✗", 9), fail,
		color("-", 8), skip,
		elapsed.Seconds())
}

func runPackage() {
	filtered := getFilteredTests()
	if selected >= len(filtered) {
		return
	}

	pkgName := filtered[selected].Package
	fmt.Printf("\n  Running all tests in %s...\n", pkgName)

	pkg := findPackage(pkgName)
	if pkg == nil {
		return
	}

	pass, fail, skip := 0, 0, 0

	config := testrunner.RunConfig{
		Verbose:      false,
		ShowOutput:   false,
		OutputFormat: "summary",
	}

	results, err := testrunner.RunPackage(*pkg, moduleRoot, config)
	if err != nil {
		fmt.Printf("  Error: %v\n", err)
		return
	}

	for result := range results {
		for i, t := range allTests {
			if t.Package == result.Package && t.TestName == result.TestName {
				allTests[i].Result = result.Status
				allTests[i].Duration = result.Duration.String()
				saveToHistory(t.Package, t.TestName, result.Status, result.Duration)

				switch result.Status {
				case "pass":
					pass++
				case "fail":
					fail++
				case "skip":
					skip++
				}
				break
			}
		}
	}

	fmt.Printf("\n  Package results: %s %d  %s %d  %s %d\n",
		color("✓", 10), pass,
		color("✗", 9), fail,
		color("-", 8), skip)
}

func filter() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\n  Filter query: ")
	query, _ := reader.ReadString('\n')
	filterQuery = strings.TrimSpace(query)
	selected = 0
}

func clearResults() {
	for i := range allTests {
		allTests[i].Result = ""
		allTests[i].Duration = ""
	}
}

func findPackage(name string) *testrunner.Package {
	for i := range packages {
		if packages[i].ImportPath == name {
			return &packages[i]
		}
	}
	return nil
}

func saveToHistory(pkg, testName, status string, dur time.Duration) {
	db, err := history.New("")
	if err != nil {
		return
	}
	defer db.Close()

	run := &history.TestRun{
		Package:   pkg,
		TestName:  testName,
		Status:    status,
		Duration:  int64(dur.Milliseconds()),
		Timestamp: time.Now(),
	}
	db.SaveTestRun(run)
}

func findGoMod(dir string) string {
	for {
		modPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(modPath); err == nil {
			return modPath
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func shortenPackage(pkg string) string {
	parts := strings.Split(pkg, "/")
	if len(parts) > 2 {
		return strings.Join(parts[len(parts)-2:], "/")
	}
	return pkg
}

func color(s string, code int) string {
	if runtime.GOOS == "windows" {
		return s
	}
	return fmt.Sprintf("\033[38;5;%dm%s\033[0m", code, s)
}

func clearScreen() {
	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else {
		fmt.Print("\033[2J\033[H")
	}
}

func init() {
	// Needed for history timestamps
	_ = time.Now()
}
