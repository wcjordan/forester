package game

import (
	"fmt"
	"math/rand"
	"time"
)

// Game is the top-level orchestrator that owns the game state and loop.
type Game struct {
	State          *State
	rng            *rand.Rand
	regrowCooldown time.Time
	clock          Clock
}

// New creates a new Game with default state using the system clock.
func New() *Game {
	return NewWithClock(RealClock{})
}

// NewWithClock creates a new Game with the given clock. Use in tests to
// inject a FakeClock for deterministic time control.
func NewWithClock(clock Clock) *Game {
	return &Game{
		State:          newState(),
		rng:            rand.New(rand.NewSource(time.Now().UnixNano())),
		regrowCooldown: clock.Now().Add(RegrowthCooldown),
		clock:          clock,
	}
}

// Tick advances the game: harvests trees, advances any in-progress build,
// handles adjacent-structure interactions, and fires probabilistic regrowth.
func (g *Game) Tick() {
	now := g.clock.Now()
	g.State.Harvest()
	g.State.AdvanceBuild()
	g.State.Player.TryDeposit(g.State, now)
	if now.After(g.regrowCooldown) {
		g.State.World.Regrow(g.rng)
		g.regrowCooldown = now.Add(RegrowthCooldown)
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
