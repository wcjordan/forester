package game

import "testing"

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
