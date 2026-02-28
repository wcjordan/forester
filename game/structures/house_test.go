package structures

import (
	"testing"
	"time"

	"forester/game"
)

func makeHouseEnv() (*game.State, *game.StorageManager, *game.VillagerManager) {
	w := game.NewWorld(30, 30)
	p := game.NewPlayer(15, 15)
	s := &game.State{
		Player:              p,
		World:               w,
		FoundationDeposited: make(map[game.Point]int),
	}
	return s, game.NewStorageManager(), game.NewVillagerManager()
}

func TestVillagerSpawnsOnHouseBuilt(t *testing.T) {
	s, stores, vm := makeHouseEnv()
	g := &game.Game{State: s, Stores: stores, Villagers: vm}

	// Place a house foundation adjacent to the player and build it.
	origin := game.Point{X: 10, Y: 10}
	s.World.PlaceFoundation(origin.X, origin.Y, houseDef{})

	// Player adjacent to the foundation; give enough wood to build.
	s.Player.X = 9
	s.Player.Y = 10
	s.Player.Inventory[game.Wood] = houseBuildCost

	t0 := time.Now()
	for i := range houseBuildCost {
		g.TickAdjacentStructures(t0.Add(time.Duration(i) * (game.DepositTickInterval + time.Millisecond)))
	}

	if !s.World.HasStructureOfType(game.House) {
		t.Fatal("house was not built after depositing build cost")
	}
	if vm.Count() != 1 {
		t.Errorf("villager count = %d, want 1 after house is built", vm.Count())
	}
}

func TestVillagerSpawnsAdjacentToHouse(t *testing.T) {
	s, stores, vm := makeHouseEnv()
	g := &game.Game{State: s, Stores: stores, Villagers: vm}

	origin := game.Point{X: 10, Y: 10}
	s.World.PlaceFoundation(origin.X, origin.Y, houseDef{})

	s.Player.X = 9
	s.Player.Y = 10
	s.Player.Inventory[game.Wood] = houseBuildCost

	t0 := time.Now()
	for i := range houseBuildCost {
		g.TickAdjacentStructures(t0.Add(time.Duration(i) * (game.DepositTickInterval + time.Millisecond)))
	}

	if vm.Count() == 0 {
		t.Fatal("no villager spawned")
	}

	v := vm.Villagers[0]
	// Villager must not be inside the house footprint.
	insideHouse := v.X >= origin.X && v.X < origin.X+2 && v.Y >= origin.Y && v.Y < origin.Y+2
	if insideHouse {
		t.Errorf("villager spawned inside house footprint at (%d, %d)", v.X, v.Y)
	}

	// Villager must be on a clear tile.
	tile := s.World.TileAt(v.X, v.Y)
	if tile == nil {
		t.Fatalf("villager spawned out of bounds at (%d, %d)", v.X, v.Y)
	}
	if tile.Structure != game.NoStructure {
		t.Errorf("villager spawned on a structure tile at (%d, %d)", v.X, v.Y)
	}
}

func TestEachHouseSpawnsOneVillager(t *testing.T) {
	s, stores, vm := makeHouseEnv()
	g := &game.Game{State: s, Stores: stores, Villagers: vm}

	buildHouse := func(ox, oy, playerX, playerY int) {
		t.Helper()
		s.World.PlaceFoundation(ox, oy, houseDef{})
		s.Player.X = playerX
		s.Player.Y = playerY
		s.Player.Inventory[game.Wood] = houseBuildCost
		// Reset build cooldown so a prior house build doesn't block this one.
		s.Player.SetCooldown(game.Build, time.Time{})
		t0 := time.Now()
		for i := range houseBuildCost {
			g.TickAdjacentStructures(t0.Add(time.Duration(i) * (game.DepositTickInterval + time.Millisecond)))
		}
	}

	buildHouse(5, 5, 4, 5)
	if vm.Count() != 1 {
		t.Fatalf("after first house: villager count = %d, want 1", vm.Count())
	}

	buildHouse(20, 20, 19, 20)
	if vm.Count() != 2 {
		t.Errorf("after second house: villager count = %d, want 2", vm.Count())
	}
}
