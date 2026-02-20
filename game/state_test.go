package game

import "testing"

// makeStateWithForest creates a small test state with a clear grassland area
// and one forest tile in front of the player for harvesting.
func makeStateWithForest(playerX, playerY int) *State {
	w := NewWorld(20, 20)
	// Place forest in front of the player (player faces north by default)
	w.Tiles[playerY-1][playerX] = Tile{Terrain: Forest, TreeSize: 5}
	return &State{Player: NewPlayer(playerX, playerY), World: w}
}

func TestHarvestTracksTotalWoodCut(t *testing.T) {
	s := makeStateWithForest(5, 5)
	s.Harvest()
	if s.TotalWoodCut != 1 {
		t.Errorf("TotalWoodCut = %d, want 1", s.TotalWoodCut)
	}
	s.Harvest()
	if s.TotalWoodCut != 2 {
		t.Errorf("TotalWoodCut = %d, want 2", s.TotalWoodCut)
	}
}

func TestGhostSpawnsAfter10WoodCut(t *testing.T) {
	// Build a world big enough that there's a clear path from player to center.
	w := NewWorld(30, 30)
	// Player at (5, 5) facing north; forest tile at (5, 4) with enough wood.
	w.Tiles[4][5] = Tile{Terrain: Forest, TreeSize: 20}
	p := NewPlayer(5, 5)
	s := &State{Player: p, World: w}

	// Harvest 9 times — ghost should not appear yet.
	for range 9 {
		s.Harvest()
	}
	if s.HasStructureOfType(GhostLogStorage) {
		t.Fatal("ghost appeared before 10 wood cut")
	}

	// 10th harvest — ghost should now appear.
	s.Harvest()
	if !s.HasStructureOfType(GhostLogStorage) {
		t.Error("ghost did not appear after 10 wood cut")
	}
}

func TestGhostDoesNotSpawnTwice(t *testing.T) {
	w := NewWorld(30, 30)
	for i := 0; i < 15; i++ {
		w.Tiles[4][5+i] = Tile{Terrain: Forest, TreeSize: 1}
	}
	p := NewPlayer(5, 5)
	s := &State{Player: p, World: w, TotalWoodCut: 10}

	s.maybeSpawnGhost()
	// Count ghost tiles.
	count := 0
	for y := range w.Tiles {
		for x := range w.Tiles[y] {
			if w.Tiles[y][x].Structure == GhostLogStorage {
				count++
			}
		}
	}
	firstCount := count

	// Call again — should not add more.
	s.maybeSpawnGhost()
	count = 0
	for y := range w.Tiles {
		for x := range w.Tiles[y] {
			if w.Tiles[y][x].Structure == GhostLogStorage {
				count++
			}
		}
	}
	if count != firstCount {
		t.Errorf("ghost tile count changed from %d to %d on second spawn attempt", firstCount, count)
	}
}

func TestGhostLocationIsAllGrassland(t *testing.T) {
	w := NewWorld(30, 30)
	p := NewPlayer(5, 15)
	s := &State{Player: p, World: w, TotalWoodCut: 10}
	s.maybeSpawnGhost()

	// Find the ghost and verify all 16 tiles are on grassland terrain (underlying).
	for y := range w.Tiles {
		for x := range w.Tiles[y] {
			if w.Tiles[y][x].Structure == GhostLogStorage {
				// This tile is part of the ghost footprint — check original terrain.
				// Since we built on grassland, terrain should still be Grassland.
				if w.Tiles[y][x].Terrain != Grassland {
					t.Errorf("ghost tile (%d,%d) is on non-grassland terrain", x, y)
				}
			}
		}
	}
}

func TestGhostLocationBetweenPlayerAndSpawn(t *testing.T) {
	w := NewWorld(30, 30)
	// Player at (2, 15); spawn at (15, 15).
	p := NewPlayer(2, 15)
	s := &State{Player: p, World: w, TotalWoodCut: 10}
	s.maybeSpawnGhost()

	spawnX := w.Width / 2
	// Find ghost top-left.
	gx, gy := -1, -1
	for y := range w.Tiles {
		for x := range w.Tiles[y] {
			if w.Tiles[y][x].Structure == GhostLogStorage && gx == -1 {
				gx, gy = x, y
			}
		}
	}
	if gx == -1 {
		t.Fatal("no ghost placed")
	}
	_ = gy
	// Ghost x-coordinate should be between player and spawn center.
	if gx < p.X || gx > spawnX {
		t.Errorf("ghost x=%d not between player x=%d and spawn x=%d", gx, p.X, spawnX)
	}
}

func TestHasStructureOfType(t *testing.T) {
	w := NewWorld(10, 10)
	s := &State{Player: NewPlayer(5, 5), World: w}

	if s.HasStructureOfType(LogStorage) {
		t.Error("should have no LogStorage initially")
	}
	w.SetStructure(1, 1, 2, 2, LogStorage)
	if !s.HasStructureOfType(LogStorage) {
		t.Error("should detect LogStorage after SetStructure")
	}
	if s.HasStructureOfType(GhostLogStorage) {
		t.Error("should not detect GhostLogStorage when none placed")
	}
}
