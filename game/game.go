package game

import "fmt"

// Game is the top-level orchestrator that owns the game state and loop.
type Game struct {
	State *State
}

// New creates a new Game with default state.
func New() *Game {
	return &Game{
		State: newState(),
	}
}

// Run starts the game loop. For now it just prints a message and returns.
func (g *Game) Run() error {
	fmt.Println("Forester - A village grows where you walk")
	fmt.Printf("World: %dx%d | Player at (%d, %d)\n",
		g.State.World.Width, g.State.World.Height,
		g.State.Player.X, g.State.Player.Y,
	)
	return nil
}
