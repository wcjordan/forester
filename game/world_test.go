package game

import (
	"math/rand"
	"testing"
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

func TestRegrow(t *testing.T) {
	t.Run("cut tree eventually grows", func(t *testing.T) {
		rng := rand.New(rand.NewSource(0))
		w := NewWorld(3, 3)
		w.Tiles[1][1] = Tile{Terrain: Forest, TreeSize: 0}
		grew := false
		for i := 0; i < 1000; i++ {
			w.Regrow(rng)
			if w.Tiles[1][1].TreeSize > 0 {
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
		w := NewWorld(3, 3)
		w.Tiles[1][1] = Tile{Terrain: Forest, TreeSize: 5}
		grew := false
		for i := 0; i < 1000; i++ {
			w.Regrow(rng)
			if w.Tiles[1][1].TreeSize > 5 {
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
		w := NewWorld(3, 3)
		w.Tiles[1][1] = Tile{Terrain: Forest, TreeSize: maxTreeSize}
		for i := 0; i < 1000; i++ {
			w.Regrow(rng)
		}
		if w.Tiles[1][1].TreeSize != maxTreeSize {
			t.Errorf("TreeSize = %d, want %d", w.Tiles[1][1].TreeSize, maxTreeSize)
		}
	})

	t.Run("grassland is unaffected", func(t *testing.T) {
		rng := rand.New(rand.NewSource(0))
		w := NewWorld(3, 3)
		w.Tiles[1][1] = Tile{Terrain: Grassland}
		for i := 0; i < 1000; i++ {
			w.Regrow(rng)
		}
		tile := w.Tiles[1][1]
		if tile.Terrain != Grassland {
			t.Errorf("Terrain = %v, want Grassland", tile.Terrain)
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
		w.SetStructure(3, 3, 4, 4, GhostLogStorage)
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

func TestIsAdjacentToStructure(t *testing.T) {
	w := NewWorld(10, 10)
	w.SetStructure(5, 5, 1, 1, LogStorage)

	// Cardinal neighbors.
	for _, d := range [][2]int{{5, 4}, {5, 6}, {4, 5}, {6, 5}} {
		if !w.IsAdjacentToStructure(d[0], d[1], LogStorage) {
			t.Errorf("(%d,%d) should be adjacent to LogStorage", d[0], d[1])
		}
	}
	// Diagonal — not adjacent.
	if w.IsAdjacentToStructure(4, 4, LogStorage) {
		t.Error("(4,4) diagonal should not count as adjacent")
	}
	// The tile itself — not adjacent to itself via cardinal check.
	if w.IsAdjacentToStructure(5, 5, LogStorage) {
		t.Error("(5,5) should not be adjacent to itself")
	}
	// Far tile — false.
	if w.IsAdjacentToStructure(0, 0, LogStorage) {
		t.Error("(0,0) should not be adjacent to LogStorage at (5,5)")
	}
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
