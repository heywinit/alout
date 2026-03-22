package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/heywinit/alout/internal/history"
	"github.com/heywinit/alout/internal/testrunner"
)

func main() {
	rootDir := "."
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}

	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.SetOutput(os.Stdout)

	fmt.Println("========================================")
	fmt.Println("  alout debug CLI")
	fmt.Println("========================================")
	fmt.Println()

	fmt.Printf("[DEBUG] Starting at: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("[DEBUG] Root directory: %s\n", rootDir)
	fmt.Println()

	fmt.Println("[STEP 1] Discovering tests...")
	packages, err := testrunner.Discover(rootDir)
	if err != nil {
		log.Fatalf("[ERROR] Failed to discover tests: %v", err)
	}

	fmt.Printf("[OK] Found %d packages with tests\n\n", len(packages))

	for i, pkg := range packages {
		fmt.Printf("  Package %d/%d: %s\n", i+1, len(packages), pkg.ImportPath)
		for _, tf := range pkg.TestFiles {
			relPath := tf.Path
			if strings.HasPrefix(tf.Path, "./") {
				relPath = tf.Path[2:]
			}
			fmt.Printf("    File: %s (%d tests)\n", filepath.Base(relPath), len(tf.Tests))
			for _, test := range tf.Tests {
				fmt.Printf("      - %s\n", test)
			}
		}
	}

	if len(packages) == 0 {
		fmt.Println("[WARN] No packages with tests found")
	}

	fmt.Println()
	fmt.Println("[STEP 2] Running tests...")

	moduleRoot := testrunner.GetModuleRoot(rootDir)
	fmt.Printf("[DEBUG] Module root: %s\n", moduleRoot)

	config := testrunner.RunConfig{
		Verbose:      true,
		ShowOutput:   true,
		OutputFormat: "verbose",
	}

	startTime := time.Now()
	results, err := testrunner.RunAll(packages, moduleRoot, config)
	if err != nil {
		log.Fatalf("[ERROR] Failed to run tests: %v", err)
	}

	var passed, failed, skipped int
	var testResults []testrunner.RunResult

	for result := range results {
		testResults = append(testResults, result)

		status := result.Status
		duration := result.Duration.String()

		switch result.Status {
		case "pass":
			passed++
			if result.TestName != "" {
				fmt.Printf("[PASS] %s %s (%s)\n", result.Package, result.TestName, duration)
			}
		case "fail":
			failed++
			if result.TestName != "" {
				fmt.Printf("[FAIL] %s %s (%s)\n", result.Package, result.TestName, duration)
			}
			if result.Output != "" {
				fmt.Printf("       Output: %s\n", truncate(result.Output, 200))
			}
		case "skip":
			skipped++
			if result.TestName != "" {
				fmt.Printf("[SKIP] %s %s\n", result.Package, result.TestName)
			}
		default:
			if result.TestName != "" {
				fmt.Printf("[INFO] %s %s: %s\n", result.Package, result.TestName, status)
			}
		}
	}

	elapsed := time.Since(startTime)

	fmt.Println()
	fmt.Println("========================================")
	fmt.Printf("  Test Results Summary\n")
	fmt.Println("========================================")
	fmt.Printf("  Total time: %s\n", elapsed.Round(time.Millisecond))
	fmt.Printf("  Passed:  %d\n", passed)
	fmt.Printf("  Failed:  %d\n", failed)
	fmt.Printf("  Skipped: %d\n", skipped)
	fmt.Println("========================================")

	fmt.Println()
	fmt.Println("[STEP 3] Saving to history...")

	db, err := history.New("")
	if err != nil {
		fmt.Printf("[WARN] Failed to save history: %v\n", err)
	} else {
		defer db.Close()

		for _, result := range testResults {
			if result.TestName == "" {
				continue
			}
			run := &history.TestRun{
				Package:   result.Package,
				TestName:  result.TestName,
				Status:    result.Status,
				Duration:  int64(result.Duration.Milliseconds()),
				Timestamp: time.Now(),
				Output:    result.Output,
			}

			if err := db.SaveTestRun(run); err != nil {
				fmt.Printf("[WARN] Failed to save run: %v\n", err)
			}
		}

		fmt.Println("[OK] History saved")
	}

	fmt.Println()
	fmt.Printf("[DEBUG] Finished at: %s\n", time.Now().Format(time.RFC3339))

	if failed > 0 {
		os.Exit(1)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func printJSON(v interface{}) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("JSON error: %v\n", err)
		return
	}
	fmt.Println(string(data))
}
