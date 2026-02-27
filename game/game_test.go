package game

import (
	"math/rand"
	"testing"
)

func TestNew(t *testing.T) {
	g := New()

	if g == nil {
		t.Fatal("New() returned nil")
	}
	if g.State == nil {
		t.Fatal("game state is nil")
	}
	if g.State.Player == nil {
		t.Fatal("player is nil")
	}
	if g.State.World == nil {
		t.Fatal("world is nil")
	}
}

func TestNewPlayerPosition(t *testing.T) {
	g := New()

	// Player should start at center of 100x100 world
	if g.State.Player.X != 50 {
		t.Errorf("player X = %d, want 50", g.State.Player.X)
	}
	if g.State.Player.Y != 50 {
		t.Errorf("player Y = %d, want 50", g.State.Player.Y)
	}
}

// testCarryUpgrade is a minimal UpgradeDef used only in TestAddOfferAndSelectCard
// so the test stays decoupled from the game/upgrades package.
type testCarryUpgrade struct{}

func (testCarryUpgrade) ID() string          { return "test_carry" }
func (testCarryUpgrade) Name() string        { return "Test Carry" }
func (testCarryUpgrade) Description() string { return "test" }
func (testCarryUpgrade) Apply(p *Player)     { p.MaxCarry = 100 }

func TestAddOfferAndSelectCard(t *testing.T) {
	upgradeRegistry["test_carry"] = testCarryUpgrade{}
	t.Cleanup(func() { delete(upgradeRegistry, "test_carry") })

	g := NewWithClockAndRNG(RealClock{}, rand.New(rand.NewSource(1)))
	g.State.Player = NewPlayer(5, 5)
	g.State.World = NewWorld(10, 10)
	g.State.FoundationDeposited = make(map[Point]int)
	g.State.CompletedBeats = make(map[string]bool)
	p := g.State.Player

	if g.HasPendingOffer() {
		t.Fatal("should have no pending offer initially")
	}

	g.State.AddOffer([]string{"test_carry"})

	if !g.HasPendingOffer() {
		t.Fatal("should have pending offer after AddOffer")
	}

	g.SelectCard(0)

	if g.HasPendingOffer() {
		t.Error("should have no pending offer after SelectCard")
	}
	if p.MaxCarry != 100 {
		t.Errorf("MaxCarry = %d, want 100 after carry capacity upgrade", p.MaxCarry)
	}
}
