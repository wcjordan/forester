package game

import (
	"math/rand"
	"testing"
	"time"
)

// makeVillagerEnv creates a small world with one log storage and one house
// pre-built and indexed, plus a registered storage manager.
func makeVillagerEnv(t *testing.T) (*State, *Env) {
	t.Helper()
	w := NewWorld(40, 40)

	// Log storage at (5, 5) — 4×4
	lsOrigin := Point{X: 5, Y: 5}
	w.SetStructure(lsOrigin.X, lsOrigin.Y, 4, 4, LogStorage)
	w.IndexStructure(lsOrigin.X, lsOrigin.Y, 4, 4, testLogStorageDef{})

	// House at (20, 20) — 2×2
	hOrigin := Point{X: 20, Y: 20}
	w.SetStructure(hOrigin.X, hOrigin.Y, 2, 2, House)
	w.IndexStructure(hOrigin.X, hOrigin.Y, 2, 2, testHouseDef{})

	stores := NewStorageManager()
	stores.Register(lsOrigin, Wood, 500)

	s := &State{
		Player:              NewPlayer(10, 30),
		World:               w,
		FoundationDeposited: make(map[Point]int),
		CompletedBeats:      make(map[string]bool),
	}
	env := &Env{State: s, Stores: stores, Villagers: NewVillagerManager()}
	return s, env
}

func advanceVillager(v *Villager, env *Env, rng *rand.Rand, steps int) {
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range steps {
		v.Tick(env, rng, t0.Add(time.Duration(i+1)*VillagerMoveCooldown*2))
	}
}

// --- WithdrawFrom tests ---

func TestWithdrawFrom(t *testing.T) {
	m := NewStorageManager()
	origin := Point{X: 1, Y: 1}
	m.Register(origin, Wood, 100)
	m.DepositAt(origin, 10)

	t.Run("withdraws up to amount", func(t *testing.T) {
		withdrawn := m.WithdrawFrom(origin, 5)
		if withdrawn != 5 {
			t.Errorf("WithdrawFrom returned %d, want 5", withdrawn)
		}
		if m.Total(Wood) != 5 {
			t.Errorf("Total(Wood) = %d, want 5 after withdrawal", m.Total(Wood))
		}
	})

	t.Run("cannot withdraw more than stored", func(t *testing.T) {
		withdrawn := m.WithdrawFrom(origin, 100)
		if withdrawn != 5 {
			t.Errorf("WithdrawFrom returned %d, want 5 (only 5 left)", withdrawn)
		}
		if m.Total(Wood) != 0 {
			t.Errorf("Total(Wood) = %d, want 0 after draining", m.Total(Wood))
		}
	})

	t.Run("withdraw from unknown origin returns 0", func(t *testing.T) {
		n := m.WithdrawFrom(Point{99, 99}, 10)
		if n != 0 {
			t.Errorf("WithdrawFrom unknown = %d, want 0", n)
		}
	})

	t.Run("non-positive amount returns 0", func(t *testing.T) {
		m2 := NewStorageManager()
		m2.Register(Point{1, 1}, Wood, 100)
		m2.DepositAt(Point{1, 1}, 10)
		if m2.WithdrawFrom(Point{1, 1}, 0) != 0 {
			t.Error("WithdrawFrom(0) should return 0")
		}
		if m2.WithdrawFrom(Point{1, 1}, -5) != 0 {
			t.Error("WithdrawFrom(-5) should return 0")
		}
		if m2.Total(Wood) != 10 {
			t.Errorf("Total(Wood) = %d, want 10 after zero-amount withdrawals", m2.Total(Wood))
		}
	})
}

// --- TotalCapacity tests ---

func TestTotalCapacity(t *testing.T) {
	m := NewStorageManager()
	if m.TotalCapacity(Wood) != 0 {
		t.Error("TotalCapacity should be 0 when no storage registered")
	}
	m.Register(Point{1, 1}, Wood, 200)
	m.Register(Point{2, 2}, Wood, 300)
	if m.TotalCapacity(Wood) != 500 {
		t.Errorf("TotalCapacity(Wood) = %d, want 500", m.TotalCapacity(Wood))
	}
}

// --- findNearestTree tests ---

func TestFindNearestTree(t *testing.T) {
	w := NewWorld(20, 20)

	t.Run("returns false when no trees", func(t *testing.T) {
		_, _, ok := findNearestTree(w, 10, 10)
		if ok {
			t.Error("findNearestTree should return false on a world with no trees")
		}
	})

	t.Run("finds the closest tree", func(t *testing.T) {
		w.Tiles[5][8] = Tile{Terrain: Forest, TreeSize: 5}  // dist^2 = 4+25 = 29
		w.Tiles[5][10] = Tile{Terrain: Forest, TreeSize: 3} // dist^2 = 0+25 = 25 (closer)
		tx, ty, ok := findNearestTree(w, 10, 10)
		if !ok {
			t.Fatal("findNearestTree returned false, want true")
		}
		if tx != 10 || ty != 5 {
			t.Errorf("nearest tree = (%d,%d), want (10,5)", tx, ty)
		}
	})

	t.Run("skips exhausted trees (TreeSize=0)", func(t *testing.T) {
		w2 := NewWorld(10, 10)
		w2.Tiles[5][5] = Tile{Terrain: Forest, TreeSize: 0} // stump — skip
		w2.Tiles[8][5] = Tile{Terrain: Forest, TreeSize: 2} // live
		tx, ty, ok := findNearestTree(w2, 5, 5)
		if !ok {
			t.Fatal("findNearestTree returned false")
		}
		if tx != 5 || ty != 8 {
			t.Errorf("nearest live tree = (%d,%d), want (5,8)", tx, ty)
		}
	})
}

// --- Villager task: chop wood → carry to storage ---

func TestVillagerChopsAndCarries(t *testing.T) {
	s, env := makeVillagerEnv(t)

	// Place a tree near (15, 15).
	s.World.Tiles[15][15] = Tile{Terrain: Forest, TreeSize: 10}

	// Villager starts at (15, 14) — one north of the tree.
	v := &Villager{X: 15, Y: 14}

	// Force chop task (storage is empty → fillRatio=0 → always chop).
	rng := rand.New(rand.NewSource(0))
	advanceVillager(v, env, rng, 1) // pickTask → WalkingToTree
	if v.Task != VillagerWalkingToTree {
		t.Fatalf("task = %d after pickTask, want VillagerWalkingToTree", v.Task)
	}

	// Advance until villager reaches tree and fills up.
	advanceVillager(v, env, rng, VillagerMaxCarry+5)
	if v.Wood != 0 {
		// If still has wood it hasn't deposited yet — that's fine, check it's heading to storage.
		if v.Task != VillagerCarryingToStorage && v.Task != VillagerIdle {
			t.Errorf("after harvesting: task = %d, want CarryingToStorage or Idle", v.Task)
		}
	}
	// Tree should have lost some wood.
	if s.World.Tiles[15][15].TreeSize >= 10 {
		t.Error("tree size should have decreased after villager harvested it")
	}
}

// --- Villager task: fetch from storage → deliver to house foundation ---

func TestVillagerFetchesAndDelivers(t *testing.T) {
	s, env := makeVillagerEnv(t)
	lsOrigin := Point{X: 5, Y: 5}

	// Add a FoundationHouse so tryAssignDeliverTask succeeds.
	fhOrigin := Point{X: 30, Y: 30}
	s.World.SetStructure(fhOrigin.X, fhOrigin.Y, 2, 2, FoundationHouse)
	s.World.IndexStructure(fhOrigin.X, fhOrigin.Y, 2, 2, testHouseDef{})

	// Pre-fill storage so fillRatio=1 → villager always wants to deliver.
	env.Stores.DepositAt(lsOrigin, 500)

	// Villager starts adjacent to storage: at (4, 5) — one left of top-left.
	v := &Villager{X: 4, Y: 5}

	rng := rand.New(rand.NewSource(42))
	advanceVillager(v, env, rng, 1) // pickTask → WalkingToStorage
	if v.Task != VillagerWalkingToStorage {
		t.Fatalf("task = %d, want VillagerWalkingToStorage (storage full → deliver)", v.Task)
	}

	// One step: arrives at target (should already be there or very close), fetches wood.
	advanceVillager(v, env, rng, 5)

	// After fetching, should have wood and be heading to house foundation.
	if v.Task == VillagerWalkingToStorage {
		// Give it more steps to reach target.
		advanceVillager(v, env, rng, 20)
	}

	// Storage should have less wood than it started with.
	if env.Stores.Total(Wood) >= 500 {
		t.Errorf("storage still full (%d); villager should have fetched some", env.Stores.Total(Wood))
	}
}

// --- nearestClearTileAdjacent tests ---

func TestNearestClearTileAdjacent(t *testing.T) {
	w := NewWorld(20, 20)
	// Log storage at (5,5) 4×4 — must call IndexStructure so StructureInstanceIndex is populated.
	w.SetStructure(5, 5, 4, 4, LogStorage)
	w.IndexStructure(5, 5, 4, 4, testLogStorageDef{})

	tx, ty, ok := nearestClearTileAdjacent(w, LogStorage, 5, 4)
	if !ok {
		t.Fatal("nearestClearTileAdjacent returned false, want true")
	}
	// The tile returned must not be a structure tile.
	tile := w.TileAt(tx, ty)
	if tile == nil {
		t.Fatal("returned tile is out of bounds")
	}
	if tile.Structure != NoStructure {
		t.Errorf("returned tile (%d,%d) has structure %d, want NoStructure", tx, ty, tile.Structure)
	}

	t.Run("returns false when type not present", func(t *testing.T) {
		_, _, ok := nearestClearTileAdjacent(w, House, 10, 10)
		if ok {
			t.Error("should return false when no House exists")
		}
	})
}

// --- Villager routing around obstacles ---

func TestVillagerRoutesAroundObstacle(t *testing.T) {
	w := NewWorld(20, 20)
	// Vertical wall at X=10, Y=0..14 (width=1, height=15).
	w.SetStructure(10, 0, 1, 15, LogStorage)

	// Villager at (5,7), target at (15,7). Direct route blocked by wall.
	v := &Villager{X: 5, Y: 7, TargetX: 15, TargetY: 7}

	for range 60 {
		v.move(w)
	}

	if v.X != 15 || v.Y != 7 {
		t.Errorf("villager at (%d,%d) after 60 moves, want (15,7)", v.X, v.Y)
	}
}

// --- CountStructureInstances ---

func TestCountStructureInstances(t *testing.T) {
	w := NewWorld(30, 30)
	if w.CountStructureInstances(House) != 0 {
		t.Error("want 0 initially")
	}
	w.SetStructure(2, 2, 2, 2, House)
	w.IndexStructure(2, 2, 2, 2, testHouseDef{})
	if w.CountStructureInstances(House) != 1 {
		t.Errorf("want 1 after placing one house, got %d", w.CountStructureInstances(House))
	}
	w.SetStructure(10, 10, 2, 2, House)
	w.IndexStructure(10, 10, 2, 2, testHouseDef{})
	if w.CountStructureInstances(House) != 2 {
		t.Errorf("want 2 after placing two houses, got %d", w.CountStructureInstances(House))
	}
}
