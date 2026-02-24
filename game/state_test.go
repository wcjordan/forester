package game

import (
	"testing"
	"time"
)

func TestFoundationSpawnsWhenInventoryFull(t *testing.T) {
	// Build a world big enough that there's a clear path from player to center.
	w := NewWorld(30, 30)
	// Player at (5, 5) facing north; forest tile at (5, 4) with enough wood.
	w.Tiles[4][5] = Tile{Terrain: Forest, TreeSize: InitialCarryingCapacity}
	p := NewPlayer(5, 5)
	s := &State{Player: p, World: w, FoundationDeposited: make(map[Point]int)}
	stores := NewStorageManager()
	env := &Env{State: s, Stores: stores}

	// Harvest InitialCarryingCapacity-1 times — foundation should not appear yet.
	for range InitialCarryingCapacity - 1 {
		s.Harvest(env)
	}
	if s.HasStructureOfType(FoundationLogStorage) {
		t.Fatal("foundation appeared before inventory full")
	}

	// Final harvest fills inventory — foundation should now appear.
	s.Harvest(env)
	if !s.HasStructureOfType(FoundationLogStorage) {
		t.Error("foundation did not appear when inventory became full")
	}
}

func TestFoundationDoesNotSpawnTwice(t *testing.T) {
	w := NewWorld(30, 30)
	for i := 0; i < 15; i++ {
		w.Tiles[4][5+i] = Tile{Terrain: Forest, TreeSize: 1}
	}
	p := NewPlayer(5, 5)
	p.Wood = InitialCarryingCapacity
	s := &State{Player: p, World: w, FoundationDeposited: make(map[Point]int)}
	stores := NewStorageManager()
	env := &Env{State: s, Stores: stores}

	s.maybeSpawnFoundation(env)
	// Count foundation tiles.
	count := 0
	for y := range w.Tiles {
		for x := range w.Tiles[y] {
			if w.Tiles[y][x].Structure == FoundationLogStorage {
				count++
			}
		}
	}
	firstCount := count

	// Call again — should not add more.
	s.maybeSpawnFoundation(env)
	count = 0
	for y := range w.Tiles {
		for x := range w.Tiles[y] {
			if w.Tiles[y][x].Structure == FoundationLogStorage {
				count++
			}
		}
	}
	if count != firstCount {
		t.Errorf("foundation tile count changed from %d to %d on second spawn attempt", firstCount, count)
	}
}

func TestFoundationLocationIsAllGrassland(t *testing.T) {
	w := NewWorld(30, 30)
	p := NewPlayer(5, 15)
	p.Wood = InitialCarryingCapacity
	s := &State{Player: p, World: w, FoundationDeposited: make(map[Point]int)}
	stores := NewStorageManager()
	env := &Env{State: s, Stores: stores}
	s.maybeSpawnFoundation(env)

	// Find the foundation and verify all 16 tiles are on grassland terrain (underlying).
	for y := range w.Tiles {
		for x := range w.Tiles[y] {
			if w.Tiles[y][x].Structure == FoundationLogStorage {
				// This tile is part of the foundation footprint — check original terrain.
				// Since we built on grassland, terrain should still be Grassland.
				if w.Tiles[y][x].Terrain != Grassland {
					t.Errorf("foundation tile (%d,%d) is on non-grassland terrain", x, y)
				}
			}
		}
	}
}

func TestFoundationLocationBetweenPlayerAndSpawn(t *testing.T) {
	w := NewWorld(30, 30)
	// Player at (2, 15); spawn at (15, 15).
	p := NewPlayer(2, 15)
	p.Wood = InitialCarryingCapacity
	s := &State{Player: p, World: w, FoundationDeposited: make(map[Point]int)}
	stores := NewStorageManager()
	env := &Env{State: s, Stores: stores}
	s.maybeSpawnFoundation(env)

	spawnX := w.Width / 2
	// Find foundation top-left.
	gx, gy := -1, -1
	for y := range w.Tiles {
		for x := range w.Tiles[y] {
			if w.Tiles[y][x].Structure == FoundationLogStorage && gx == -1 {
				gx, gy = x, y
			}
		}
	}
	if gx == -1 {
		t.Fatal("no foundation placed")
	}
	_ = gy
	// Foundation x-coordinate should be between player and spawn center.
	if gx < p.X || gx > spawnX {
		t.Errorf("foundation x=%d not between player x=%d and spawn x=%d", gx, p.X, spawnX)
	}
}

func TestFoundationBuildMechanic(t *testing.T) {
	// Set up a world with an indexed foundation at (5,5) and player just west at (4,5).
	makeFoundationState := func(wood int) (*State, *StorageManager) {
		w := NewWorld(20, 20)
		w.SetStructure(5, 5, 4, 4, FoundationLogStorage)
		w.IndexStructure(5, 5, 4, 4, logStorageDef{})
		p := NewPlayer(4, 5)
		p.Wood = wood
		s := &State{
			Player:              p,
			World:               w,
			FoundationDeposited: make(map[Point]int),
		}
		return s, NewStorageManager()
	}

	t.Run("foundation blocks player movement", func(t *testing.T) {
		s, _ := makeFoundationState(0)
		s.Player.Move(1, 0, s.World, time.Now()) // try to step into (5,5) — foundation tile
		if s.Player.X != 4 {
			t.Errorf("player X = %d, want 4 (foundation should block movement)", s.Player.X)
		}
	})

	t.Run("adjacent deposit reduces player wood", func(t *testing.T) {
		s, stores := makeFoundationState(5)
		env := &Env{State: s, Stores: stores}
		s.TickAdjacentStructures(env, time.Now())
		if s.Player.Wood != 4 {
			t.Errorf("Wood = %d, want 4 after one deposit", s.Player.Wood)
		}
		origin := Point{5, 5}
		if s.FoundationDeposited[origin] != 1 {
			t.Errorf("FoundationDeposited = %d, want 1", s.FoundationDeposited[origin])
		}
	})

	t.Run("deposit respects cooldown", func(t *testing.T) {
		s, stores := makeFoundationState(5)
		env := &Env{State: s, Stores: stores}
		t0 := time.Now()
		s.TickAdjacentStructures(env, t0)
		s.TickAdjacentStructures(env, t0) // same timestamp — cooldown blocks
		origin := Point{5, 5}
		if s.FoundationDeposited[origin] != 1 {
			t.Errorf("FoundationDeposited = %d, want 1 (second tick should be blocked by cooldown)", s.FoundationDeposited[origin])
		}
	})

	t.Run("foundation completes after BuildCost deposits", func(t *testing.T) {
		s, stores := makeFoundationState(LogStorageBuildCost)
		env := &Env{State: s, Stores: stores}
		t0 := time.Now()
		for i := range LogStorageBuildCost {
			s.TickAdjacentStructures(env, t0.Add(time.Duration(i)*(DepositTickInterval+time.Millisecond)))
		}
		if s.HasStructureOfType(FoundationLogStorage) {
			t.Error("FoundationLogStorage tiles should be gone after build completes")
		}
		if !s.HasStructureOfType(LogStorage) {
			t.Error("LogStorage tiles should exist after build completes")
		}
		if s.Player.Wood != 0 {
			t.Errorf("player Wood = %d, want 0 (all wood deposited)", s.Player.Wood)
		}
	})
}

func TestTickAdjacentStructures(t *testing.T) {
	// makeDepositState creates a state with one LogStorage at origin (5,4) above the player.
	makeDepositState := func(wood int) (*State, *StorageManager) {
		w := NewWorld(10, 10)
		origin := Point{5, 4}
		w.SetStructure(origin.X, origin.Y, 4, 4, LogStorage)
		w.IndexStructure(origin.X, origin.Y, 4, 4, logStorageDef{})
		p := NewPlayer(5, 5)
		p.Wood = wood
		s := &State{Player: p, World: w, FoundationDeposited: make(map[Point]int)}
		stores := NewStorageManager()
		stores.Register(origin, Wood, LogStorageCapacity)
		return s, stores
	}

	t.Run("no deposit when Wood is 0", func(t *testing.T) {
		s, stores := makeDepositState(0)
		env := &Env{State: s, Stores: stores}
		s.TickAdjacentStructures(env, time.Now())
		if stores.Total(Wood) != 0 {
			t.Errorf("Total(Wood) = %d, want 0 when Wood == 0", stores.Total(Wood))
		}
	})

	t.Run("no deposit when not adjacent to LogStorage", func(t *testing.T) {
		w := NewWorld(10, 10)
		p := NewPlayer(5, 5)
		p.Wood = 5
		s := &State{Player: p, World: w, FoundationDeposited: make(map[Point]int)}
		stores := NewStorageManager()
		env := &Env{State: s, Stores: stores}
		s.TickAdjacentStructures(env, time.Now())
		if stores.Total(Wood) != 0 {
			t.Errorf("Total(Wood) = %d, want 0 when not adjacent", stores.Total(Wood))
		}
	})

	t.Run("deposits one wood when adjacent", func(t *testing.T) {
		s, stores := makeDepositState(5)
		env := &Env{State: s, Stores: stores}
		s.TickAdjacentStructures(env, time.Now())
		if s.Player.Wood != 4 {
			t.Errorf("Wood = %d, want 4 after deposit", s.Player.Wood)
		}
		if stores.Total(Wood) != 1 {
			t.Errorf("Total(Wood) = %d, want 1", stores.Total(Wood))
		}
	})

	t.Run("deposits one at a time with cooldown between ticks", func(t *testing.T) {
		s, stores := makeDepositState(3)
		env := &Env{State: s, Stores: stores}
		t0 := time.Now()
		s.TickAdjacentStructures(env, t0)
		s.TickAdjacentStructures(env, t0.Add(DepositTickInterval+time.Millisecond))
		if s.Player.Wood != 1 {
			t.Errorf("Wood = %d, want 1 after 2 deposits", s.Player.Wood)
		}
		if stores.Total(Wood) != 2 {
			t.Errorf("Total(Wood) = %d, want 2", stores.Total(Wood))
		}
	})

	t.Run("two adjacent instances each trigger an interaction", func(t *testing.T) {
		// Player at (5,5). LogStorage A above (y=4), LogStorage B below (y=6).
		// Cooldowns are queued during interactions and committed after all are
		// processed, so both instances fire within the same tick.
		// Each instance has its own StorageByOrigin entry so deposits route correctly.
		w := NewWorld(20, 20)
		originA := Point{5, 4}
		originB := Point{5, 6}
		w.SetStructure(originA.X, originA.Y, 1, 1, LogStorage)
		w.IndexStructure(originA.X, originA.Y, 1, 1, logStorageDef{})
		w.SetStructure(originB.X, originB.Y, 1, 1, LogStorage)
		w.IndexStructure(originB.X, originB.Y, 1, 1, logStorageDef{})
		p := NewPlayer(5, 5)
		p.Wood = 5
		s := &State{Player: p, World: w, FoundationDeposited: make(map[Point]int)}
		stores := NewStorageManager()
		stores.Register(originA, Wood, LogStorageCapacity)
		stores.Register(originB, Wood, LogStorageCapacity)
		instA := stores.FindByOrigin(originA)
		instB := stores.FindByOrigin(originB)
		env := &Env{State: s, Stores: stores}
		s.TickAdjacentStructures(env, time.Now())
		// Two instances adjacent → two deposits, one per instance.
		if s.Player.Wood != 3 {
			t.Errorf("Wood = %d, want 3 after two-instance deposit", s.Player.Wood)
		}
		if stores.Total(Wood) != 2 {
			t.Errorf("Total(Wood) = %d, want 2", stores.Total(Wood))
		}
		if instA.Stored != 1 {
			t.Errorf("instA.Stored = %d, want 1", instA.Stored)
		}
		if instB.Stored != 1 {
			t.Errorf("instB.Stored = %d, want 1", instB.Stored)
		}
	})
}

func TestDepositRoutesToSpecificInstance(t *testing.T) {
	// Two storages: A at (2,4) and B at (8,4). Player at (2,5) — adjacent to A only.
	// Deposit should go into A, not B.
	w := NewWorld(15, 10)
	originA := Point{2, 4}
	originB := Point{8, 4}
	w.SetStructure(originA.X, originA.Y, 1, 1, LogStorage)
	w.IndexStructure(originA.X, originA.Y, 1, 1, logStorageDef{})
	w.SetStructure(originB.X, originB.Y, 1, 1, LogStorage)
	w.IndexStructure(originB.X, originB.Y, 1, 1, logStorageDef{})
	p := NewPlayer(2, 5)
	p.Wood = 3
	s := &State{Player: p, World: w, FoundationDeposited: make(map[Point]int)}
	stores := NewStorageManager()
	stores.Register(originA, Wood, LogStorageCapacity)
	stores.Register(originB, Wood, LogStorageCapacity)
	instA := stores.FindByOrigin(originA)
	instB := stores.FindByOrigin(originB)
	env := &Env{State: s, Stores: stores}

	s.TickAdjacentStructures(env, time.Now())

	if instA.Stored != 1 {
		t.Errorf("instA.Stored = %d, want 1 (adjacent storage)", instA.Stored)
	}
	if instB.Stored != 0 {
		t.Errorf("instB.Stored = %d, want 0 (non-adjacent storage)", instB.Stored)
	}
}

func TestDepositRespectsInstanceCapacity(t *testing.T) {
	// Storage at capacity — deposit should be refused and no cooldown queued.
	w := NewWorld(10, 10)
	origin := Point{5, 4}
	w.SetStructure(origin.X, origin.Y, 1, 1, LogStorage)
	w.IndexStructure(origin.X, origin.Y, 1, 1, logStorageDef{})
	p := NewPlayer(5, 5)
	p.Wood = 5
	s := &State{Player: p, World: w, FoundationDeposited: make(map[Point]int)}
	stores := NewStorageManager()
	stores.Register(origin, Wood, 2)
	stores.DepositAt(origin, 2) // fill to capacity via the proper API
	inst := stores.FindByOrigin(origin)
	env := &Env{State: s, Stores: stores}

	s.TickAdjacentStructures(env, time.Now())

	if inst.Stored != 2 {
		t.Errorf("inst.Stored = %d, want 2 (full — no deposit)", inst.Stored)
	}
	if p.Wood != 5 {
		t.Errorf("player Wood = %d, want 5 (no deposit taken)", p.Wood)
	}
}

func TestAddOfferAndSelectCard(t *testing.T) {
	w := NewWorld(10, 10)
	p := NewPlayer(5, 5)
	s := &State{Player: p, World: w, FoundationDeposited: make(map[Point]int)}

	if s.HasPendingOffer() {
		t.Fatal("should have no pending offer initially")
	}

	s.AddOffer([]string{"carry_capacity"})

	if !s.HasPendingOffer() {
		t.Fatal("should have pending offer after AddOffer")
	}

	s.SelectCard(0)

	if s.HasPendingOffer() {
		t.Error("should have no pending offer after SelectCard")
	}
	if p.MaxCarry != 100 {
		t.Errorf("MaxCarry = %d, want 100 after carry capacity upgrade", p.MaxCarry)
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
	if s.HasStructureOfType(FoundationLogStorage) {
		t.Error("should not detect FoundationLogStorage when none placed")
	}
}
