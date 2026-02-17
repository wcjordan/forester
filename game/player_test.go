package game

import "testing"

func TestNewPlayer(t *testing.T) {
	p := NewPlayer(10, 20)

	if p.X != 10 {
		t.Errorf("X = %d, want 10", p.X)
	}
	if p.Y != 20 {
		t.Errorf("Y = %d, want 20", p.Y)
	}
	if p.Wood != 0 {
		t.Errorf("Wood = %d, want 0", p.Wood)
	}
}
