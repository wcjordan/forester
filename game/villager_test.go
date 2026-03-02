package game

import (
	"math/rand"
	"testing"
	"time"

	"forester/game/internal/gametest"
)

// withTestVillagerTypes registers LogStorage as a villager deposit type and
// FoundationHouse as a delivery type for the duration of t. Normally these are
// registered by game/structures init(), which does not run in game package tests.
func withTestVillagerTypes(t *testing.T) {
	t.Helper()
	origDeposit := villagerDepositTypes
	origDelivery := villagerDeliveryTypes
	villagerDepositTypes = []StructureType{gametest.LogStorage}
	villagerDeliveryTypes = []StructureType{gametest.FoundationHouse}
	t.Cleanup(func() {
		villagerDepositTypes = origDeposit
		villagerDeliveryTypes = origDelivery
	})
}

// makeVillagerEnv creates a small world with one log storage and one house
// pre-built and indexed, plus a registered storage manager.
func makeVillagerEnv(t *testing.T) (*State, *Env) {
	t.Helper()
	withTestVillagerTypes(t)
	w := NewWorld(40, 40)

	// Log storage at (5, 5) — 4×4
	lsOrigin := point{X: 5, Y: 5}
	w.PlaceBuilt(lsOrigin.X, lsOrigin.Y, gametest.LogStorageDef{})

	// House at (20, 20) — 2×2
	hOrigin := point{X: 20, Y: 20}
	w.PlaceBuilt(hOrigin.X, hOrigin.Y, testHouseDef{})

	stores := NewStorageManager()
	stores.Register(lsOrigin, Wood, 500)

	s := &State{
		Player:              NewPlayer(10, 30),
		World:               w,
		FoundationDeposited: make(map[point]int),
		completedBeats:      make(map[string]bool),
	}
	env := &Env{State: s, Stores: stores, Villagers: NewVillagerManager()}
	return s, env
}

func advanceVillager(v *Villager, env *Env, rng *rand.Rand, steps int) {
	t0 := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range steps {
		v.Tick(env, rng, t0.Add(time.Duration(i+1)*villagerMoveCooldown*2))
	}
}

// --- WithdrawFrom tests ---

func TestWithdrawFrom(t *testing.T) {
	m := NewStorageManager()
	origin := point{X: 1, Y: 1}
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
		n := m.WithdrawFrom(point{X: 99, Y: 99}, 10)
		if n != 0 {
			t.Errorf("WithdrawFrom unknown = %d, want 0", n)
		}
	})

	t.Run("non-positive amount returns 0", func(t *testing.T) {
		m2 := NewStorageManager()
		m2.Register(point{X: 1, Y: 1}, Wood, 100)
		m2.DepositAt(point{X: 1, Y: 1}, 10)
		if m2.WithdrawFrom(point{X: 1, Y: 1}, 0) != 0 {
			t.Error("WithdrawFrom(0) should return 0")
		}
		if m2.WithdrawFrom(point{X: 1, Y: 1}, -5) != 0 {
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
	m.Register(point{X: 1, Y: 1}, Wood, 200)
	m.Register(point{X: 2, Y: 2}, Wood, 300)
	if m.TotalCapacity(Wood) != 500 {
		t.Errorf("TotalCapacity(Wood) = %d, want 500", m.TotalCapacity(Wood))
	}
}

// --- findNearbyTree tests ---

func TestFindNearestTree(t *testing.T) {
	w := NewWorld(20, 20)

	t.Run("returns false when no trees", func(t *testing.T) {
		_, _, ok := findNearbyTree(w, 10, 10)
		if ok {
			t.Error("findNearbyTree should return false on a world with no trees")
		}
	})

	t.Run("finds a tree in ring-traversal order", func(t *testing.T) {
		// Both trees are at Chebyshev ring 5 from (10,10).
		// chebyshevRingDo visits the top row left-to-right (dx=-5..+5),
		// so (8,5) at dx=-2 is reached before (10,5) at dx=0.
		w.Tiles[5][8] = Tile{Terrain: Forest, TreeSize: 5}
		w.Tiles[5][10] = Tile{Terrain: Forest, TreeSize: 3}
		tx, ty, ok := findNearbyTree(w, 10, 10)
		if !ok {
			t.Fatal("findNearbyTree returned false, want true")
		}
		if tx != 8 || ty != 5 {
			t.Errorf("first tree in ring order = (%d,%d), want (8,5)", tx, ty)
		}
	})

	t.Run("skips exhausted trees (TreeSize=0)", func(t *testing.T) {
		w2 := NewWorld(10, 10)
		w2.Tiles[5][5] = Tile{Terrain: Forest, TreeSize: 0} // stump — skip
		w2.Tiles[8][5] = Tile{Terrain: Forest, TreeSize: 2} // live
		tx, ty, ok := findNearbyTree(w2, 5, 5)
		if !ok {
			t.Fatal("findNearbyTree returned false")
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
	lsOrigin := point{X: 5, Y: 5}

	// Add a FoundationHouse so tryAssignDeliverTask succeeds.
	fhOrigin := point{X: 30, Y: 30}
	s.World.PlaceFoundation(fhOrigin.X, fhOrigin.Y, testHouseDef{})

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
	w.PlaceBuilt(5, 5, gametest.LogStorageDef{})

	tx, ty, ok := nearestClearTileAdjacent(w, gametest.LogStorage, 5, 4, nil)
	if !ok {
		t.Fatal("nearestClearTileAdjacent returned false, want true")
	}
	// The tile returned must not be a structure tile.
	tile := w.TileAt(tx, ty)
	if tile == nil {
		t.Fatal("returned tile is out of bounds")
	}
	if tile.Structure != NoStructure {
		t.Errorf("returned tile (%d,%d) has structure %q, want NoStructure", tx, ty, tile.Structure)
	}

	t.Run("returns false when type not present", func(t *testing.T) {
		_, _, ok := nearestClearTileAdjacent(w, gametest.House, 10, 10, nil)
		if ok {
			t.Error("should return false when no House exists")
		}
	})
}

func TestNearestClearTileAdjacentExcludesCorners(t *testing.T) {
	// 4×4 LogStorage at (5,5) — footprint matches gametest.LogStorageDef{}.Footprint().
	// Cardinal neighbors: top y=4 x∈[5,8], bottom y=9 x∈[5,8],
	//                     left x=4 y∈[5,8], right x=9 y∈[5,8].
	// Chebyshev corners (excluded): (4,4), (9,4), (4,9), (9,9).
	w := NewWorld(20, 20)
	w.PlaceBuilt(5, 5, gametest.LogStorageDef{})

	// Block all cardinal neighbors with House tiles (not indexed as LogStorage).
	for x := 5; x < 9; x++ {
		w.TileAt(x, 4).Structure = gametest.House // top edge
		w.TileAt(x, 9).Structure = gametest.House // bottom edge
	}
	for y := 5; y < 9; y++ {
		w.TileAt(4, y).Structure = gametest.House // left edge
		w.TileAt(9, y).Structure = gametest.House // right edge
	}

	// All cardinal neighbors are blocked; only diagonal corners remain open.
	// The function must treat corners as non-adjacent and return ok=false.
	_, _, ok := nearestClearTileAdjacent(w, gametest.LogStorage, 7, 7, nil)
	if ok {
		t.Error("nearestClearTileAdjacent returned ok=true when only diagonal corners are free; corners must not be considered adjacent")
	}
}

func TestNearestClearTileAdjacentReturnedTileIsCardinallyAdjacent(t *testing.T) {
	// 4×4 LogStorage at (5,5) — footprint matches gametest.LogStorageDef{}.Footprint().
	// Chebyshev corners: (4,4), (9,4), (4,9), (9,9) — must never be returned.
	w := NewWorld(20, 20)
	w.PlaceBuilt(5, 5, gametest.LogStorageDef{})

	corners := map[[2]int]bool{
		{4, 4}: true, {9, 4}: true,
		{4, 9}: true, {9, 9}: true,
	}

	// Run from several positions and confirm no corner is ever returned.
	for _, from := range [][2]int{{0, 0}, {4, 4}, {9, 9}, {10, 10}} {
		tx, ty, ok := nearestClearTileAdjacent(w, gametest.LogStorage, from[0], from[1], nil)
		if !ok {
			t.Errorf("from (%d,%d): expected ok=true", from[0], from[1])
			continue
		}
		if corners[[2]int{tx, ty}] {
			t.Errorf("from (%d,%d): returned corner tile (%d,%d); corners must be excluded", from[0], from[1], tx, ty)
		}
	}
}

// --- headToStorage tests ---

func TestHeadToStorage(t *testing.T) {
	withTestVillagerTypes(t)
	t.Run("prefers non-full storage over closer full one", func(t *testing.T) {
		w := NewWorld(40, 40)

		// Storage A at (5,5) — closer to villager, but full.
		aOrigin := point{X: 5, Y: 5}
		w.PlaceBuilt(aOrigin.X, aOrigin.Y, gametest.LogStorageDef{})

		// Storage B at (20,5) — farther, but has space.
		bOrigin := point{X: 20, Y: 5}
		w.PlaceBuilt(bOrigin.X, bOrigin.Y, gametest.LogStorageDef{})

		stores := NewStorageManager()
		stores.Register(aOrigin, Wood, 10)
		stores.Register(bOrigin, Wood, 10)
		stores.DepositAt(aOrigin, 10) // fill A to capacity

		s := &State{World: w, FoundationDeposited: make(map[point]int), completedBeats: make(map[string]bool)}
		env := &Env{State: s, Stores: stores, Villagers: NewVillagerManager()}

		// Villager is at (0,5): A's nearest neighbor (x=4) is closer than B's (x=19).
		v := &Villager{X: 0, Y: 5, Wood: 3}
		v.headToStorage(env)

		if v.Task != VillagerCarryingToStorage {
			t.Fatalf("task = %d, want VillagerCarryingToStorage", v.Task)
		}
		// Neighbors of A have x ≤ 9; neighbors of B have x ≥ 19.
		if v.TargetX < 19 {
			t.Errorf("target (%d,%d) is adjacent to full storage A; should target non-full storage B", v.TargetX, v.TargetY)
		}
	})

	t.Run("goes idle when all storages are full", func(t *testing.T) {
		w := NewWorld(20, 20)

		origin := point{X: 5, Y: 5}
		w.PlaceBuilt(origin.X, origin.Y, gametest.LogStorageDef{})

		stores := NewStorageManager()
		stores.Register(origin, Wood, 10)
		stores.DepositAt(origin, 10) // fill to capacity

		s := &State{World: w, FoundationDeposited: make(map[point]int), completedBeats: make(map[string]bool)}
		env := &Env{State: s, Stores: stores, Villagers: NewVillagerManager()}

		v := &Villager{X: 0, Y: 5, Wood: 3}
		v.headToStorage(env)

		if v.Task != VillagerIdle {
			t.Errorf("task = %d, want VillagerIdle when all storages full", v.Task)
		}
	})

	t.Run("targets storage that has partial space remaining", func(t *testing.T) {
		w := NewWorld(20, 20)

		origin := point{X: 5, Y: 5}
		w.PlaceBuilt(origin.X, origin.Y, gametest.LogStorageDef{})

		stores := NewStorageManager()
		stores.Register(origin, Wood, 10)
		stores.DepositAt(origin, 7) // 3 units of space remain

		s := &State{World: w, FoundationDeposited: make(map[point]int), completedBeats: make(map[string]bool)}
		env := &Env{State: s, Stores: stores, Villagers: NewVillagerManager()}

		v := &Villager{X: 0, Y: 5, Wood: 3}
		v.headToStorage(env)

		if v.Task != VillagerCarryingToStorage {
			t.Errorf("task = %d, want VillagerCarryingToStorage when storage has space", v.Task)
		}
	})
}

// --- Villager routing around obstacles ---

func TestVillagerRoutesAroundObstacle(t *testing.T) {
	w := NewWorld(20, 20)
	// Vertical wall at X=10, Y=0..14 (width=1, height=15).
	w.PlaceBuilt(10, 0, gametest.WallDef{Width: 1, Height: 15})

	// Villager at (5,7), target at (15,7). Direct route blocked by wall.
	v := &Villager{X: 5, Y: 7, TargetX: 15, TargetY: 7}

	for range 60 {
		v.move(w)
	}

	if v.X != 15 || v.Y != 7 {
		t.Errorf("villager at (%d,%d) after 60 moves, want (15,7)", v.X, v.Y)
	}
}

// --- Villager chops multiple trees before heading to storage ---

// TestVillagerChopsMultipleTrees verifies that a villager retargets another tree
// after exhausting one, rather than heading to storage with a partial carry load.
func TestVillagerChopsMultipleTrees(t *testing.T) {
	withTestVillagerTypes(t)
	// Layout (Tiles[y][x]):
	//   Tree A at (10,11) TreeSize=2: exhausted before VillagerMaxCarry (5) is reached.
	//   Tree B at (10,12) TreeSize=3: brings the villager to full capacity.
	// Log storage at (0,0) is empty so fillRatio=0 and the villager always chooses chop.
	w := NewWorld(30, 30)
	lsOrigin := point{X: 0, Y: 0}
	w.PlaceBuilt(lsOrigin.X, lsOrigin.Y, gametest.LogStorageDef{})
	w.Tiles[11][10] = Tile{Terrain: Forest, TreeSize: 2} // Tree A: X=10, Y=11
	w.Tiles[12][10] = Tile{Terrain: Forest, TreeSize: 3} // Tree B: X=10, Y=12

	stores := NewStorageManager()
	stores.Register(lsOrigin, Wood, 500) // empty → fillRatio=0 → always chop

	s := &State{
		Player:              NewPlayer(20, 20),
		World:               w,
		FoundationDeposited: make(map[point]int),
		completedBeats:      make(map[string]bool),
	}
	env := &Env{State: s, Stores: stores, Villagers: NewVillagerManager()}
	v := &Villager{X: 10, Y: 10}
	rng := rand.New(rand.NewSource(0))

	// Tick 1: Idle → WalkingToTree (nearest = Tree A at (10,11)).
	advanceVillager(v, env, rng, 1)
	if v.Task != VillagerWalkingToTree {
		t.Fatalf("after pickTask: task=%v, want VillagerWalkingToTree", v.Task)
	}

	// Ticks 2–5: move to Tree A, harvest 2 wood, transition through
	// VillagerFindTree, retarget Tree B — should NOT head to storage mid-carry.
	advanceVillager(v, env, rng, 4)
	if w.Tiles[11][10].TreeSize != 0 {
		t.Fatalf("tree A: TreeSize=%d after 5 ticks, want 0 (fully harvested)", w.Tiles[11][10].TreeSize)
	}
	if v.Task == VillagerCarryingToStorage {
		t.Errorf("villager headed to storage with only %d/%d wood; should have retargeted Tree B",
			v.Wood, VillagerMaxCarry)
	}
	if v.Wood == 0 || v.Wood >= VillagerMaxCarry {
		t.Errorf("wood=%d after Tree A exhausted; want 0 < wood < VillagerMaxCarry (%d)", v.Wood, VillagerMaxCarry)
	}

	// Ticks 6–15: move to Tree B, harvest 3 more → full → CarryingToStorage.
	advanceVillager(v, env, rng, 10)
	if w.Tiles[12][10].TreeSize != 0 {
		t.Errorf("tree B: TreeSize=%d, want 0 (fully harvested)", w.Tiles[12][10].TreeSize)
	}
	if v.Task != VillagerCarryingToStorage {
		t.Errorf("after filling carry from two trees: task=%v, want VillagerCarryingToStorage", v.Task)
	}
	if v.Wood != VillagerMaxCarry {
		t.Errorf("wood=%d after filling carry capacity, want %d", v.Wood, VillagerMaxCarry)
	}
}

// --- CountStructureInstances ---

func TestCountStructureInstances(t *testing.T) {
	w := NewWorld(30, 30)
	if w.CountStructureInstances(gametest.House) != 0 {
		t.Error("want 0 initially")
	}
	w.PlaceBuilt(2, 2, testHouseDef{})
	if w.CountStructureInstances(gametest.House) != 1 {
		t.Errorf("want 1 after placing one house, got %d", w.CountStructureInstances(gametest.House))
	}
	w.PlaceBuilt(10, 10, testHouseDef{})
	if w.CountStructureInstances(gametest.House) != 2 {
		t.Errorf("want 2 after placing two houses, got %d", w.CountStructureInstances(gametest.House))
	}
}
