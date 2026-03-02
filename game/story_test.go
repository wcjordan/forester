package game

import (
	"math/rand"
	"testing"
	"time"

	"forester/game/internal/gametest"
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

// withTestStructures registers gametest.LogStorageDef and the test story beats for the
// duration of t, then restores the original registries on cleanup. Story beats must
// be set up here because game package tests do not import game/structures, so its
// init() functions (which normally register the beats) do not run.
func withTestStructures(t *testing.T) {
	t.Helper()
	origStructures := structures
	origBeats := storyBeats
	structures = map[StructureType]registeredDef{}
	storyBeats = nil
	RegisterStructure(gametest.LogStorageDef{}, StructureCallbacks{})
	RegisterStoryBeat(100, "initial_log_storage",
		func(env *Env) bool {
			p := env.State.Player
			return p.Inventory[Wood] >= p.MaxCarry
		},
		func(env *Env) bool {
			return SpawnFoundationByType(env, gametest.FoundationLogStorage)
		},
	)
	RegisterStoryBeat(200, "first_log_storage_built",
		func(env *Env) bool { return env.State.World.HasStructureOfType(gametest.LogStorage) },
		func(env *Env) bool {
			env.State.AddOffer([]string{"carry_capacity"})
			return true
		},
	)
	t.Cleanup(func() {
		structures = origStructures
		storyBeats = origBeats
	})
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
	s := &State{Player: p, World: w, FoundationDeposited: make(map[point]int), completedBeats: make(map[string]bool)}
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
	if s.World.HasStructureOfType(gametest.FoundationLogStorage) {
		t.Fatal("foundation appeared before inventory full")
	}

	// Final harvest fills inventory — story beat fires and foundation should now appear.
	now := t0.Add(time.Duration(InitialCarryingCapacity) * harvestTickInterval * 2)
	IterateResources(func(d ResourceDef) { d.Harvest(env, now) })
	maybeAdvanceStory(env)
	maybeSpawnFoundation(env)
	if !s.World.HasStructureOfType(gametest.FoundationLogStorage) {
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
	s := &State{Player: p, World: w, FoundationDeposited: make(map[point]int), completedBeats: make(map[string]bool)}
	stores := NewStorageManager()
	env := &Env{State: s, Stores: stores}

	maybeAdvanceStory(env)
	count1 := countStructureTiles(w, gametest.FoundationLogStorage)

	// Second call — beat is now marked complete; should not spawn another foundation.
	maybeAdvanceStory(env)
	count2 := countStructureTiles(w, gametest.FoundationLogStorage)

	if count1 == 0 {
		t.Error("story beat did not fire on first call")
	}
	if count2 != count1 {
		t.Errorf("story beat fired twice; foundation count changed from %d to %d", count1, count2)
	}
}
