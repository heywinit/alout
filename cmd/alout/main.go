package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/heywinit/alout/internal/history"
	"github.com/heywinit/alout/internal/testrunner"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]

	switch cmd {
	case "run":
		run()
	case "list":
		list()
	case "history":
		showHistory()
	case "help", "--help", "-h":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`alout - Go test runner

Usage:
  alout run [package] [flags]    Run tests
  alout list [query]            List tests
  alout history [limit]         Show test history

Flags:
  -v, --verbose    Show test output
  -f, --filter     Filter by name
  --format         Output format (summary, verbose, json)

Examples:
  alout run                      # Run all tests
  alout run ./internal/foo        # Run specific package
  alout run -f TestAdd           # Run tests matching filter
  alout list                     # List all tests
  alout list math                # List tests matching "math"
  alout history                  # Show recent test history`)
}

func run() {
	args := os.Args[2:]

	var filter string
	var verbose bool
	var format string
	var target string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-v", "--verbose":
			verbose = true
		case "-f", "--filter":
			if i+1 < len(args) {
				filter = args[i+1]
				i++
			}
		case "--format":
			if i+1 < len(args) {
				format = args[i+1]
				i++
			}
		default:
			if !strings.HasPrefix(arg, "-") {
				target = arg
			}
		}
	}

	if format == "" {
		format = "summary"
	}

	dir := "."
	if target != "" {
		dir = target
	}

	packages, err := testrunner.Discover(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(packages) == 0 {
		fmt.Println("No tests found")
		os.Exit(0)
	}

	root := getModuleRoot(dir)

	if filter != "" {
		packages = filterPackages(packages, filter)
	}

	fmt.Printf("Running %d packages...\n\n", len(packages))

	start := time.Now()
	pass, fail, skip := 0, 0, 0

	config := testrunner.RunConfig{
		Verbose:      verbose,
		ShowOutput:   verbose,
		OutputFormat: format,
	}

	if target != "" && filter == "" {
		for _, pkg := range packages {
			pkgPass, pkgFail, pkgSkip := runPackage(pkg, root, config)
			pass += pkgPass
			fail += pkgFail
			skip += pkgSkip
		}
	} else {
		results, err := testrunner.RunAll(packages, root, config)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		for result := range results {
			saveHistory(result)

			switch result.Status {
			case "pass":
				pass++
				if verbose {
					fmt.Printf("  %s %s\n", green("✓"), result.TestName)
				}
			case "fail":
				fail++
				fmt.Printf("  %s %s\n", red("✗"), result.TestName)
				if verbose && result.Output != "" {
					fmt.Println(result.Output)
				}
			case "skip":
				skip++
				if verbose {
					fmt.Printf("  %s %s\n", gray("-"), result.TestName)
				}
			}
		}
	}

	elapsed := time.Since(start)

	fmt.Printf("\n%s  %d passed  %s  %d failed  %s  %d skipped  (%.2fs)\n",
		green("✓"), pass,
		red("✗"), fail,
		gray("-"), skip,
		elapsed.Seconds())

	if fail > 0 {
		os.Exit(1)
	}
}

func runPackage(pkg testrunner.Package, dir string, config testrunner.RunConfig) (pass, fail, skip int) {
	results, err := testrunner.RunPackage(pkg, dir, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running %s: %v\n", pkg.ImportPath, err)
		return
	}

	for result := range results {
		if result.Status == "output" {
			continue
		}
		saveHistory(result)

		switch result.Status {
		case "pass":
			pass++
			if config.Verbose {
				fmt.Printf("  %s %s\n", green("✓"), result.TestName)
			}
		case "fail":
			fail++
			fmt.Printf("  %s %s\n", red("✗"), result.TestName)
		case "skip":
			skip++
			fmt.Printf("  %s %s\n", gray("-"), result.TestName)
		}
	}

	return
}

func list() {
	args := os.Args[2:]

	dir := "."
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		dir = args[0]
		args = args[1:]
	}

	query := ""
	if len(args) > 0 {
		query = args[0]
	}

	packages, err := testrunner.Discover(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if query != "" {
		packages = filterPackages(packages, query)
	}

	if len(packages) == 0 {
		fmt.Println("No tests found")
		return
	}

	if query != "" {
		packages = filterPackages(packages, query)
	}

	total := 0
	for _, pkg := range packages {
		for _, tf := range pkg.TestFiles {
			for _, test := range tf.Tests {
				fmt.Printf("  %s.%s\n", green(pkg.ImportPath), test)
				total++
			}
		}
	}

	fmt.Printf("\n%d tests in %d packages\n", total, len(packages))
}

func showHistory() {
	args := os.Args[2:]

	limit := 50
	if len(args) > 0 {
		fmt.Sscanf(args[0], "%d", &limit)
	}

	db, err := history.New("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	runs, err := db.GetTestRuns(limit)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(runs) == 0 {
		fmt.Println("No test history")
		return
	}

	for _, run := range runs {
		status := ""
		switch run.Status {
		case "pass":
			status = green("✓")
		case "fail":
			status = red("✗")
		case "skip":
			status = gray("-")
		}

		dur := time.Duration(run.Duration * 1e6).String()
		time := run.Timestamp.Format("2006-01-02 15:04:05")

		fmt.Printf("  %s %-30s %s %s\n", status, run.TestName, time, dur)
	}
}

func filterPackages(packages []testrunner.Package, query string) []testrunner.Package {
	q := strings.ToLower(query)
	var filtered []testrunner.Package

	for _, pkg := range packages {
		if !strings.Contains(strings.ToLower(pkg.ImportPath), q) {
			continue
		}

		var files []testrunner.TestFile
		for _, tf := range pkg.TestFiles {
			var tests []string
			for _, t := range tf.Tests {
				if strings.Contains(strings.ToLower(t), q) {
					tests = append(tests, t)
				}
			}
			if len(tests) > 0 {
				files = append(files, testrunner.TestFile{Path: tf.Path, Tests: tests})
			}
		}

		if len(files) > 0 {
			filtered = append(filtered, testrunner.Package{
				ImportPath: pkg.ImportPath,
				TestFiles:  files,
			})
		}
	}

	return filtered
}

func saveHistory(result testrunner.RunResult) {
	if result.TestName == "" {
		return
	}

	db, err := history.New("")
	if err != nil {
		return
	}
	defer db.Close()

	db.SaveTestRun(&history.TestRun{
		Package:   result.Package,
		TestName:  result.TestName,
		Status:    result.Status,
		Duration:  int64(result.Duration.Milliseconds()),
		Timestamp: time.Now(),
		Output:    result.Output,
	})
}

func getModuleRoot(dir string) string {
	for {
		modPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(modPath); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return dir
}

func green(s string) string { return color(s, 2) }
func red(s string) string   { return color(s, 1) }
func gray(s string) string  { return color(s, 8) }
func color(s string, c int) string {
	return fmt.Sprintf("\033[38;5;%dm%s\033[0m", c*10+1, s)
}
