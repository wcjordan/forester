//go:build !js

package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"forester/game"
	"forester/render"
)

// shouldRunTUI reports whether --tui was passed as a CLI argument.
func shouldRunTUI() bool {
	for _, arg := range os.Args[1:] {
		if arg == "--tui" {
			return true
		}
	}
	return false
}

// runTUI starts the bubbletea terminal UI.
func runTUI() {
	g, err := game.LoadFromFile()
	if err != nil {
		g = game.New()
	}
	p := tea.NewProgram(render.NewModel(g), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
