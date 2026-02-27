package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"forester/game"
	_ "forester/game/resources"
	_ "forester/game/structures"
	_ "forester/game/upgrades"
	"forester/render"
)

func main() {
	g := game.New()
	p := tea.NewProgram(render.NewModel(g), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
