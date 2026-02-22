package game

import (
	"fmt"
	"math/rand"
	"time"
)

// DepositTickInterval is how often the player auto-deposits one wood when adjacent to a storage structure.
const DepositTickInterval = 500 * time.Millisecond

// Game is the top-level orchestrator that owns the game state and loop.
type Game struct {
	State           *State
	depositCooldown time.Time
	rng             *rand.Rand
	regrowCooldown  time.Time
}

// New creates a new Game with default state.
func New() *Game {
	return &Game{
		State:          newState(),
		rng:            rand.New(rand.NewSource(time.Now().UnixNano())),
		regrowCooldown: time.Now().Add(RegrowthCooldown),
	}
}

// Tick advances the game: harvests trees, advances any in-progress build,
// handles adjacent-structure interactions, and fires probabilistic regrowth.
func (g *Game) Tick() {
	g.State.Harvest()
	g.State.AdvanceBuild()
	if time.Now().After(g.depositCooldown) {
		before := g.State.LogStorageDeposited
		g.State.TickAdjacentStructures()
		if g.State.LogStorageDeposited > before {
			g.depositCooldown = time.Now().Add(DepositTickInterval)
		}
	}
	if time.Now().After(g.regrowCooldown) {
		g.State.World.Regrow(g.rng)
		g.regrowCooldown = time.Now().Add(RegrowthCooldown)
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
