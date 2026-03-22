package main

import (
	"fmt"
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/heywinit/alout/internal/testrunner"
	"github.com/heywinit/alout/internal/tui"
)

func main() {
	rootDir := "."
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}

	packages, err := testrunner.Discover(rootDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error discovering tests: %v\n", err)
		os.Exit(1)
	}

	if len(packages) == 0 {
		fmt.Println("No tests found.")
		os.Exit(0)
	}

	model := tui.NewModel()
	model.Packages = packages

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}
