package testrunner

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type TestFile struct {
	Path  string   `json:"path"`
	Tests []string `json:"tests"`
}

type Package struct {
	ImportPath string     `json:"importPath"`
	TestFiles  []TestFile `json:"testFiles"`
}

var testFuncRegex = regexp.MustCompile(`^func\s+(Test\w+)\s*\(`)

type rawPkgInfo struct {
	ImportPath   string   `json:"ImportPath"`
	Dir          string   `json:"Dir"`
	TestGoFiles  []string `json:"TestGoFiles"`
	XTestGoFiles []string `json:"XTestGoFiles"`
}

func Discover(rootDir string) ([]Package, error) {
	modPath := findGoMod(rootDir)
	if modPath == "" {
		return nil, fmt.Errorf("no go.mod found in %s", rootDir)
	}

	modDir := filepath.Dir(modPath)

	output, err := runGoList(modDir)
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	var packages []Package

	content := string(output)
	reader := strings.NewReader(content)
	scanner := bufio.NewScanner(reader)

	onObject := false
	var currentJSON strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		if line == "{" {
			onObject = true
		}

		if onObject {
			currentJSON.WriteString(line)
			currentJSON.WriteString("\n")
		}

		if line == "}" && onObject {
			onObject = false
			jsonStr := strings.TrimSpace(currentJSON.String())
			currentJSON.Reset()

			if jsonStr == "" {
				continue
			}

			var pkg rawPkgInfo
			if err := json.Unmarshal([]byte(jsonStr), &pkg); err != nil {
				continue
			}

			if len(pkg.TestGoFiles) == 0 && len(pkg.XTestGoFiles) == 0 {
				continue
			}

			var testFiles []TestFile

			for _, testFile := range pkg.TestGoFiles {
				fullPath := filepath.Join(pkg.Dir, testFile)
				tests, err := extractTestFuncs(fullPath)
				if err != nil || len(tests) == 0 {
					continue
				}
				pkgRelDir, _ := filepath.Rel(modDir, pkg.Dir)
				if pkgRelDir == "." {
					pkgRelDir = ""
				}
				relPath := filepath.Join(pkgRelDir, testFile)
				testFiles = append(testFiles, TestFile{
					Path:  relPath,
					Tests: tests,
				})
			}

			for _, testFile := range pkg.XTestGoFiles {
				fullPath := filepath.Join(pkg.Dir, testFile)
				tests, err := extractTestFuncs(fullPath)
				if err != nil || len(tests) == 0 {
					continue
				}
				pkgRelDir, _ := filepath.Rel(modDir, pkg.Dir)
				if pkgRelDir == "." {
					pkgRelDir = ""
				}
				relPath := filepath.Join(pkgRelDir, testFile)
				testFiles = append(testFiles, TestFile{
					Path:  relPath,
					Tests: tests,
				})
			}

			if len(testFiles) > 0 {
				packages = append(packages, Package{
					ImportPath: pkg.ImportPath,
					TestFiles:  testFiles,
				})
			}
		}
	}

	return packages, nil
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

func runGoList(dir string) ([]byte, error) {
	cmd := exec.Command("go", "list", "-json", "./...")
	cmd.Dir = dir
	return cmd.Output()
}

func extractTestFuncs(filePath string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var tests []string
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		matches := testFuncRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			tests = append(tests, matches[1])
		}
	}

	return tests, nil
}

func GetModuleRoot(dir string) string {
	modPath := findGoMod(dir)
	if modPath == "" {
		return dir
	}
	return filepath.Dir(modPath)
}
