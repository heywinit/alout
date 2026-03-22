package testrunner

import (
	"bufio"
	"encoding/json"
	"io"
	"os/exec"
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
	Verbose      bool
	ShowOutput   bool
	OutputFormat string
}

func Run(pkg Package, testName string, dir string, config RunConfig) (<-chan RunResult, error) {
	args := buildArgs(config)

	if testName != "" {
		args = append(args, "-run", testName)
	}

	cmd := exec.Command("go", args...)
	cmd.Dir = dir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	results := make(chan RunResult)
	go parseJSONOutput(stdout, stderr, results, config)

	go func() {
		cmd.Wait()
		close(results)
	}()

	return results, nil
}

func RunPackage(pkg Package, dir string, config RunConfig) (<-chan RunResult, error) {
	args := buildArgs(config)

	cmd := exec.Command("go", args...)
	cmd.Dir = dir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	results := make(chan RunResult)
	go parseJSONOutput(stdout, stderr, results, config)

	go func() {
		cmd.Wait()
		close(results)
	}()

	return results, nil
}

func RunAll(packages []Package, dir string, config RunConfig) (<-chan RunResult, error) {
	args := buildArgs(config)
	args = append(args, "./...")

	cmd := exec.Command("go", args...)
	cmd.Dir = dir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	results := make(chan RunResult)
	go parseJSONOutput(stdout, stderr, results, config)

	go func() {
		cmd.Wait()
		close(results)
	}()

	return results, nil
}

func buildArgs(config RunConfig) []string {
	args := []string{"test", "-json"}

	switch config.OutputFormat {
	case "verbose":
		args = []string{"test", "-v", "-json"}
	case "summary":
		args = []string{"test", "-json"}
	case "quiet":
		args = []string{"test", "-json"}
	}

	return args
}

func parseJSONOutput(stdout io.Reader, stderr io.Reader, results chan RunResult, config RunConfig) {
	scanner := bufio.NewScanner(stdout)
	var currentTest, currentPackage, currentOutput string

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
			if currentTest != "" {
				results <- RunResult{
					Package:  currentPackage,
					TestName: currentTest,
					Status:   "running",
				}
			}
			currentTest = event.Test
			currentPackage = event.Package
			currentOutput = ""

		case "output":
			if config.ShowOutput {
				currentOutput += event.Output
			}

		case "pass", "fail", "skip":
			duration := time.Duration(event.Elapsed * float64(time.Second))
			results <- RunResult{
				Package:  currentPackage,
				TestName: currentTest,
				Status:   event.Action,
				Duration: duration,
				Output:   currentOutput,
			}
			currentTest = ""
			currentPackage = ""
			currentOutput = ""

		case "pause", "cont":
		}
	}

	if currentTest != "" {
		results <- RunResult{
			Package:  currentPackage,
			TestName: currentTest,
			Status:   "unknown",
			Output:   currentOutput,
		}
	}
}
