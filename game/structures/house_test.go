package structures

import (
	"math/rand"
	"testing"
	"time"

	"forester/game"
	"forester/game/geom"
)

// makeHouseGame creates a game with a 30×30 world and a fresh seeded RNG for XP milestone handling.
func makeHouseGame() *game.Game {
	g := game.NewWithClockAndRNG(game.RealClock{}, rand.New(rand.NewSource(0)))
	// Replace state with a minimal world.
	w := game.NewWorld(30, 30)
	p := game.NewPlayer(15, 15)
	g.State = &game.State{
		Player:              p,
		World:               w,
		FoundationDeposited: make(map[geom.Point]int),
		HouseOccupancy:      make(map[geom.Point]bool),
	}
	g.Stores = game.NewStorageManager()
	g.Villagers = game.NewVillagerManager()
	return g
}

// buildHouseAt builds a house foundation at (ox, oy) with the player at (playerX, playerY).
func buildHouseAt(g *game.Game, ox, oy, playerX, playerY int) {
	g.State.World.PlaceFoundation(ox, oy, houseDef{})
	g.State.Player.X = playerX
	g.State.Player.Y = playerY
	g.State.Player.Inventory[game.Wood] = houseBuildCost
	g.State.Player.SetCooldown(game.Build, time.Time{})
	t0 := time.Now()
	for i := range houseBuildCost {
		g.TickAdjacentStructures(t0.Add(time.Duration(i) * (game.DepositTickInterval + time.Millisecond)))
	}
}

// TestHouseBuiltMarkedUnoccupied verifies that completing a house registers it as
// unoccupied in HouseOccupancy (villager spawning is now card-gated, not automatic).
func TestHouseBuiltMarkedUnoccupied(t *testing.T) {
	g := makeHouseGame()
	origin := geom.Point{X: 10, Y: 10}
	buildHouseAt(g, origin.X, origin.Y, 9, 10)

	if !g.State.World.HasStructureOfType(game.House) {
		t.Fatal("house was not built after depositing build cost")
	}
	// No villager should be auto-spawned.
	if g.Villagers.Count() != 0 {
		t.Errorf("villager count = %d, want 0 (villager spawning is card-gated)", g.Villagers.Count())
	}
	// House must be tracked as unoccupied.
	occupied, exists := g.State.HouseOccupancy[origin]
	if !exists {
		t.Error("house origin not found in HouseOccupancy after build")
	}
	if occupied {
		t.Error("house should be unoccupied immediately after build")
	}
}

// TestSpawnVillagerAtHouse verifies that SpawnVillagerAtHouse places the villager
// on a clear tile adjacent to the house and marks it occupied.
func TestSpawnVillagerAtHouse(t *testing.T) {
	g := makeHouseGame()
	origin := geom.Point{X: 10, Y: 10}
	buildHouseAt(g, origin.X, origin.Y, 9, 10)

	env := &game.Env{State: g.State, Stores: g.Stores, Villagers: g.Villagers, RNG: rand.New(rand.NewSource(0))}
	spawned := game.SpawnVillagerAtHouse(env, origin)
	if !spawned {
		t.Fatal("SpawnVillagerAtHouse returned false")
	}
	if g.Villagers.Count() != 1 {
		t.Fatalf("villager count = %d, want 1", g.Villagers.Count())
	}

	v := g.Villagers.Villagers[0]
	// Villager must not be inside the house footprint.
	insideHouse := v.X >= origin.X && v.X < origin.X+2 && v.Y >= origin.Y && v.Y < origin.Y+2
	if insideHouse {
		t.Errorf("villager spawned inside house footprint at (%d, %d)", v.X, v.Y)
	}
	// Villager must be on a clear tile.
	tile := g.State.World.TileAt(v.X, v.Y)
	if tile == nil {
		t.Fatalf("villager spawned out of bounds at (%d, %d)", v.X, v.Y)
	}
	if tile.Structure != game.NoStructure {
		t.Errorf("villager spawned on a structure tile at (%d, %d)", v.X, v.Y)
	}
	// House must now be marked occupied.
	if !g.State.HouseOccupancy[origin] {
		t.Error("house should be occupied after SpawnVillagerAtHouse")
	}
}

// TestEachHouseTrackedInOccupancy verifies that multiple built houses are each
// independently tracked as unoccupied until a villager is explicitly spawned.
func TestEachHouseTrackedInOccupancy(t *testing.T) {
	g := makeHouseGame()

	origin1 := geom.Point{X: 5, Y: 5}
	buildHouseAt(g, origin1.X, origin1.Y, 4, 5)
	if g.Villagers.Count() != 0 {
		t.Fatalf("after first house: villager count = %d, want 0", g.Villagers.Count())
	}
	if g.State.HouseOccupancy[origin1] {
		t.Error("first house should be unoccupied after build")
	}

	origin2 := geom.Point{X: 20, Y: 20}
	buildHouseAt(g, origin2.X, origin2.Y, 19, 20)
	if g.Villagers.Count() != 0 {
		t.Fatalf("after second house: villager count = %d, want 0", g.Villagers.Count())
	}
	if g.State.HouseOccupancy[origin2] {
		t.Error("second house should be unoccupied after build")
	}

	unoccupied := 0
	for _, occupied := range g.State.HouseOccupancy {
		if !occupied {
			unoccupied++
		}
	}
	if unoccupied != 2 {
		t.Errorf("expected 2 unoccupied houses, got %d", unoccupied)
	}
}
