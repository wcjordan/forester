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
	s := &State{Player: p, World: w, FoundationDeposited: make(map[Point]int), completedBeats: make(map[string]bool)}
	stores := NewStorageManager()
	env := &Env{State: s, Stores: stores}

	// Harvest InitialCarryingCapacity-1 times — foundation should not appear yet.
	t0 := time.Now()
	for i := range InitialCarryingCapacity - 1 {
		now := t0.Add(time.Duration(i+1) * harvestTickInterval * 2)
		IterateResources(func(d ResourceDef) { d.Harvest(env, now) })
		maybeAdvanceStory(env)
		maybeSpawnFoundation(env)
	}
	if s.World.HasStructureOfType(FoundationLogStorage) {
		t.Fatal("foundation appeared before inventory full")
	}

	// Final harvest fills inventory — story beat fires and foundation should now appear.
	now := t0.Add(time.Duration(InitialCarryingCapacity) * harvestTickInterval * 2)
	IterateResources(func(d ResourceDef) { d.Harvest(env, now) })
	maybeAdvanceStory(env)
	maybeSpawnFoundation(env)
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
	s := &State{Player: p, World: w, FoundationDeposited: make(map[Point]int), completedBeats: make(map[string]bool)}
	stores := NewStorageManager()
	env := &Env{State: s, Stores: stores}

	maybeAdvanceStory(env)
	count1 := countStructureTiles(w, FoundationLogStorage)

	// Second call — beat is now marked complete; should not spawn another foundation.
	maybeAdvanceStory(env)
	count2 := countStructureTiles(w, FoundationLogStorage)

	if count1 == 0 {
		t.Error("story beat did not fire on first call")
	}
	if count2 != count1 {
		t.Errorf("story beat fired twice; foundation count changed from %d to %d", count1, count2)
	}
}
