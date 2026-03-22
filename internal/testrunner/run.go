package testrunner

import (
	"bufio"
	"encoding/json"
	"io"
	"os/exec"
	"strings"
	"time"
)

type RunResult struct {
	Package  string
	TestName string
	Status   string
	Duration time.Duration
	Output   string
	ErrorMsg string
}

type TestEvent struct {
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Output  string  `json:"Output"`
	Error   string  `json:"Error"`
	Elapsed float64 `json:"Elapsed"`
}

type RunConfig struct {
	Verbose       bool
	VerboseFailed bool
	ShowOutput    bool
	OutputFormat  string
}

func Run(pkg Package, testName string, dir string, config RunConfig) (<-chan RunResult, error) {
	args := []string{"test", "-json", "-run", testName}

	cmd := exec.Command("go", args...)
	cmd.Dir = dir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	results := make(chan RunResult)
	go parseOutput(stdout, results, config)

	go func() {
		cmd.Wait()
		close(results)
	}()

	return results, nil
}

func RunPackage(pkg Package, dir string, config RunConfig) (<-chan RunResult, error) {
	args := []string{"test", "-json"}
	if config.Verbose {
		args = []string{"test", "-v", "-json"}
	}
	args = append(args, pkg.ImportPath)

	cmd := exec.Command("go", args...)
	cmd.Dir = dir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	results := make(chan RunResult)

	go func() {
		parseOutput(stdout, results, config)
		cmd.Wait()
		close(results)
	}()

	return results, nil
}

func RunAll(packages []Package, dir string, config RunConfig) (<-chan RunResult, error) {
	args := []string{"test", "-json"}
	if config.Verbose {
		args = []string{"test", "-v", "-json"}
	}
	args = append(args, "./...")

	cmd := exec.Command("go", args...)
	cmd.Dir = dir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	results := make(chan RunResult)

	go func() {
		parseOutput(stdout, results, config)
		cmd.Wait()
		close(results)
	}()

	return results, nil
}

func parseOutput(stdout io.Reader, results chan RunResult, config RunConfig) {
	scanner := bufio.NewScanner(stdout)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var event TestEvent
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			continue
		}

		switch event.Action {
		case "run":
			if event.Test != "" && !strings.Contains(event.Test, "/") {
				results <- RunResult{
					Package:  event.Package,
					TestName: event.Test,
					Status:   "running",
				}
			}

		case "output":
			if config.ShowOutput && event.Output != "" {
				results <- RunResult{
					Package:  event.Package,
					TestName: event.Test,
					Status:   "output",
					Output:   event.Output,
				}
			}

		case "pass", "fail", "skip":
			duration := time.Duration(event.Elapsed * float64(time.Second))
			testName := event.Test
			if testName == "" || strings.Contains(testName, "/") {
				continue
			}
			results <- RunResult{
				Package:  event.Package,
				TestName: testName,
				Status:   event.Action,
				Duration: duration,
				Output:   event.Output,
			}

		case "pause", "cont", "start":
		}
	}
}
