package game

import "testing"

func TestGenerateWorld_Dimensions(t *testing.T) {
	w := GenerateWorld(50, 30, 42)

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

func TestGenerateWorld_Deterministic(t *testing.T) {
	w1 := GenerateWorld(50, 50, 42)
	w2 := GenerateWorld(50, 50, 42)

	for y := range w1.Tiles {
		for x := range w1.Tiles[y] {
			if w1.Tiles[y][x].Terrain != w2.Tiles[y][x].Terrain {
				t.Errorf("tile (%d,%d) differs between same-seed runs", x, y)
			}
		}
	}
}

func TestGenerateWorld_DifferentSeeds(t *testing.T) {
	w1 := GenerateWorld(50, 50, 42)
	w2 := GenerateWorld(50, 50, 99)

	different := false
outer:
	for y := range w1.Tiles {
		for x := range w1.Tiles[y] {
			if w1.Tiles[y][x].Terrain != w2.Tiles[y][x].Terrain {
				different = true
				break outer
			}
		}
	}

	if !different {
		t.Error("different seeds produced identical maps")
	}
}

func TestGenerateWorld_ForestDensity(t *testing.T) {
	w := GenerateWorld(100, 100, 42)

	forest := 0
	total := w.Width * w.Height
	for y := range w.Tiles {
		for x := range w.Tiles[y] {
			if w.Tiles[y][x].Terrain == Forest {
				forest++
			}
		}
	}

	pct := float64(forest) / float64(total)
	if pct < 0.40 || pct > 0.80 {
		t.Errorf("forest density = %.2f, want between 0.40 and 0.80", pct)
	}
}

func TestGenerateWorld_TreeSizes(t *testing.T) {
	w := GenerateWorld(50, 50, 42)

	for y := range w.Tiles {
		for x := range w.Tiles[y] {
			tile := &w.Tiles[y][x]
			if tile.Terrain == Forest {
				if tile.TreeSize < 4 || tile.TreeSize > 10 {
					t.Errorf("Forest tile (%d,%d) has TreeSize=%d, want 4–10", x, y, tile.TreeSize)
				}
			} else {
				if tile.TreeSize != 0 {
					t.Errorf("Non-Forest tile (%d,%d) has TreeSize=%d, want 0", x, y, tile.TreeSize)
				}
			}
		}
	}
}

func TestGenerateWorld_SpawnClear(t *testing.T) {
	w := GenerateWorld(100, 100, 42)
	cx, cy := w.Width/2, w.Height/2

	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			tile := w.TileAt(cx+dx, cy+dy)
			if tile == nil {
				t.Errorf("tile (%d,%d) is out of bounds", cx+dx, cy+dy)
				continue
			}
			if tile.Terrain != Grassland {
				t.Errorf("spawn tile (%d,%d) = Forest, want Grassland", cx+dx, cy+dy)
			}
			if tile.TreeSize != 0 {
				t.Errorf("spawn tile (%d,%d) has TreeSize=%d, want 0", cx+dx, cy+dy, tile.TreeSize)
			}
		}
	}
}
