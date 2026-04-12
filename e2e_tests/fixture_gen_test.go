// Package e2e_tests contains end-to-end tests that drive the full game stack
// through the bubbletea model — the same code path as a real user.
package e2e_tests

import (
	"flag"
	"testing"

	"forester/game"
	_ "forester/game/resources"
	"forester/game/structures"
	_ "forester/game/upgrades"
)

var updateFixtures = flag.Bool("update-fixtures", false, "regenerate testdata fixture files")

// TestGenerateFixtures builds game state to well-known checkpoints and writes
// them as JSON to e2e_tests/testdata/. Skipped by default; run with:
//
//	go test -run TestGenerateFixtures -update-fixtures ./e2e_tests/
//
// Fixtures capture the complete serialisable game state (world, player, story
// beats, XP, etc.) so individual tests can skip setup ticks they don't care about.
//
// Checkpoints produced:
//
//	checkpoint_log_storage — LogStorage built, MaxCarry=100 (carry upgrade accepted),
//	                         player at (48,45) facing north.
//	checkpoint_house       — First house built, 1 villager spawned,
//	                         player at (53,41) after east-side harvest sweep.
func TestGenerateFixtures(t *testing.T) {
	if !*updateFixtures {
		t.Skip("run with -update-fixtures to regenerate")
	}

	g, clock, m := newTestGame()

	// ── Navigate to harvest position (48,45) ─────────────────────────────────
	// Path from start (50,50): west×2 → (48,50), north×5 → (48,45).
	for _, dir := range []string{"a", "a", "w", "w", "w", "w", "w"} {
		moveDir(&m, clock, g, dir)
	}
	if g.State.Player.TileX() != 48 || g.State.Player.TileY() != 45 {
		t.Fatalf("navigate: expected player at (48,45), got (%d,%d)",
			g.State.Player.TileX(), g.State.Player.TileY())
	}

	// ── Build log storage ─────────────────────────────────────────────────────
	// Player harvests from the forest arc to the north and auto-deposits into
	// the foundation. Log storage completes after 20 deposits (~120–200 ticks).
	const maxLogStorageTicks = 200
	for i := range maxLogStorageTicks {
		tick(&m, clock)
		if g.State.World.HasStructureOfType(structures.LogStorage) {
			break
		}
		if i == maxLogStorageTicks-1 {
			t.Fatal("log storage not built within expected ticks")
		}
	}

	// One extra tick: first_log_storage_built story beat fires → carry upgrade offer queued.
	tick(&m, clock)
	if !g.HasPendingOffer() {
		t.Fatal("carry upgrade offer not queued after log storage built")
	}
	g.SelectCard(0) // carry_capacity: MaxCarry 20 → 100
	if g.State.Player.MaxCarry != 100 {
		t.Fatalf("MaxCarry = %d, want 100 after carry upgrade", g.State.Player.MaxCarry)
	}

	// ── checkpoint_log_storage ────────────────────────────────────────────────
	// State: LogStorage built, MaxCarry=100, player at (48,45).
	writeFixture(t, "checkpoint_log_storage", g)

	// ── Harvest north to fill MaxCarry ────────────────────────────────────────
	// Walk north 5 steps from (48,45) to (48,40), ticking 40 times each stop.
	const (
		stepsNorth   = 5
		ticksPerStep = 40
	)
	for range stepsNorth {
		moveSafe(&m, clock, g, "w") // north
		for range ticksPerStep {
			tickDraining(&m, clock, g)
			if g.State.Player.Inventory[game.Wood] >= g.State.Player.MaxCarry {
				break
			}
		}
	}
	if g.State.Player.Inventory[game.Wood] < houseBuildCost {
		t.Fatalf("need %d wood for house; only have %d", houseBuildCost, g.State.Player.Inventory[game.Wood])
	}

	// ── Return south to deposit position (48,45) ──────────────────────────────
	for range stepsNorth {
		moveSafe(&m, clock, g, "s")
	}
	if g.State.Player.TileX() != 48 || g.State.Player.TileY() != 45 {
		t.Fatalf("return south: expected (48,45), got (%d,%d)",
			g.State.Player.TileX(), g.State.Player.TileY())
	}

	// ── Deposit 50 wood to trigger house foundation ───────────────────────────
	const maxDepositTicks = 200
	for i := range maxDepositTicks {
		tickDraining(&m, clock, g)
		if g.Stores.Total(game.Wood) >= houseSpawnThreshold {
			break
		}
		if i == maxDepositTicks-1 {
			t.Fatalf("stores only %d after %d ticks", g.Stores.Total(game.Wood), maxDepositTicks)
		}
	}

	// ── Navigate to house deposit position (46,51) ────────────────────────────
	for range 2 {
		moveSafe(&m, clock, g, "a") // west
	}
	for range 6 {
		moveSafe(&m, clock, g, "s") // south
	}
	if g.State.Player.TileX() != 46 || g.State.Player.TileY() != 51 {
		t.Fatalf("nav to house: expected (46,51), got (%d,%d)",
			g.State.Player.TileX(), g.State.Player.TileY())
	}

	// ── Build first house ─────────────────────────────────────────────────────
	const maxHouseBuildTicks = 150
	for i := range maxHouseBuildTicks {
		tickDraining(&m, clock, g)
		if g.State.World.HasStructureOfType(structures.House) {
			break
		}
		if i == maxHouseBuildTicks-1 {
			t.Fatalf("house not built after %d ticks", maxHouseBuildTicks)
		}
	}
	// Extra tick: first_house_built story beat fires → 2-card offer queued.
	tick(&m, clock)
	if !g.HasPendingOffer() {
		t.Fatal("house upgrade offer not queued after house built")
	}
	g.SelectCard(0) // build_speed upgrade

	// ── Spawn first villager ──────────────────────────────────────────────────
	g.State.AddOffer([]string{"spawn_villager"})
	g.SelectCard(0)
	if g.Villagers.Count() != 1 {
		t.Fatalf("expected 1 villager, got %d", g.Villagers.Count())
	}

	// ── Harvest east belt to replenish wood ───────────────────────────────────
	// Route from (46,51): north×1→(46,50), east×7→(53,50), north×6→(53,44).
	moveSafe(&m, clock, g, "w") // north → (46,50)
	for range 7 {
		moveSafe(&m, clock, g, "d") // east → (53,50)
	}
	for range 6 {
		moveSafe(&m, clock, g, "w") // north → (53,44)
	}
	if g.State.Player.TileX() != 53 || g.State.Player.TileY() != 44 {
		t.Fatalf("east harvest start: expected (53,44), got (%d,%d)",
			g.State.Player.TileX(), g.State.Player.TileY())
	}
	for range 15 {
		tickDraining(&m, clock, g)
	}
	for range 3 {
		moveSafe(&m, clock, g, "w") // north
		for range 15 {
			tickDraining(&m, clock, g)
		}
	}
	if g.State.Player.TileX() != 53 || g.State.Player.TileY() != 41 {
		t.Fatalf("east harvest end: expected (53,41), got (%d,%d)",
			g.State.Player.TileX(), g.State.Player.TileY())
	}

	// ── checkpoint_house ──────────────────────────────────────────────────────
	// State: 1st house built, 1 villager spawned, player at (53,41).
	writeFixture(t, "checkpoint_house", g)
}
