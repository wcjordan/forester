package render

import (
	ebiten "github.com/hajimehoshi/ebiten/v2"

	"forester/game"
)

const (
	tileSize     = 32
	screenWidth  = 1280
	screenHeight = 720
)

// EbitenGame implements ebiten.Game and renders the world using solid-color rectangles.
type EbitenGame struct {
	game *game.Game
}

// NewEbitenGame creates an EbitenGame wrapping the given game.
func NewEbitenGame(g *game.Game) *EbitenGame {
	return &EbitenGame{game: g}
}

// Update is called every frame by Ebitengine. Placeholder for Stage 2 implementation.
func (e *EbitenGame) Update() error {
	return nil
}

// Draw is called every frame by Ebitengine. Placeholder for Stage 2 implementation.
func (e *EbitenGame) Draw(_ *ebiten.Image) {}

// Layout returns the logical screen dimensions.
func (e *EbitenGame) Layout(_, _ int) (w, h int) {
	return screenWidth, screenHeight
}
