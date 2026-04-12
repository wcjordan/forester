package game

import (
	"testing"
	"time"

	"forester/game/internal/gametest"
)

func TestMovePlayer(t *testing.T) {
	w := NewWorld(10, 10)
	// 200ms at DefaultMoveSpeed ≈ 1.33 tiles. Starting from tile center (5.5)
	// keeps the player in the adjacent tile for both positive and negative directions.
	dt := 200 * time.Millisecond

	for _, tc := range []struct {
		name   string
		dx, dy float64
		wantX  int
		wantY  int
	}{
		{"right", 1, 0, 6, 5},
		{"down", 0, 1, 5, 6},
		{"left", -1, 0, 4, 5},
		{"up", 0, -1, 5, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := NewPlayer(5, 5)
			p.PosX, p.PosY = 5.5, 5.5 // tile center avoids boundary asymmetry
			p.MoveSmooth(tc.dx, tc.dy, w, dt)
			if p.TileX() != tc.wantX || p.TileY() != tc.wantY {
				t.Errorf("got (%d,%d), want (%d,%d)", p.TileX(), p.TileY(), tc.wantX, tc.wantY)
			}
		})
	}
}

func TestMovePlayerBounds(t *testing.T) {
	w := NewWorld(10, 10)
	dt := 200 * time.Millisecond

	// At left/top edge — cannot move further.
	p := NewPlayer(0, 0)
	p.MoveSmooth(-1, 0, w, dt)
	if p.TileX() != 0 {
		t.Errorf("moved past left edge: X = %d, want 0", p.TileX())
	}
	p.MoveSmooth(0, -1, w, dt)
	if p.TileY() != 0 {
		t.Errorf("moved past top edge: Y = %d, want 0", p.TileY())
	}

	// At right/bottom edge — cannot move further.
	p = NewPlayer(9, 9)
	p.MoveSmooth(1, 0, w, dt)
	if p.TileX() != 9 {
		t.Errorf("moved past right edge: X = %d, want 9", p.TileX())
	}
	p.MoveSmooth(0, 1, w, dt)
	if p.TileY() != 9 {
		t.Errorf("moved past bottom edge: Y = %d, want 9", p.TileY())
	}
}

func TestTerrainSpeedFor(t *testing.T) {
	forest := TerrainSpeedFor(&Tile{Terrain: Forest, TreeSize: 5})
	grass := TerrainSpeedFor(&Tile{Terrain: Grassland})
	cutTree := TerrainSpeedFor(&Tile{Terrain: Forest, TreeSize: 0})
	trodden := TerrainSpeedFor(&Tile{Terrain: Grassland, WalkCount: WalkCountTrodden})
	road := TerrainSpeedFor(&Tile{Terrain: Grassland, WalkCount: WalkCountRoad})
	nilTile := TerrainSpeedFor(nil)

	if forest >= grass {
		t.Errorf("Forest speed %v should be < Grassland %v", forest, grass)
	}
	if cutTree != grass {
		t.Errorf("cut tree speed %v should equal Grassland %v", cutTree, grass)
	}
	if trodden <= grass {
		t.Errorf("trodden speed %v should be > Grassland %v", trodden, grass)
	}
	if road <= trodden {
		t.Errorf("road speed %v should be > trodden %v", road, trodden)
	}
	if nilTile != 1.0 {
		t.Errorf("nil tile speed = %v, want 1.0", nilTile)
	}
}

func TestTerrainSpeedFor_RoadLevels(t *testing.T) {
	plain := &Tile{Terrain: Grassland, WalkCount: 0}
	trodden := &Tile{Terrain: Grassland, WalkCount: WalkCountTrodden}
	road := &Tile{Terrain: Grassland, WalkCount: WalkCountRoad}
	forest := &Tile{Terrain: Forest, TreeSize: 5, WalkCount: WalkCountRoad}

	if got := TerrainSpeedFor(plain); got != 1.0 {
		t.Errorf("plain Grassland: %v, want 1.0", got)
	}
	if got := TerrainSpeedFor(trodden); got != troddenSpeedFactor {
		t.Errorf("trodden: %v, want %v", got, troddenSpeedFactor)
	}
	if got := TerrainSpeedFor(road); got != roadSpeedFactor {
		t.Errorf("road: %v, want %v", got, roadSpeedFactor)
	}
	// Forest tiles ignore WalkCount.
	if got := TerrainSpeedFor(forest); got != forestSpeedFactor {
		t.Errorf("Forest with high WalkCount: %v, want %v", got, forestSpeedFactor)
	}
}

func TestMovePlayerStructureBlocking(t *testing.T) {
	dt := 200 * time.Millisecond

	t.Run("blocked by LogStorage", func(t *testing.T) {
		w := NewWorld(10, 10)
		w.PlaceBuilt(6, 5, gametest.LogStorageDef{})
		p := NewPlayer(5, 5)
		p.MoveSmooth(1, 0, w, dt) // try to move into (6,5)
		if p.TileX() != 5 {
			t.Errorf("X = %d, want 5 (should be blocked by LogStorage)", p.TileX())
		}
	})

	t.Run("blocked by FoundationLogStorage", func(t *testing.T) {
		w := NewWorld(10, 10)
		w.PlaceFoundation(6, 5, gametest.LogStorageDef{})
		p := NewPlayer(5, 5)
		p.MoveSmooth(1, 0, w, dt)
		if p.TileX() != 5 {
			t.Errorf("X = %d, want 5 (foundation tiles should block movement)", p.TileX())
		}
	})
}

func TestMoveCost_RoadLevels(t *testing.T) {
	w := NewWorld(5, 5)

	// Default Grassland: cost should be roadSpeedFactor / 1.0.
	wantGrass := roadSpeedFactor / TerrainSpeedFor(&Tile{Terrain: Grassland})
	if got := w.MoveCost(2, 2); got != wantGrass {
		t.Errorf("Grassland MoveCost = %v, want %v", got, wantGrass)
	}

	// Road tile: cost should be exactly 1.0.
	w.TileAt(2, 2).WalkCount = WalkCountRoad
	if got := w.MoveCost(2, 2); got != 1.0 {
		t.Errorf("Road MoveCost = %v, want 1.0", got)
	}

	// All terrain types must have MoveCost >= 1.0 (A* admissibility).
	for _, tile := range []*Tile{
		{Terrain: Grassland, WalkCount: 0},
		{Terrain: Grassland, WalkCount: WalkCountTrodden},
		{Terrain: Grassland, WalkCount: WalkCountRoad},
		{Terrain: Forest, TreeSize: 5},
		{Terrain: Forest, TreeSize: 0},
	} {
		cost := roadSpeedFactor / TerrainSpeedFor(tile)
		if cost < 1.0 {
			t.Errorf("MoveCost for tile %+v = %v < 1.0; breaks A* admissibility", tile, cost)
		}
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

func TestMoveSmooth_SubTile(t *testing.T) {
	w := NewWorld(10, 10)
	p := NewPlayer(5, 5)

	// 50ms is less than one full tile at default cooldown (150ms), so position
	// should advance but stay within tile 5.
	p.MoveSmooth(1, 0, w, 50*time.Millisecond)

	if p.PosX <= 5.0 || p.PosX >= 6.0 {
		t.Errorf("PosX = %v, want in (5.0, 6.0)", p.PosX)
	}
	if p.TileX() != 5 {
		t.Errorf("X = %d, want 5 (still within tile 5)", p.TileX())
	}
}

func TestMoveSmooth_TileCrossing(t *testing.T) {
	w := NewWorld(10, 10)
	p := NewPlayer(5, 5)

	// 200ms is more than one full tile at default cooldown (150ms).
	p.MoveSmooth(1, 0, w, 200*time.Millisecond)

	if p.TileX() != 6 {
		t.Errorf("X = %d, want 6 (crossed tile boundary)", p.TileX())
	}
	if p.PosX < 6.0 {
		t.Errorf("PosX = %v, want >= 6.0", p.PosX)
	}
}

func TestMoveSmooth_Collision(t *testing.T) {
	w := NewWorld(10, 10)
	w.PlaceBuilt(6, 5, gametest.LogStorageDef{})
	p := NewPlayer(5, 5)

	// Would cross into tile 6, but it is blocked by a structure.
	p.MoveSmooth(1, 0, w, 200*time.Millisecond)

	if p.TileX() != 5 {
		t.Errorf("X = %d, want 5 (blocked by structure)", p.TileX())
	}
	if p.PosX >= 6.0 {
		t.Errorf("PosX = %v, should be < 6.0 (stopped at boundary)", p.PosX)
	}
}

func TestMoveSmooth_Bounds(t *testing.T) {
	w := NewWorld(10, 10)
	p := NewPlayer(0, 5)

	// Moving left at the left edge: should stop at boundary.
	p.MoveSmooth(-1, 0, w, 200*time.Millisecond)

	if p.PosX < 0 {
		t.Errorf("PosX = %v, should be >= 0 (world boundary)", p.PosX)
	}
	if p.TileX() < 0 {
		t.Errorf("X = %d, should be >= 0", p.TileX())
	}
}

func TestMoveSmooth_WalkCount(t *testing.T) {
	w := NewWorld(10, 10)
	p := NewPlayer(5, 5)

	p.MoveSmooth(1, 0, w, 200*time.Millisecond) // crosses into tile (6,5)

	if w.TileAt(6, 5).WalkCount != 1 {
		t.Errorf("WalkCount = %d, want 1", w.TileAt(6, 5).WalkCount)
	}
}

func TestMoveSmooth_MultiTile(t *testing.T) {
	w := NewWorld(10, 10)
	p := NewPlayer(5, 5)

	// 400ms at DefaultMoveSpeed (≈6.667 tiles/sec) = ≈2.667 tiles — crosses tiles 6 and 7.
	p.MoveSmooth(1, 0, w, 400*time.Millisecond)

	if p.TileX() < 7 {
		t.Errorf("X = %d, want >= 7 (should cross multiple tiles)", p.TileX())
	}
	// Both tiles 6 and 7 should have been entered.
	if w.TileAt(6, 5).WalkCount != 1 {
		t.Errorf("tile (6,5) WalkCount = %d, want 1", w.TileAt(6, 5).WalkCount)
	}
	if w.TileAt(7, 5).WalkCount != 1 {
		t.Errorf("tile (7,5) WalkCount = %d, want 1", w.TileAt(7, 5).WalkCount)
	}
}

func TestMoveSmooth_MultiTileBlockedMidway(t *testing.T) {
	w := NewWorld(10, 10)
	// Block tile 7, leave 6 open.
	w.PlaceBuilt(7, 5, gametest.LogStorageDef{})
	p := NewPlayer(5, 5)

	// Large dt would reach tile 7, but it's blocked — should stop at 6.
	p.MoveSmooth(1, 0, w, 400*time.Millisecond)

	if p.TileX() != 6 {
		t.Errorf("X = %d, want 6 (stopped before blocked tile 7)", p.TileX())
	}
	if p.PosX >= 7.0 {
		t.Errorf("PosX = %v, should be < 7.0", p.PosX)
	}
}

func TestMoveSmooth_FacingUpdated(t *testing.T) {
	w := NewWorld(10, 10)
	p := NewPlayer(5, 5)

	p.MoveSmooth(0, 1, w, 10*time.Millisecond) // move down

	if p.FacingDX != 0 || p.FacingDY != 1 {
		t.Errorf("facing = (%d,%d), want (0,1)", p.FacingDX, p.FacingDY)
	}
}

func TestPlayerMove_IncrementsWalkCount(t *testing.T) {
	w := NewWorld(10, 10)
	dt := 200 * time.Millisecond

	// Grassland destination: WalkCount should increment.
	p := NewPlayer(5, 5)
	p.MoveSmooth(1, 0, w, dt) // moves to (6,5)
	if w.TileAt(6, 5).WalkCount != 1 {
		t.Errorf("Grassland tile WalkCount = %d, want 1", w.TileAt(6, 5).WalkCount)
	}

	// Forest destination: WalkCount should NOT increment (not road-eligible).
	w.TileAt(5, 5).Terrain = Forest
	w.TileAt(5, 5).TreeSize = 5
	p2 := NewPlayer(4, 5)
	p2.MoveSmooth(1, 0, w, dt) // moves to (5,5) — Forest
	if w.TileAt(5, 5).WalkCount != 0 {
		t.Errorf("Forest tile WalkCount = %d, want 0", w.TileAt(5, 5).WalkCount)
	}
}

func TestNewPlayer(t *testing.T) {
	p := NewPlayer(10, 20)

	if p.TileX() != 10 {
		t.Errorf("X = %d, want 10", p.TileX())
	}
	if p.TileY() != 20 {
		t.Errorf("Y = %d, want 20", p.TileY())
	}
	if p.Inventory[Wood] != 0 {
		t.Errorf("Inventory[Wood] = %d, want 0", p.Inventory[Wood])
	}
	if p.FacingDX != 0 || p.FacingDY != -1 {
		t.Errorf("facing = (%d,%d), want (0,-1)", p.FacingDX, p.FacingDY)
	}
}
