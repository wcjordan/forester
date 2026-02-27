package game

import "testing"

// TestFindPath_DirectPath verifies a straight-line path on an open grassland world.
func TestFindPath_DirectPath(t *testing.T) {
	w := NewWorld(20, 20)
	path := findPath(w, 0, 0, 5, 0)
	if path == nil {
		t.Fatal("findPath returned nil, want a valid path")
	}
	last := path[len(path)-1]
	if last != (Point{5, 0}) {
		t.Errorf("last point = %v, want {5,0}", last)
	}
	// 5 steps in X direction.
	if len(path) != 5 {
		t.Errorf("path length = %d, want 5", len(path))
	}
}

// TestFindPath_RouteAroundWall verifies the path routes around a blocking structure wall.
func TestFindPath_RouteAroundWall(t *testing.T) {
	w := NewWorld(20, 20)
	// Vertical wall at X=5, Y=0..14 (width=1, height=15).
	w.SetStructure(5, 0, 1, 15, LogStorage)

	// Villager at (2,7), goal at (10,7). Direct route blocked by wall.
	path := findPath(w, 2, 7, 10, 7)
	if path == nil {
		t.Fatal("findPath returned nil, want a path around the wall")
	}
	last := path[len(path)-1]
	if last != (Point{10, 7}) {
		t.Errorf("goal = %v, want {10,7}", last)
	}
	// No path point should be inside the wall (X=5, Y=0..14).
	for _, p := range path {
		if p.X == 5 && p.Y < 15 {
			t.Errorf("path passes through wall tile at %v", p)
		}
	}
}

// TestFindPath_PrefersGrassOverForest verifies the path avoids costly forest tiles when
// an equally short grass detour exists.
func TestFindPath_PrefersGrassOverForest(t *testing.T) {
	// 10×10 world, villager at (0,5), goal at (6,5).
	// Three forest tiles block the direct row: (2,5),(3,5),(4,5).
	// Clear rows above (y=4) and below (y=6) let A* find a lower-cost detour.
	w := NewWorld(10, 10)
	w.Tiles[5][2] = Tile{Terrain: Forest, TreeSize: 5}
	w.Tiles[5][3] = Tile{Terrain: Forest, TreeSize: 5}
	w.Tiles[5][4] = Tile{Terrain: Forest, TreeSize: 5}

	path := findPath(w, 0, 5, 6, 5)
	if path == nil {
		t.Fatal("findPath returned nil")
	}
	// Count forest tiles traversed.
	forestHits := 0
	for _, p := range path {
		tile := w.TileAt(p.X, p.Y)
		if tile != nil && tile.Terrain == Forest && tile.TreeSize > 0 {
			forestHits++
		}
	}
	// A detour of 1 step above/below the 3-forest strip costs 6+2 = 8
	// vs. going straight through all 3 forests: 6 + 3*2 - 3 = 9
	// (effectively direct path costs 6 + extra 3 from forest = 9 vs detour 8).
	// The optimal path avoids all three forest tiles.
	if forestHits > 0 {
		t.Errorf("path traversed %d forest tiles, want 0 (detour should be cheaper)", forestHits)
	}
}

// TestFindPath_Unreachable verifies nil is returned when the goal is enclosed.
func TestFindPath_Unreachable(t *testing.T) {
	w := NewWorld(20, 20)
	// Full vertical wall at X=10, blocking the entire height.
	w.SetStructure(10, 0, 1, 20, LogStorage)

	path := findPath(w, 5, 5, 15, 5)
	if path != nil {
		t.Errorf("findPath returned non-nil path through impassable wall")
	}
}

// TestFindPath_StartEqualsGoal verifies an empty path is returned when start == goal.
func TestFindPath_StartEqualsGoal(t *testing.T) {
	w := NewWorld(10, 10)
	path := findPath(w, 5, 5, 5, 5)
	if path == nil {
		t.Fatal("findPath returned nil for start==goal, want empty slice")
	}
	if len(path) != 0 {
		t.Errorf("path length = %d, want 0 for start==goal", len(path))
	}
}
