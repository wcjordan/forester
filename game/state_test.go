package game

import (
	"math/rand"
	"testing"
	"time"
)

// testWoodDef is a minimal ResourceDef for use in package game tests.
// It mimics the harvest logic of woodDef without importing game/resources.
type testWoodDef struct{}

func (testWoodDef) Type() ResourceType { return Wood }
func (testWoodDef) Harvest(env *Env, now time.Time) {
	p := env.State.Player
	if !p.CooldownExpired(Harvest, now) {
		return
	}
	p.SetCooldown(Harvest, now.Add(p.HarvestInterval))
	if p.Inventory[Wood] >= p.MaxCarry {
		return
	}
	dx, dy := p.FacingDX, p.FacingDY
	targets := [4][2]int{
		{p.X, p.Y},
		{p.X + dx, p.Y + dy},
		{p.X + dx - dy, p.Y + dy + dx},
		{p.X + dx + dy, p.Y + dy - dx},
	}
	for _, coord := range targets {
		tile := env.State.World.TileAt(coord[0], coord[1])
		if tile == nil || tile.Terrain != Forest {
			continue
		}
		canTake := min(1, p.MaxCarry-p.Inventory[Wood])
		harvest := min(canTake, tile.TreeSize)
		tile.TreeSize -= harvest
		p.Inventory[Wood] += harvest
	}
}
func (testWoodDef) Regrow(_ *Env, _ *rand.Rand, _ time.Time) {}

// withTestResources registers testWoodDef for the duration of t and
// restores the original registry on cleanup.
func withTestResources(t *testing.T) {
	t.Helper()
	orig := resourceRegistry
	resourceRegistry = map[ResourceType]ResourceDef{}
	RegisterResource(testWoodDef{})
	t.Cleanup(func() { resourceRegistry = orig })
}

// testLogStorageDef is a minimal StructureDef for spawning tests in package game.
// It mimics just enough of logStorageDef (now in game/structures) to exercise
// spawnFoundationAt and placement helpers without importing the structures subpackage.
// ShouldSpawn returns false: initial log storage spawning is owned by story beats.
type testLogStorageDef struct{}

func (testLogStorageDef) FoundationType() StructureType                    { return FoundationLogStorage }
func (testLogStorageDef) BuiltType() StructureType                         { return LogStorage }
func (testLogStorageDef) Footprint() (w, h int)                            { return 4, 4 }
func (testLogStorageDef) BuildCost() int                                   { return 20 }
func (testLogStorageDef) ShouldSpawn(_ *Env) bool                          { return false }
func (testLogStorageDef) OnPlayerInteraction(_ *Env, _ Point, _ time.Time) {}
func (testLogStorageDef) OnBuilt(_ *Env, _ Point)                          {}

// testHouseDef is a minimal StructureDef for house world-condition tests.
// ShouldSpawn implements the same gate as houseDef: at least one built house,
// no pending house foundation.
type testHouseDef struct{}

func (testHouseDef) FoundationType() StructureType { return FoundationHouse }
func (testHouseDef) BuiltType() StructureType      { return House }
func (testHouseDef) Footprint() (w, h int)         { return 2, 2 }
func (testHouseDef) BuildCost() int                { return 50 }
func (testHouseDef) ShouldSpawn(env *Env) bool {
	built := len(env.State.World.StructureTypeIndex[House])
	pending := len(env.State.World.StructureTypeIndex[FoundationHouse])
	return built >= 1 && pending == 0
}
func (testHouseDef) OnPlayerInteraction(_ *Env, _ Point, _ time.Time) {}
func (testHouseDef) OnBuilt(_ *Env, _ Point)                          {}

// withTestStructures registers testLogStorageDef for the duration of t and
// restores the original registry on cleanup.
func withTestStructures(t *testing.T) {
	t.Helper()
	orig := structures
	structures = map[StructureType]StructureDef{}
	RegisterStructure(testLogStorageDef{})
	t.Cleanup(func() { structures = orig })
}

// countStructureTiles returns the number of tiles in w that have the given structure type.
func countStructureTiles(w *World, st StructureType) int {
	n := 0
	for y := range w.Tiles {
		for x := range w.Tiles[y] {
			if w.Tiles[y][x].Structure == st {
				n++
			}
		}
	}
	return n
}

func TestFoundationSpawnsWhenInventoryFull(t *testing.T) {
	withTestStructures(t)
	withTestResources(t)
	// Build a world big enough that there's a clear path from player to center.
	w := NewWorld(30, 30)
	// Player at (5, 5) facing north; forest tile at (5, 4) with enough wood.
	w.Tiles[4][5] = Tile{Terrain: Forest, TreeSize: InitialCarryingCapacity}
	p := NewPlayer(5, 5)
	s := &State{Player: p, World: w, FoundationDeposited: make(map[Point]int), CompletedBeats: make(map[string]bool)}
	stores := NewStorageManager()
	env := &Env{State: s, Stores: stores}

	// Harvest InitialCarryingCapacity-1 times — foundation should not appear yet.
	t0 := time.Now()
	for i := range InitialCarryingCapacity - 1 {
		s.Harvest(env, t0.Add(time.Duration(i+1)*HarvestTickInterval*2))
	}
	if s.World.HasStructureOfType(FoundationLogStorage) {
		t.Fatal("foundation appeared before inventory full")
	}

	// Final harvest fills inventory — story beat fires and foundation should now appear.
	s.Harvest(env, t0.Add(time.Duration(InitialCarryingCapacity)*HarvestTickInterval*2))
	if !s.World.HasStructureOfType(FoundationLogStorage) {
		t.Error("foundation did not appear when inventory became full")
	}
}

func TestStoryBeatFiresOnce(t *testing.T) {
	withTestStructures(t)
	w := NewWorld(30, 30)
	for i := 0; i < 15; i++ {
		w.Tiles[4][5+i] = Tile{Terrain: Forest, TreeSize: 1}
	}
	p := NewPlayer(5, 5)
	p.Inventory[Wood] = InitialCarryingCapacity
	s := &State{Player: p, World: w, FoundationDeposited: make(map[Point]int), CompletedBeats: make(map[string]bool)}
	stores := NewStorageManager()
	env := &Env{State: s, Stores: stores}

	s.maybeAdvanceStory(env)
	count1 := countStructureTiles(w, FoundationLogStorage)

	// Second call — beat is now marked complete; should not spawn another foundation.
	s.maybeAdvanceStory(env)
	count2 := countStructureTiles(w, FoundationLogStorage)

	if count1 == 0 {
		t.Error("story beat did not fire on first call")
	}
	if count2 != count1 {
		t.Errorf("story beat fired twice; foundation count changed from %d to %d", count1, count2)
	}
}

func TestFoundationLocationIsAllGrassland(t *testing.T) {
	w := NewWorld(30, 30)
	p := NewPlayer(5, 15)
	spawnFoundationAt(w, p.X, p.Y, testLogStorageDef{})

	// Find the foundation and verify all 16 tiles are on grassland terrain (underlying).
	for y := range w.Tiles {
		for x := range w.Tiles[y] {
			if w.Tiles[y][x].Structure == FoundationLogStorage {
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
	spawnFoundationAt(w, p.X, p.Y, testLogStorageDef{})

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

func TestHouseWorldConditionSpawnsAfterBuild(t *testing.T) {
	orig := structures
	structures = map[StructureType]StructureDef{}
	RegisterStructure(testHouseDef{})
	t.Cleanup(func() { structures = orig })

	w := NewWorld(30, 30)
	p := NewPlayer(5, 15)
	s := &State{Player: p, World: w, FoundationDeposited: make(map[Point]int), CompletedBeats: make(map[string]bool)}
	stores := NewStorageManager()
	env := &Env{State: s, Stores: stores}

	// No house built yet — world condition should not fire.
	maybeSpawnFoundation(env)
	if s.World.HasStructureOfType(FoundationHouse) {
		t.Error("house foundation spawned before any house was built")
	}

	// Place a built house to satisfy the world condition.
	w.SetStructure(10, 10, 2, 2, House)
	w.IndexStructure(10, 10, 2, 2, testHouseDef{})

	// World condition now satisfied: built house exists, no pending foundation.
	maybeSpawnFoundation(env)
	if !s.World.HasStructureOfType(FoundationHouse) {
		t.Error("house foundation did not spawn after a house was built")
	}

	// Second call must not spawn another foundation while one is already pending.
	maybeSpawnFoundation(env)
	if countStructureTiles(w, FoundationHouse) > 4 { // one 2×2 foundation = 4 tiles
		t.Error("world condition spawned a second house foundation while one was already pending")
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
	s := &State{Player: p, World: w, FoundationDeposited: make(map[Point]int), CompletedBeats: make(map[string]bool)}

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
