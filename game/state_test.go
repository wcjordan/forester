package game

import (
	"testing"
	"time"
)

// testLogStorageDef is a minimal StructureDef for spawning tests in package game.
// It mimics just enough of logStorageDef (now in game/structures) to exercise
// maybeSpawnFoundation without importing the structures subpackage.
type testLogStorageDef struct{}

func (testLogStorageDef) FoundationType() StructureType                    { return FoundationLogStorage }
func (testLogStorageDef) BuiltType() StructureType                         { return LogStorage }
func (testLogStorageDef) Footprint() (w, h int)                            { return 4, 4 }
func (testLogStorageDef) BuildCost() int                                   { return 20 }
func (testLogStorageDef) ShouldSpawn(env *Env) bool                        { return env.State.Player.Wood >= 20 }
func (testLogStorageDef) OnPlayerInteraction(_ *Env, _ Point, _ time.Time) {}
func (testLogStorageDef) OnBuilt(_ *Env, _ Point)                          {}

// withTestStructures registers testLogStorageDef for the duration of t and
// restores the original registry on cleanup.
func withTestStructures(t *testing.T) {
	t.Helper()
	orig := structures
	structures = []StructureDef{testLogStorageDef{}}
	t.Cleanup(func() { structures = orig })
}

func TestFoundationSpawnsWhenInventoryFull(t *testing.T) {
	withTestStructures(t)
	// Build a world big enough that there's a clear path from player to center.
	w := NewWorld(30, 30)
	// Player at (5, 5) facing north; forest tile at (5, 4) with enough wood.
	w.Tiles[4][5] = Tile{Terrain: Forest, TreeSize: InitialCarryingCapacity}
	p := NewPlayer(5, 5)
	s := &State{Player: p, World: w, FoundationDeposited: make(map[Point]int)}
	stores := NewStorageManager()
	env := &Env{State: s, Stores: stores}

	// Harvest InitialCarryingCapacity-1 times — foundation should not appear yet.
	t0 := time.Now()
	for i := range InitialCarryingCapacity - 1 {
		s.Harvest(env, t0.Add(time.Duration(i+1)*HarvestTickInterval*2))
	}
	if s.HasStructureOfType(FoundationLogStorage) {
		t.Fatal("foundation appeared before inventory full")
	}

	// Final harvest fills inventory — foundation should now appear.
	s.Harvest(env, t0.Add(time.Duration(InitialCarryingCapacity)*HarvestTickInterval*2))
	if !s.HasStructureOfType(FoundationLogStorage) {
		t.Error("foundation did not appear when inventory became full")
	}
}

func TestFoundationDoesNotSpawnTwice(t *testing.T) {
	withTestStructures(t)
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
	withTestStructures(t)
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
	withTestStructures(t)
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

// testCarryUpgrade is a minimal UpgradeDef used only in TestAddOfferAndSelectCard
// so the test stays decoupled from the game/upgrades package.
type testCarryUpgrade struct{}

func (testCarryUpgrade) ID() string          { return "test_carry" }
func (testCarryUpgrade) Name() string        { return "Test Carry" }
func (testCarryUpgrade) Description() string { return "test" }
func (testCarryUpgrade) Apply(p *Player)     { p.MaxCarry = 100 }

func TestAddOfferAndSelectCard(t *testing.T) {
	upgradeRegistry["test_carry"] = testCarryUpgrade{}
	t.Cleanup(func() { delete(upgradeRegistry, "test_carry") })

	w := NewWorld(10, 10)
	p := NewPlayer(5, 5)
	s := &State{Player: p, World: w, FoundationDeposited: make(map[Point]int)}

	if s.HasPendingOffer() {
		t.Fatal("should have no pending offer initially")
	}

	s.AddOffer([]string{"test_carry"})

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
