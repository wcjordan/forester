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

func TestAddStructure(t *testing.T) {
	def := testLogStorageDef{} // 4×4, BuiltType=LogStorage, FoundationType=FoundationLogStorage

	t.Run("stamps correct tiles", func(t *testing.T) {
		w := NewWorld(10, 10)
		w.PlaceBuilt(2, 3, def)
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
		w.PlaceFoundation(3, 3, def)
	})

	t.Run("populates NoGrowTiles within noGrowRadius", func(t *testing.T) {
		// 30×30 world. Place a 4×4 structure at (15,15); footprint covers (15,15)–(18,18).
		// Nearest footprint point to (15,20) is (15,18): distance 2 ≤ 8 → must be in NoGrowTiles.
		// Nearest footprint point to (15,27) is (15,18): distance 9 > 8 → must NOT be in NoGrowTiles.
		w := NewWorld(30, 30)
		w.PlaceBuilt(15, 15, def)
		if _, ok := w.NoGrowTiles[point{X: 15, Y: 20}]; !ok {
			t.Error("tile at distance 2 from structure should be in NoGrowTiles")
		}
		if _, ok := w.NoGrowTiles[point{X: 15, Y: 27}]; ok {
			t.Error("tile at distance 9 from structure should not be in NoGrowTiles")
		}
	})

	t.Run("clearing a structure does not add to NoGrowTiles", func(t *testing.T) {
		w := NewWorld(30, 30)
		w.PlaceBuilt(1, 1, def)
		after := len(w.NoGrowTiles)
		w.clearStructure(1, 1, def)
		if len(w.NoGrowTiles) != after {
			t.Errorf("NoGrowTiles grew by %d entries after clearing, want 0", len(w.NoGrowTiles)-after)
		}
	})

	t.Run("stamps populate type index", func(t *testing.T) {
		w := NewWorld(10, 10)
		w.PlaceBuilt(2, 3, def)
		pts := w.StructureTypeIndex[LogStorage]
		if len(pts) != 16 {
			t.Fatalf("type index has %d points, want 16 (4×4)", len(pts))
		}
		for dy := 0; dy < 4; dy++ {
			for dx := 0; dx < 4; dx++ {
				p := point{X: 2 + dx, Y: 3 + dy}
				if _, ok := pts[p]; !ok {
					t.Errorf("expected point %v in type index", p)
				}
			}
		}
	})

	t.Run("foundation to built transition removes foundation entries", func(t *testing.T) {
		w := NewWorld(10, 10)
		w.PlaceFoundation(1, 1, def)
		if len(w.StructureTypeIndex[FoundationLogStorage]) != 16 {
			t.Fatalf("expected 16 foundation entries after placement (4×4)")
		}
		w.PlaceBuilt(1, 1, def)
		if len(w.StructureTypeIndex[FoundationLogStorage]) != 0 {
			t.Errorf("foundation entries should be gone after overwrite, got %d", len(w.StructureTypeIndex[FoundationLogStorage]))
		}
		if _, exists := w.StructureTypeIndex[FoundationLogStorage]; exists {
			t.Error("foundation key should be removed from index when empty")
		}
		if len(w.StructureTypeIndex[LogStorage]) != 16 {
			t.Errorf("expected 16 built entries (4×4), got %d", len(w.StructureTypeIndex[LogStorage]))
		}
	})

	t.Run("clearing tiles removes type index entries", func(t *testing.T) {
		w := NewWorld(10, 10)
		w.PlaceBuilt(0, 0, def)
		w.clearStructure(0, 0, def)
		if len(w.StructureTypeIndex[LogStorage]) != 0 {
			t.Errorf("expected no entries after clear, got %d", len(w.StructureTypeIndex[LogStorage]))
		}
		if _, exists := w.StructureTypeIndex[LogStorage]; exists {
			t.Error("key should be removed from index when empty")
		}
	})

	t.Run("origin tile has correct structureIndex entry", func(t *testing.T) {
		w := NewWorld(10, 10)
		w.PlaceBuilt(3, 4, def)
		entry, ok := w.structureIndex[point{X: 3, Y: 4}]
		if !ok {
			t.Fatal("expected entry at (3,4)")
		}
		if entry.Origin != (point{X: 3, Y: 4}) {
			t.Errorf("Origin = %v, want {3,4}", entry.Origin)
		}
	})

	t.Run("all footprint tiles indexed with same origin", func(t *testing.T) {
		w := NewWorld(20, 20)
		w.PlaceBuilt(2, 3, def)
		origin := point{X: 2, Y: 3}
		for dy := 0; dy < 4; dy++ {
			for dx := 0; dx < 4; dx++ {
				p := point{X: 2 + dx, Y: 3 + dy}
				entry, ok := w.structureIndex[p]
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
		fw, fh := def.Footprint()
		w.PlaceBuilt(1, 1, def)
		w.PlaceBuilt(1, 1, def)
		if got := len(w.structureIndex); got != fw*fh {
			t.Errorf("structureIndex len = %d, want %d (one entry per tile, not duplicated)", got, fw*fh)
		}
		if w.CountStructureInstances(LogStorage) != 1 {
			t.Errorf("CountStructureInstances = %d, want 1 (idempotent)", w.CountStructureInstances(LogStorage))
		}
	})
}

func TestHasStructureOfType(t *testing.T) {
	w := NewWorld(10, 10)

	if w.HasStructureOfType(LogStorage) {
		t.Error("should have no LogStorage initially")
	}
	w.PlaceBuilt(1, 1, testLogStorageDef{})
	if !w.HasStructureOfType(LogStorage) {
		t.Error("should detect LogStorage after PlaceBuilt")
	}
	if w.HasStructureOfType(FoundationLogStorage) {
		t.Error("should not detect FoundationLogStorage when none placed")
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
