package game

import (
	"testing"
	"time"

	"forester/game/internal/gametest"
)

func TestMovePlayer(t *testing.T) {
	w := NewWorld(10, 10)
	p := NewPlayer(5, 5)
	t0 := time.Now()

	p.Move(1, 0, w, t0)
	if p.X != 6 || p.Y != 5 {
		t.Errorf("after move right: got (%d,%d), want (6,5)", p.X, p.Y)
	}

	p.Move(0, 1, w, t0.Add(200*time.Millisecond))
	if p.X != 6 || p.Y != 6 {
		t.Errorf("after move down: got (%d,%d), want (6,6)", p.X, p.Y)
	}

	p.Move(-1, 0, w, t0.Add(400*time.Millisecond))
	if p.X != 5 || p.Y != 6 {
		t.Errorf("after move left: got (%d,%d), want (5,6)", p.X, p.Y)
	}

	p.Move(0, -1, w, t0.Add(600*time.Millisecond))
	if p.X != 5 || p.Y != 5 {
		t.Errorf("after move up: got (%d,%d), want (5,5)", p.X, p.Y)
	}
}

func TestMovePlayerBounds(t *testing.T) {
	w := NewWorld(10, 10)
	t0 := time.Now()

	// At left/top edge — cannot move further.
	p := NewPlayer(0, 0)
	p.Move(-1, 0, w, t0)
	if p.X != 0 {
		t.Errorf("moved past left edge: X = %d, want 0", p.X)
	}
	p.Move(0, -1, w, t0.Add(200*time.Millisecond))
	if p.Y != 0 {
		t.Errorf("moved past top edge: Y = %d, want 0", p.Y)
	}

	// At right/bottom edge — cannot move further.
	p = NewPlayer(9, 9)
	p.Move(1, 0, w, t0)
	if p.X != 9 {
		t.Errorf("moved past right edge: X = %d, want 9", p.X)
	}
	p.Move(0, 1, w, t0.Add(200*time.Millisecond))
	if p.Y != 9 {
		t.Errorf("moved past bottom edge: Y = %d, want 9", p.Y)
	}
}

func TestPlayerMoveCooldown(t *testing.T) {
	w := NewWorld(10, 10)
	p := NewPlayer(5, 5)
	t0 := time.Now()

	// First move always succeeds (Move cooldown unset — zero time is always expired).
	p.Move(1, 0, w, t0)
	if p.X != 6 {
		t.Fatalf("first move: X = %d, want 6", p.X)
	}

	// Same timestamp: cooldown not elapsed — move blocked.
	p.Move(1, 0, w, t0)
	if p.X != 6 {
		t.Errorf("same-timestamp move: X = %d, want 6 (cooldown should block)", p.X)
	}

	// After cooldown elapses: move succeeds.
	p.Move(1, 0, w, t0.Add(defaultMoveCooldown))
	if p.X != 7 {
		t.Errorf("after cooldown: X = %d, want 7", p.X)
	}
}

func TestMovePlayerStructureBlocking(t *testing.T) {
	t.Run("blocked by LogStorage", func(t *testing.T) {
		w := NewWorld(10, 10)
		w.PlaceBuilt(6, 5, gametest.LogStorageDef{})
		p := NewPlayer(5, 5)
		p.Move(1, 0, w, time.Now()) // try to move into (6,5)
		if p.X != 5 {
			t.Errorf("X = %d, want 5 (should be blocked by LogStorage)", p.X)
		}
	})

	t.Run("blocked by FoundationLogStorage", func(t *testing.T) {
		w := NewWorld(10, 10)
		w.PlaceFoundation(6, 5, gametest.LogStorageDef{})
		p := NewPlayer(5, 5)
		p.Move(1, 0, w, time.Now())
		if p.X != 5 {
			t.Errorf("X = %d, want 5 (foundation tiles should block movement)", p.X)
		}
	})
}

func TestMoveCooldowns(t *testing.T) {
	forestCooldown := MoveCooldownFor(&Tile{Terrain: Forest, TreeSize: 5})
	cutTreeCooldown := MoveCooldownFor(&Tile{Terrain: Forest, TreeSize: 0})
	grassCooldown := MoveCooldownFor(&Tile{Terrain: Grassland})

	if forestCooldown <= grassCooldown {
		t.Errorf("Forest cooldown (%v) should be longer than Grassland (%v)", forestCooldown, grassCooldown)
	}
	if forestCooldown <= cutTreeCooldown {
		t.Errorf("Forest cooldown (%v) should be longer than cut tree (%v)", forestCooldown, cutTreeCooldown)
	}
	if grassCooldown != cutTreeCooldown {
		t.Errorf("Grassland (%v) and cut tree (%v) cooldowns should be equal", grassCooldown, cutTreeCooldown)
	}
}

func TestRoadLevelFor(t *testing.T) {
	cases := []struct {
		terrain   TerrainType
		walkCount int
		want      int
	}{
		{Grassland, 0, 0},
		{Grassland, WalkCountTrodden - 1, 0},
		{Grassland, WalkCountTrodden, 1},
		{Grassland, WalkCountRoad - 1, 1},
		{Grassland, WalkCountRoad, 2},
		{Grassland, WalkCountRoad + 100, 2},
		{Forest, WalkCountRoad, 0}, // Forest tiles are never roads
	}
	for _, c := range cases {
		tile := &Tile{Terrain: c.terrain, WalkCount: c.walkCount}
		if got := RoadLevelFor(tile); got != c.want {
			t.Errorf("RoadLevelFor(%v, wc=%d) = %d, want %d", c.terrain, c.walkCount, got, c.want)
		}
	}
}

func TestPlayerMove_IncrementsWalkCount(t *testing.T) {
	w := NewWorld(10, 10)
	p := NewPlayer(5, 5)
	t0 := time.Now()

	// Grassland destination: WalkCount should increment.
	p.Move(1, 0, w, t0) // moves to (6,5)
	tile := w.TileAt(6, 5)
	if tile.WalkCount != 1 {
		t.Errorf("Grassland tile WalkCount = %d, want 1", tile.WalkCount)
	}

	// Forest destination: WalkCount should NOT increment.
	w.TileAt(5, 5).Terrain = Forest
	w.TileAt(5, 5).TreeSize = 5
	p2 := NewPlayer(4, 5)
	p2.Move(1, 0, w, t0) // moves to (5,5) — Forest
	forestTile := w.TileAt(5, 5)
	if forestTile.WalkCount != 0 {
		t.Errorf("Forest tile WalkCount = %d, want 0", forestTile.WalkCount)
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
	if p.Inventory[Wood] != 0 {
		t.Errorf("Inventory[Wood] = %d, want 0", p.Inventory[Wood])
	}
	if p.FacingDX != 0 || p.FacingDY != -1 {
		t.Errorf("facing = (%d,%d), want (0,-1)", p.FacingDX, p.FacingDY)
	}
}
