package game

import "testing"

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
	t.Run("stump regrows into small forest", func(t *testing.T) {
		w := NewWorld(3, 3)
		w.Tiles[1][1] = Tile{Terrain: Stump, TreeSize: 0}
		w.Regrow()
		tile := w.Tiles[1][1]
		if tile.Terrain != Forest {
			t.Errorf("Terrain = %v, want Forest", tile.Terrain)
		}
		if tile.TreeSize != 1 {
			t.Errorf("TreeSize = %d, want 1", tile.TreeSize)
		}
	})

	t.Run("forest grows toward maxTreeSize", func(t *testing.T) {
		w := NewWorld(3, 3)
		w.Tiles[1][1] = Tile{Terrain: Forest, TreeSize: 5}
		w.Regrow()
		if w.Tiles[1][1].TreeSize != 6 {
			t.Errorf("TreeSize = %d, want 6", w.Tiles[1][1].TreeSize)
		}
	})

	t.Run("forest at maxTreeSize does not grow further", func(t *testing.T) {
		w := NewWorld(3, 3)
		w.Tiles[1][1] = Tile{Terrain: Forest, TreeSize: maxTreeSize}
		w.Regrow()
		if w.Tiles[1][1].TreeSize != maxTreeSize {
			t.Errorf("TreeSize = %d, want %d", w.Tiles[1][1].TreeSize, maxTreeSize)
		}
	})

	t.Run("grassland is unaffected", func(t *testing.T) {
		w := NewWorld(3, 3)
		w.Tiles[1][1] = Tile{Terrain: Grassland}
		w.Regrow()
		tile := w.Tiles[1][1]
		if tile.Terrain != Grassland {
			t.Errorf("Terrain = %v, want Grassland", tile.Terrain)
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
