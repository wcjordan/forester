package game

import (
	"testing"
	"time"
)

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

	s.maybeSpawnGhosts()
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
	s.maybeSpawnGhosts()
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
	s.maybeSpawnGhosts()

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
	s.maybeSpawnGhosts()

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

func TestBuildMechanic(t *testing.T) {
	// Set up a world with a ghost at (5, 5) and player adjacent.
	makeGhostState := func() *State {
		w := NewWorld(20, 20)
		w.SetStructure(5, 5, 4, 4, GhostLogStorage)
		p := NewPlayer(4, 5) // just outside the ghost footprint
		return &State{Player: p, World: w}
	}

	t.Run("walking onto ghost tile starts build", func(t *testing.T) {
		s := makeGhostState()
		s.Move(1, 0) // step into (5,5) — ghost tile
		if s.Building == nil {
			t.Fatal("Building should be non-nil after stepping onto ghost")
		}
		if s.Building.TotalTicks != LogStorageBuildTicks {
			t.Errorf("TotalTicks = %d, want %d", s.Building.TotalTicks, LogStorageBuildTicks)
		}
	})

	t.Run("player nudged outside footprint after ghost contact", func(t *testing.T) {
		s := makeGhostState()
		s.Move(1, 0) // step into (5,5)
		px, py := s.Player.X, s.Player.Y
		// Player must be outside the 4×4 footprint [5..8] x [5..8].
		insideX := px >= 5 && px <= 8
		insideY := py >= 5 && py <= 8
		if insideX && insideY {
			t.Errorf("player at (%d,%d) is still inside ghost footprint [5-8,5-8]", px, py)
		}
	})

	t.Run("AdvanceBuild increments progress", func(t *testing.T) {
		s := makeGhostState()
		s.Move(1, 0)
		if s.Building == nil {
			t.Fatal("Building is nil")
		}
		s.AdvanceBuild()
		if s.Building.ProgressTicks != 1 {
			t.Errorf("ProgressTicks = %d, want 1", s.Building.ProgressTicks)
		}
	})

	t.Run("build completes and tiles become LogStorage", func(t *testing.T) {
		s := makeGhostState()
		s.Move(1, 0)
		if s.Building == nil {
			t.Fatal("Building is nil")
		}
		s.Building.ProgressTicks = s.Building.TotalTicks - 1
		s.AdvanceBuild()
		if s.Building != nil {
			t.Error("Building should be nil after completion")
		}
		if !s.HasStructureOfType(LogStorage) {
			t.Error("LogStorage tiles should exist after build completes")
		}
	})

	t.Run("ghost tiles replaced by LogStorage after build", func(t *testing.T) {
		s := makeGhostState()
		s.Move(1, 0)
		s.Building.ProgressTicks = s.Building.TotalTicks - 1
		s.AdvanceBuild()
		if s.HasStructureOfType(GhostLogStorage) {
			t.Error("GhostLogStorage tiles should be gone after build completes")
		}
	})
}

func TestTickAdjacentStructures(t *testing.T) {
	makeDepositState := func(wood int) *State {
		w := NewWorld(10, 10)
		w.SetStructure(5, 4, 4, 4, LogStorage) // storage above player
		w.IndexStructure(5, 4, 4, 4, logStorageDef{})
		p := NewPlayer(5, 5)
		p.Wood = wood
		s := &State{Player: p, World: w, Storage: make(map[ResourceType]*ResourceStorage)}
		s.getStorage(Wood).AddInstance(Wood, LogStorageCapacity)
		return s
	}

	t.Run("no deposit when Wood is 0", func(t *testing.T) {
		s := makeDepositState(0)
		s.TickAdjacentStructures(time.Now())
		if s.TotalStored(Wood) != 0 {
			t.Errorf("TotalStored(Wood) = %d, want 0 when Wood == 0", s.TotalStored(Wood))
		}
	})

	t.Run("no deposit when not adjacent to LogStorage", func(t *testing.T) {
		w := NewWorld(10, 10)
		p := NewPlayer(5, 5)
		p.Wood = 5
		s := &State{Player: p, World: w, Storage: make(map[ResourceType]*ResourceStorage)}
		s.TickAdjacentStructures(time.Now())
		if s.TotalStored(Wood) != 0 {
			t.Errorf("TotalStored(Wood) = %d, want 0 when not adjacent", s.TotalStored(Wood))
		}
	})

	t.Run("deposits one wood when adjacent", func(t *testing.T) {
		s := makeDepositState(5)
		s.TickAdjacentStructures(time.Now())
		if s.Player.Wood != 4 {
			t.Errorf("Wood = %d, want 4 after deposit", s.Player.Wood)
		}
		if s.TotalStored(Wood) != 1 {
			t.Errorf("TotalStored(Wood) = %d, want 1", s.TotalStored(Wood))
		}
	})

	t.Run("deposits one at a time with cooldown between ticks", func(t *testing.T) {
		s := makeDepositState(3)
		t0 := time.Now()
		s.TickAdjacentStructures(t0)
		s.TickAdjacentStructures(t0.Add(DepositTickInterval + time.Millisecond))
		if s.Player.Wood != 1 {
			t.Errorf("Wood = %d, want 1 after 2 deposits", s.Player.Wood)
		}
		if s.TotalStored(Wood) != 2 {
			t.Errorf("TotalStored(Wood) = %d, want 2", s.TotalStored(Wood))
		}
	})

	t.Run("two adjacent instances each trigger an interaction", func(t *testing.T) {
		// Player at (5,5). LogStorage A above (y=4), LogStorage B below (y=6).
		// Cooldowns are queued during interactions and committed after all are
		// processed, so both instances fire within the same tick.
		w := NewWorld(20, 20)
		w.SetStructure(5, 4, 1, 1, LogStorage)
		w.IndexStructure(5, 4, 1, 1, logStorageDef{})
		w.SetStructure(5, 6, 1, 1, LogStorage)
		w.IndexStructure(5, 6, 1, 1, logStorageDef{})
		p := NewPlayer(5, 5)
		p.Wood = 5
		s := &State{Player: p, World: w, Storage: make(map[ResourceType]*ResourceStorage)}
		s.getStorage(Wood).AddInstance(Wood, LogStorageCapacity)
		s.TickAdjacentStructures(time.Now())
		// Two instances adjacent → two deposits.
		if s.Player.Wood != 3 {
			t.Errorf("Wood = %d, want 3 after two-instance deposit", s.Player.Wood)
		}
		if s.TotalStored(Wood) != 2 {
			t.Errorf("TotalStored(Wood) = %d, want 2", s.TotalStored(Wood))
		}
	})
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
