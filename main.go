package main

import (
	"fmt"
	"os"

	ebiten "github.com/hajimehoshi/ebiten/v2"

	"forester/game"
	_ "forester/game/resources"
	_ "forester/game/structures"
	_ "forester/game/upgrades"
	gui "forester/render/gui"
)

func main() {
	if shouldRunTUI() {
		runTUI()
		return
	}

	g := game.New()
	g.Load()
	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("Forester")
	if err := ebiten.RunGame(gui.NewEbitenGame(g)); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
