package game

import "testing"

func TestMovePlayer(t *testing.T) {
	w := NewWorld(10, 10)
	p := NewPlayer(5, 5)

	p.MovePlayer(1, 0, w)
	if p.X != 6 || p.Y != 5 {
		t.Errorf("after move right: got (%d,%d), want (6,5)", p.X, p.Y)
	}

	p.MovePlayer(0, 1, w)
	if p.X != 6 || p.Y != 6 {
		t.Errorf("after move down: got (%d,%d), want (6,6)", p.X, p.Y)
	}

	p.MovePlayer(-1, 0, w)
	if p.X != 5 || p.Y != 6 {
		t.Errorf("after move left: got (%d,%d), want (5,6)", p.X, p.Y)
	}

	p.MovePlayer(0, -1, w)
	if p.X != 5 || p.Y != 5 {
		t.Errorf("after move up: got (%d,%d), want (5,5)", p.X, p.Y)
	}
}

func TestMovePlayerBounds(t *testing.T) {
	w := NewWorld(10, 10)

	// At left/top edge — cannot move further.
	p := NewPlayer(0, 0)
	p.MovePlayer(-1, 0, w)
	if p.X != 0 {
		t.Errorf("moved past left edge: X = %d, want 0", p.X)
	}
	p.MovePlayer(0, -1, w)
	if p.Y != 0 {
		t.Errorf("moved past top edge: Y = %d, want 0", p.Y)
	}

	// At right/bottom edge — cannot move further.
	p = NewPlayer(9, 9)
	p.MovePlayer(1, 0, w)
	if p.X != 9 {
		t.Errorf("moved past right edge: X = %d, want 9", p.X)
	}
	p.MovePlayer(0, 1, w)
	if p.Y != 9 {
		t.Errorf("moved past bottom edge: Y = %d, want 9", p.Y)
	}
}

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
