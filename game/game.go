package game

import (
	"math/rand"
	"time"
)

// Game is the top-level orchestrator that owns the game state and loop.
type Game struct {
	State     *State
	Stores    *StorageManager
	Villagers *VillagerManager
	Status    SaveStatus
	rng       *rand.Rand
	clock     Clock
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
		State:     newState(),
		Stores:    NewStorageManager(),
		Villagers: NewVillagerManager(),
		rng:       rng,
		clock:     clock,
	}
}

// env returns the runtime context for the current tick.
func (g *Game) env() *Env {
	return &Env{State: g.State, Stores: g.Stores, Villagers: g.Villagers, RNG: g.rng}
}

// Tick advances the game: harvests trees, handles adjacent-structure interactions,
// ticks villagers, and fires probabilistic regrowth. Returns early when a card offer is pending.
func (g *Game) Tick() {
	if g.HasPendingOffer() {
		return
	}
	now := g.clock.Now()
	env := g.env()
	IterateResources(func(d ResourceDef) { d.Harvest(env, now) })
	maybeAdvanceStory(env)
	maybeSpawnFoundation(env)
	g.TickAdjacentStructures(now)
	g.Villagers.Tick(env, g.rng, now)
	IterateResources(func(d ResourceDef) { d.Regrow(env, g.rng, now) })
}

// HasPendingOffer reports whether there is at least one offer waiting.
func (g *Game) HasPendingOffer() bool {
	return len(g.State.pendingOfferIDs) > 0
}

// CurrentOffer resolves the front offer's IDs to UpgradeDef values.
// Returns nil when there is no pending offer or no IDs resolve.
func (g *Game) CurrentOffer() []UpgradeDef {
	if len(g.State.pendingOfferIDs) == 0 {
		return nil
	}
	ids := g.State.pendingOfferIDs[0]
	result := make([]UpgradeDef, 0, len(ids))
	for _, id := range ids {
		if u, ok := upgradeRegistry[id]; ok {
			result = append(result, u)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// TickAdjacentStructures calls OnPlayerInteraction once per structure instance
// that the player is cardinally adjacent to, then commits any pending cooldowns.
// Cooldowns are committed after all interactions so that multiple adjacent
// structures of the same type all fire within the same tick.
func (g *Game) TickAdjacentStructures(now time.Time) {
	env := g.env()
	s := g.State
	seen := make(map[point]bool)
	for _, d := range [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
		p := point{X: s.Player.X + d[0], Y: s.Player.Y + d[1]}
		entry, ok := s.World.structureIndex[p]
		if !ok || seen[entry.Origin] {
			continue
		}
		seen[entry.Origin] = true
		cb := lookupCallbacks(entry.Def.BuiltType())
		if cb.OnPlayerInteraction != nil {
			cb.OnPlayerInteraction(env, entry.Origin, now)
		}
	}
	s.Player.commitCooldowns()
}

// XPInfo returns the player's current XP and the threshold for the next milestone.
func (g *Game) XPInfo() (xp, nextMilestone int) {
	return g.State.XP, xpMilestoneAt(g.State.XPMilestoneIdx)
}

// SelectCard applies the card at idx from the front offer and pops it from the queue.
func (g *Game) SelectCard(idx int) {
	offer := g.CurrentOffer()
	if len(offer) == 0 {
		return
	}
	if idx >= 0 && idx < len(offer) {
		offer[idx].Apply(g.env())
		g.State.pendingOfferIDs = g.State.pendingOfferIDs[1:]
	}
}
