package game

import (
	"math/rand"
	"time"
)

// Game is the top-level orchestrator that owns the game state and loop.
type Game struct {
	State          *State
	Stores         *StorageManager
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
	return NewWithClockAndRNG(clock, rand.New(rand.NewSource(time.Now().UnixNano())))
}

// NewWithClockAndRNG creates a new Game with injected clock and RNG. Use in
// tests to get fully deterministic behavior (e.g. rand.New(rand.NewSource(0))).
func NewWithClockAndRNG(clock Clock, rng *rand.Rand) *Game {
	return &Game{
		State:          newState(),
		Stores:         NewStorageManager(),
		rng:            rng,
		regrowCooldown: clock.Now().Add(RegrowthCooldown),
		clock:          clock,
	}
}

// env returns the runtime context for the current tick.
func (g *Game) env() *Env {
	return &Env{State: g.State, Stores: g.Stores}
}

// Tick advances the game: harvests trees, handles adjacent-structure interactions,
// and fires probabilistic regrowth.
func (g *Game) Tick() {
	now := g.clock.Now()
	env := g.env()
	g.State.Harvest(env)
	g.State.TickAdjacentStructures(env, now)
	if now.After(g.regrowCooldown) {
		g.State.World.Regrow(g.rng)
		g.regrowCooldown = now.Add(RegrowthCooldown)
	}
}
