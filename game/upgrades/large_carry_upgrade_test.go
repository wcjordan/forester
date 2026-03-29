package upgrades

import (
	"math/rand"
	"testing"

	"forester/game"
	"forester/game/geom"
)

func makeUpgradeTestGame() *game.Game {
	g := game.NewWithClockAndRNG(game.RealClock{}, rand.New(rand.NewSource(0)))
	w := game.NewWorld(20, 20)
	p := game.NewPlayer(5, 5)
	g.State = &game.State{
		Player:              p,
		World:               w,
		FoundationDeposited: make(map[geom.Point]int),
		HouseOccupancy:      make(map[geom.Point]bool),
	}
	g.Stores = game.NewStorageManager()
	g.Villagers = game.NewVillagerManager()
	return g
}

// TestLargeCarryUpgradeApply verifies that Apply adds 100 to MaxCarry.
func TestLargeCarryUpgradeApply(t *testing.T) {
	g := makeUpgradeTestGame()
	initial := g.State.Player.MaxCarry

	env := &game.Env{State: g.State, Stores: g.Stores, Villagers: g.Villagers, RNG: rand.New(rand.NewSource(0))}
	u := largeCarryCapacityUpgrade{}
	u.Apply(env)

	want := initial + 100
	if g.State.Player.MaxCarry != want {
		t.Errorf("MaxCarry = %d, want %d (initial %d + 100)", g.State.Player.MaxCarry, want, initial)
	}
}

// TestLargeCarryUpgradeIsAdditive verifies that applying the upgrade twice adds 200.
func TestLargeCarryUpgradeIsAdditive(t *testing.T) {
	g := makeUpgradeTestGame()
	initial := g.State.Player.MaxCarry

	env := &game.Env{State: g.State, Stores: g.Stores, Villagers: g.Villagers, RNG: rand.New(rand.NewSource(0))}
	u := largeCarryCapacityUpgrade{}
	u.Apply(env)
	u.Apply(env)

	want := initial + 200
	if g.State.Player.MaxCarry != want {
		t.Errorf("MaxCarry = %d, want %d after two applications", g.State.Player.MaxCarry, want)
	}
}
