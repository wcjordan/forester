package game

import (
	"testing"

	"forester/game/internal/gametest"
)

// testHouseDef is a minimal StructureDef descriptor for house world-condition tests.
// ShouldSpawn logic is wired in as a StructureCallbacks when registering.
type testHouseDef struct{}

func (testHouseDef) FoundationType() StructureType { return gametest.FoundationHouse }
func (testHouseDef) BuiltType() StructureType      { return gametest.House }
func (testHouseDef) Footprint() (w, h int)         { return 2, 2 }
func (testHouseDef) BuildCost() int                { return 50 }

func TestFoundationLocationIsAllGrassland(t *testing.T) {
	w := NewWorld(30, 30)
	p := NewPlayer(5, 15)
	s := &State{Player: p, World: w, FoundationDeposited: make(map[point]int)}
	env := &Env{State: s, Stores: NewStorageManager()}
	spawnFoundationAt(env, gametest.LogStorageDef{})

	// Find the foundation and verify all 16 tiles are on grassland terrain (underlying).
	for y := range w.Tiles {
		for x := range w.Tiles[y] {
			if w.Tiles[y][x].Structure == gametest.FoundationLogStorage {
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
	s := &State{Player: p, World: w, FoundationDeposited: make(map[point]int)}
	env := &Env{State: s, Stores: NewStorageManager()}
	spawnFoundationAt(env, gametest.LogStorageDef{})

	spawnX := w.Width / 2
	// Find foundation top-left.
	gx := -1
	for y := range w.Tiles {
		for x := range w.Tiles[y] {
			if w.Tiles[y][x].Structure == gametest.FoundationLogStorage && gx == -1 {
				gx = x
			}
		}
	}
	if gx == -1 {
		t.Fatal("no foundation placed")
	}
	// Foundation x-coordinate should be between player and spawn center.
	if gx < p.TileX() || gx > spawnX {
		t.Errorf("foundation x=%d not between player x=%d and spawn x=%d", gx, p.TileX(), spawnX)
	}
}

func TestHouseWorldConditionSpawnsAfterBuild(t *testing.T) {
	orig := structures
	structures = map[StructureType]registeredDef{}
	RegisterStructure(testHouseDef{}, StructureCallbacks{
		ShouldSpawn: func(env *Env) bool {
			built := len(env.State.World.StructureTypeIndex[gametest.House])
			pending := len(env.State.World.StructureTypeIndex[gametest.FoundationHouse])
			return built >= 1 && pending == 0
		},
	})
	t.Cleanup(func() { structures = orig })

	w := NewWorld(30, 30)
	p := NewPlayer(5, 15)
	s := &State{Player: p, World: w, FoundationDeposited: make(map[point]int), completedBeats: make(map[string]bool)}
	stores := NewStorageManager()
	env := &Env{State: s, Stores: stores}

	// No house built yet — world condition should not fire.
	maybeSpawnFoundation(env)
	if s.World.HasStructureOfType(gametest.FoundationHouse) {
		t.Error("house foundation spawned before any house was built")
	}

	// Place a built house to satisfy the world condition.
	w.PlaceBuilt(10, 10, testHouseDef{})

	// World condition now satisfied: built house exists, no pending foundation.
	maybeSpawnFoundation(env)
	if !s.World.HasStructureOfType(gametest.FoundationHouse) {
		t.Error("house foundation did not spawn after a house was built")
	}

	// Second call must not spawn another foundation while one is already pending.
	maybeSpawnFoundation(env)
	if countStructureTiles(w, gametest.FoundationHouse) > 4 { // one 2×2 foundation = 4 tiles
		t.Error("world condition spawned a second house foundation while one was already pending")
	}
}
