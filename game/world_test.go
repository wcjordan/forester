package game

import (
	"math/rand"
	"testing"
	"time"
)

func TestNewWorld(t *testing.T) {
	w := NewWorld(50, 30)

	if w.Width != 50 {
		t.Errorf("Width = %d, want 50", w.Width)
	}
	if w.Height != 30 {
		t.Errorf("Height = %d, want 30", w.Height)
	}
	if len(w.Tiles) != 30 {
		t.Errorf("tile rows = %d, want 30", len(w.Tiles))
	}
	if len(w.Tiles[0]) != 50 {
		t.Errorf("tile cols = %d, want 50", len(w.Tiles[0]))
	}
}

func TestNewWorldDefaultTerrain(t *testing.T) {
	w := NewWorld(10, 10)

	for y := range w.Tiles {
		for x := range w.Tiles[y] {
			if w.Tiles[y][x].Terrain != Grassland {
				t.Errorf("tile (%d,%d) terrain = %d, want Grassland", x, y, w.Tiles[y][x].Terrain)
			}
		}
	}
}

func TestInBounds(t *testing.T) {
	w := NewWorld(10, 10)

	tests := []struct {
		x, y int
		want bool
	}{
		{0, 0, true},
		{9, 9, true},
		{5, 5, true},
		{-1, 0, false},
		{0, -1, false},
		{10, 0, false},
		{0, 10, false},
	}

	for _, tt := range tests {
		got := w.InBounds(tt.x, tt.y)
		if got != tt.want {
			t.Errorf("InBounds(%d, %d) = %v, want %v", tt.x, tt.y, got, tt.want)
		}
	}
}

// regrowTick advances the regrowth cooldown by 2x RegrowthCooldown per iteration,
// guaranteeing each call fires the regrowth logic regardless of current cooldown state.
func regrowTick(w *World, rng *rand.Rand, i int) {
	t0 := time.Time{}
	w.Regrow(rng, t0.Add(time.Duration(i+1)*RegrowthCooldown*2))
}

func TestRegrow(t *testing.T) {
	// Use a 20×20 world and place Forest tiles at (0,0), which is ~14 tiles
	// from the spawn center (10,10) — well outside the no-grow radius of 8.
	t.Run("cut tree eventually grows", func(t *testing.T) {
		rng := rand.New(rand.NewSource(0))
		w := NewWorld(20, 20)
		w.Tiles[0][0] = Tile{Terrain: Forest, TreeSize: 0}
		grew := false
		for i := 0; i < 1000; i++ {
			regrowTick(w, rng, i)
			if w.Tiles[0][0].TreeSize > 0 {
				grew = true
				break
			}
		}
		if !grew {
			t.Error("cut tree (Forest/TreeSize=0) should eventually grow")
		}
	})

	t.Run("forest eventually grows toward maxTreeSize", func(t *testing.T) {
		rng := rand.New(rand.NewSource(0))
		w := NewWorld(20, 20)
		w.Tiles[0][0] = Tile{Terrain: Forest, TreeSize: 5}
		grew := false
		for i := 0; i < 1000; i++ {
			regrowTick(w, rng, i)
			if w.Tiles[0][0].TreeSize > 5 {
				grew = true
				break
			}
		}
		if !grew {
			t.Error("forest should eventually grow toward maxTreeSize")
		}
	})

	t.Run("forest at maxTreeSize does not grow further", func(t *testing.T) {
		rng := rand.New(rand.NewSource(0))
		w := NewWorld(20, 20)
		w.Tiles[0][0] = Tile{Terrain: Forest, TreeSize: maxTreeSize}
		for i := 0; i < 1000; i++ {
			regrowTick(w, rng, i)
		}
		if w.Tiles[0][0].TreeSize != maxTreeSize {
			t.Errorf("TreeSize = %d, want %d", w.Tiles[0][0].TreeSize, maxTreeSize)
		}
	})

	t.Run("grassland is unaffected", func(t *testing.T) {
		rng := rand.New(rand.NewSource(0))
		w := NewWorld(20, 20)
		w.Tiles[0][0] = Tile{Terrain: Grassland}
		for i := 0; i < 1000; i++ {
			regrowTick(w, rng, i)
		}
		tile := w.Tiles[0][0]
		if tile.Terrain != Grassland {
			t.Errorf("Terrain = %v, want Grassland", tile.Terrain)
		}
	})

	t.Run("cut tree within spawn no-grow zone converts to Grassland", func(t *testing.T) {
		rng := rand.New(rand.NewSource(0))
		// 20×20 world: spawn = (10,10). Tile at (10,10) is distance 0 ≤ 8.
		w := NewWorld(20, 20)
		w.Tiles[10][10] = Tile{Terrain: Forest, TreeSize: 0}
		regrowTick(w, rng, 0)
		if w.Tiles[10][10].Terrain != Grassland {
			t.Errorf("Terrain = %v, want Grassland (cut tree in no-grow zone should convert)", w.Tiles[10][10].Terrain)
		}
	})

	t.Run("living forest within spawn no-grow zone does not grow or convert", func(t *testing.T) {
		rng := rand.New(rand.NewSource(0))
		w := NewWorld(20, 20)
		w.Tiles[10][10] = Tile{Terrain: Forest, TreeSize: 5}
		for i := 0; i < 1000; i++ {
			regrowTick(w, rng, i)
		}
		tile := w.Tiles[10][10]
		if tile.Terrain != Forest {
			t.Errorf("Terrain = %v, want Forest (living tree should not convert)", tile.Terrain)
		}
		if tile.TreeSize != 5 {
			t.Errorf("TreeSize = %d, want 5 (living tree in no-grow zone should not grow)", tile.TreeSize)
		}
	})

	t.Run("cut tree within building no-grow zone converts to Grassland", func(t *testing.T) {
		rng := rand.New(rand.NewSource(0))
		// 40×40 world: spawn = (20,20). Place a structure at (5,5) and a Forest
		// tile at (5,10) — distance 5 ≤ 8 from the structure, and distance
		// sqrt(225+100)=~18 from spawn (safely outside the spawn zone).
		w := NewWorld(40, 40)
		w.SetStructure(5, 5, 1, 1, LogStorage)
		w.Tiles[10][5] = Tile{Terrain: Forest, TreeSize: 0}
		regrowTick(w, rng, 0)
		if w.Tiles[10][5].Terrain != Grassland {
			t.Errorf("Terrain = %v, want Grassland (cut tree in building no-grow zone should convert)", w.Tiles[10][5].Terrain)
		}
	})
}

func TestSetStructure(t *testing.T) {
	t.Run("stamps correct tiles", func(t *testing.T) {
		w := NewWorld(10, 10)
		w.SetStructure(2, 3, 4, 4, LogStorage)
		for dy := 0; dy < 4; dy++ {
			for dx := 0; dx < 4; dx++ {
				tile := w.TileAt(2+dx, 3+dy)
				if tile.Structure != LogStorage {
					t.Errorf("tile (%d,%d) Structure = %v, want LogStorage", 2+dx, 3+dy, tile.Structure)
				}
			}
		}
		// Outside footprint is unchanged.
		if w.TileAt(1, 3).Structure != NoStructure {
			t.Error("tile outside footprint should have NoStructure")
		}
	})

	t.Run("clips out-of-bounds tiles gracefully", func(t *testing.T) {
		w := NewWorld(5, 5)
		// Should not panic even if rect extends outside world.
		w.SetStructure(3, 3, 4, 4, FoundationLogStorage)
	})

	t.Run("populates NoGrowTiles within noGrowRadius", func(t *testing.T) {
		// 30×30 world. Place a 1×1 structure at (15,15).
		// Tile at (15,20) is distance 5 ≤ 8: must be in NoGrowTiles.
		// Tile at (15,24) is distance 9 > 8: must NOT be in NoGrowTiles.
		w := NewWorld(30, 30)
		w.SetStructure(15, 15, 1, 1, LogStorage)
		if _, ok := w.NoGrowTiles[Point{15, 20}]; !ok {
			t.Error("tile at distance 5 from structure should be in NoGrowTiles")
		}
		if _, ok := w.NoGrowTiles[Point{15, 24}]; ok {
			t.Error("tile at distance 9 from structure should not be in NoGrowTiles")
		}
	})

	t.Run("NoStructure stamp does not add to NoGrowTiles", func(t *testing.T) {
		w := NewWorld(30, 30)
		before := len(w.NoGrowTiles)
		w.SetStructure(1, 1, 1, 1, NoStructure) // outside spawn zone
		if len(w.NoGrowTiles) != before {
			t.Errorf("NoGrowTiles grew by %d entries after NoStructure stamp, want 0", len(w.NoGrowTiles)-before)
		}
	})
}

func TestSetStructureTypeIndex(t *testing.T) {
	t.Run("stamps populate type index", func(t *testing.T) {
		w := NewWorld(10, 10)
		w.SetStructure(2, 3, 2, 2, LogStorage)
		pts := w.StructureTypeIndex[LogStorage]
		if len(pts) != 4 {
			t.Fatalf("type index has %d points, want 4", len(pts))
		}
		for dy := 0; dy < 2; dy++ {
			for dx := 0; dx < 2; dx++ {
				p := Point{2 + dx, 3 + dy}
				if _, ok := pts[p]; !ok {
					t.Errorf("expected point %v in type index", p)
				}
			}
		}
	})

	t.Run("foundation to built transition removes foundation entries", func(t *testing.T) {
		w := NewWorld(10, 10)
		w.SetStructure(1, 1, 2, 2, FoundationLogStorage)
		if len(w.StructureTypeIndex[FoundationLogStorage]) != 4 {
			t.Fatalf("expected 4 foundation entries after placement")
		}
		w.SetStructure(1, 1, 2, 2, LogStorage)
		if len(w.StructureTypeIndex[FoundationLogStorage]) != 0 {
			t.Errorf("foundation entries should be gone after overwrite, got %d", len(w.StructureTypeIndex[FoundationLogStorage]))
		}
		if _, exists := w.StructureTypeIndex[FoundationLogStorage]; exists {
			t.Error("foundation key should be removed from index when empty")
		}
		if len(w.StructureTypeIndex[LogStorage]) != 4 {
			t.Errorf("expected 4 built entries, got %d", len(w.StructureTypeIndex[LogStorage]))
		}
	})

	t.Run("clearing tiles removes entries", func(t *testing.T) {
		w := NewWorld(10, 10)
		w.SetStructure(0, 0, 3, 3, LogStorage)
		w.SetStructure(0, 0, 3, 3, NoStructure)
		if len(w.StructureTypeIndex[LogStorage]) != 0 {
			t.Errorf("expected no entries after clear, got %d", len(w.StructureTypeIndex[LogStorage]))
		}
		if _, exists := w.StructureTypeIndex[LogStorage]; exists {
			t.Error("key should be removed from index when empty")
		}
	})
}

func TestIndexStructure(t *testing.T) {
	t.Run("single tile entry has correct origin", func(t *testing.T) {
		w := NewWorld(10, 10)
		w.IndexStructure(3, 4, 1, 1, logStorageDef{})
		entry, ok := w.StructureIndex[Point{3, 4}]
		if !ok {
			t.Fatal("expected entry at (3,4)")
		}
		if entry.Origin != (Point{3, 4}) {
			t.Errorf("Origin = %v, want {3,4}", entry.Origin)
		}
	})

	t.Run("4x4 footprint: all 16 tiles indexed with same origin", func(t *testing.T) {
		w := NewWorld(20, 20)
		w.IndexStructure(2, 3, 4, 4, logStorageDef{})
		origin := Point{2, 3}
		for dy := 0; dy < 4; dy++ {
			for dx := 0; dx < 4; dx++ {
				p := Point{2 + dx, 3 + dy}
				entry, ok := w.StructureIndex[p]
				if !ok {
					t.Errorf("missing entry at %v", p)
					continue
				}
				if entry.Origin != origin {
					t.Errorf("tile %v Origin = %v, want %v", p, entry.Origin, origin)
				}
			}
		}
	})

	t.Run("second call with same origin is idempotent", func(t *testing.T) {
		w := NewWorld(10, 10)
		w.IndexStructure(1, 1, 1, 1, logStorageDef{})
		w.IndexStructure(1, 1, 1, 1, logStorageDef{})
		if len(w.StructureIndex) != 1 {
			t.Errorf("index len = %d, want 1", len(w.StructureIndex))
		}
	})
}

func TestTileAt(t *testing.T) {
	w := NewWorld(10, 10)

	tile := w.TileAt(5, 5)
	if tile == nil {
		t.Fatal("TileAt(5,5) returned nil")
	}

	// Out of bounds returns nil.
	if w.TileAt(-1, 0) != nil {
		t.Error("TileAt(-1,0) should return nil")
	}
	if w.TileAt(10, 10) != nil {
		t.Error("TileAt(10,10) should return nil")
	}
}
