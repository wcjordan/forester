package game

import (
	"testing"
	"time"
)

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
	gx := -1
	for y := range w.Tiles {
		for x := range w.Tiles[y] {
			if w.Tiles[y][x].Structure == FoundationLogStorage && gx == -1 {
				gx = x
			}
		}
	}
	if gx == -1 {
		t.Fatal("no foundation placed")
	}
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
