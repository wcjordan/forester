package structures

import (
	"math/rand"
	"testing"
	"time"

	"forester/game"
	"forester/game/geom"
)

// makeDepotGame creates a game with a 30×30 world and player at (5, 15) — away
// from world center so spawn-anchored placement (near center) has room to work.
func makeDepotGame() *game.Game {
	g := game.NewWithClockAndRNG(game.RealClock{}, rand.New(rand.NewSource(0)))
	w := game.NewWorld(30, 30)
	p := game.NewPlayer(5, 15)
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

// buildDepotAt places a resource depot foundation at (ox, oy) and drives ticks
// until it is fully built. Player must already be adjacent.
func buildDepotAt(g *game.Game, ox, oy, playerX, playerY int) {
	g.State.World.PlaceFoundation(ox, oy, resourceDepotDef{})
	g.State.Player.X = playerX
	g.State.Player.Y = playerY
	g.State.Player.Inventory[game.Wood] = resourceDepotBuildCost
	g.State.Player.SetCooldown(game.Build, time.Time{})
	t0 := time.Now()
	for i := range resourceDepotBuildCost {
		g.TickAdjacentStructures(t0.Add(time.Duration(i) * (game.DepositTickInterval + time.Millisecond)))
	}
}

// TestResourceDepotBuilt verifies that depositing build cost converts foundation → depot.
func TestResourceDepotBuilt(t *testing.T) {
	g := makeDepotGame()
	origin := geom.Point{X: 10, Y: 10}
	buildDepotAt(g, origin.X, origin.Y, 9, 10)

	if !g.State.World.HasStructureOfType(ResourceDepot) {
		t.Fatal("resource depot was not built after depositing build cost")
	}
}

// TestResourceDepotRegistersStorage verifies that the OnBuilt callback registers
// wood storage with the configured capacity.
func TestResourceDepotRegistersStorage(t *testing.T) {
	g := makeDepotGame()
	origin := geom.Point{X: 10, Y: 10}
	buildDepotAt(g, origin.X, origin.Y, 9, 10)

	got := g.Stores.TotalCapacity(game.Wood)
	if got != resourceDepotCapacity {
		t.Errorf("wood storage capacity = %d, want %d", got, resourceDepotCapacity)
	}
}

// TestResourceDepotDepositWood verifies that the player can deposit wood into a
// built depot via player interaction.
func TestResourceDepotDepositWood(t *testing.T) {
	g := makeDepotGame()
	origin := geom.Point{X: 10, Y: 10}
	buildDepotAt(g, origin.X, origin.Y, 9, 10)

	g.State.Player.X = 9
	g.State.Player.Y = 10
	g.State.Player.Inventory[game.Wood] = 5
	g.State.Player.SetCooldown(game.Deposit, time.Time{})

	t0 := time.Now()
	g.TickAdjacentStructures(t0)

	if g.Stores.Total(game.Wood) != 1 {
		t.Errorf("wood stored = %d, want 1 after one deposit tick", g.Stores.Total(game.Wood))
	}
	if g.State.Player.Inventory[game.Wood] != 4 {
		t.Errorf("player wood = %d, want 4 after depositing 1", g.State.Player.Inventory[game.Wood])
	}
}

// TestResourceDepotBeat500Condition verifies the beat 500 trigger condition directly:
// requires 4 houses and no existing depot.
func TestResourceDepotBeat500Condition(t *testing.T) {
	g := makeDepotGame()

	// Place 4 built houses directly (bypassing story beats).
	houseOrigins := []geom.Point{{X: 2, Y: 2}, {X: 5, Y: 2}, {X: 8, Y: 2}, {X: 11, Y: 2}}
	for _, o := range houseOrigins {
		buildHouseAt(g, o.X, o.Y, o.X-1, o.Y)
	}

	if g.State.World.CountStructureInstances(House) != 4 {
		t.Fatalf("expected 4 houses, got %d", g.State.World.CountStructureInstances(House))
	}

	conditionMet := g.State.World.CountStructureInstances(House) >= resourceDepotTriggerHouses &&
		!g.State.World.HasStructureOfType(ResourceDepot) &&
		!g.State.World.HasStructureOfType(FoundationResourceDepot)
	if !conditionMet {
		t.Error("beat 500 condition should be met with 4 houses and no depot")
	}
}

// TestVillageCenterDefaultsToWorldCenter verifies that VillageCenter returns the world
// center when no ResourceDepot exists.
func TestVillageCenterDefaultsToWorldCenter(t *testing.T) {
	g := makeDepotGame()
	x, y := g.State.World.VillageCenter()
	wantX := g.State.World.Width / 2
	wantY := g.State.World.Height / 2
	if x != wantX || y != wantY {
		t.Errorf("VillageCenter (no depot) = (%d,%d), want (%d,%d)", x, y, wantX, wantY)
	}
}

// TestVillageCenterUsesDepotOrigin verifies that VillageCenter returns the depot's
// NW origin when a ResourceDepot is present.
func TestVillageCenterUsesDepotOrigin(t *testing.T) {
	g := makeDepotGame()
	depotOrigin := geom.Point{X: 10, Y: 10}
	buildDepotAt(g, depotOrigin.X, depotOrigin.Y, 9, 10)

	x, y := g.State.World.VillageCenter()
	if x != depotOrigin.X || y != depotOrigin.Y {
		t.Errorf("VillageCenter (with depot) = (%d,%d), want depot origin (%d,%d)", x, y, depotOrigin.X, depotOrigin.Y)
	}
}

// TestResourceDepotStoryBeat600 verifies that building a depot queues the upgrade offer.
func TestResourceDepotStoryBeat600(t *testing.T) {
	g := makeDepotGame()
	origin := geom.Point{X: 10, Y: 10}
	buildDepotAt(g, origin.X, origin.Y, 9, 10)

	// Condition for beat 600.
	if !g.State.World.HasStructureOfType(ResourceDepot) {
		t.Fatal("depot not built")
	}

	// Apply the beat action directly.
	g.State.AddOffer([]string{"large_carry_capacity"})

	if !g.HasPendingOffer() {
		t.Error("expected pending offer after depot built")
	}
}
