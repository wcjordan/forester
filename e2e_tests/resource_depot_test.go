package e2e_tests

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"forester/game"
	"forester/game/geom"
	_ "forester/game/resources"
	"forester/game/structures"
	_ "forester/game/upgrades"
	"forester/render"
)

// E2E build-cost constants mirrored from game/structures (package-private there).
const (
	e2eLogStorageBuildCost = 20
	e2eHouseBuildCost      = 50
	e2eDepotBuildCost      = 800
	e2eHouseSpawnThreshold = 50 // wood stored to trigger initial_house beat
)

// firstOriginOf returns the NW-anchor point of the first structure instance of
// the given type. Used to find dynamically-placed foundations in E2E tests.
func firstOriginOf(g *game.Game, stype game.StructureType) geom.Point {
	for pt := range g.State.World.StructureTypeIndex[stype] {
		if g.State.World.IsStructureOrigin(pt.X, pt.Y) {
			return pt
		}
	}
	panic("firstOriginOf: no " + string(stype) + " structure found")
}

// TestResourceDepotWorkflow drives the full resource-depot progression path:
//
//  1. Trigger log-storage beat by filling player inventory → foundation spawns.
//  2. Build log storage (20 wood) → carry upgrade queued → accept it (MaxCarry=100).
//  3. Deposit 50 wood into storage → initial_house beat fires → house 1 foundation spawns.
//  4. Build houses 1–4 (each auto-spawned after previous built).
//  5. After house 4: beat 500 fires → FoundationResourceDepot spawns near center.
//  6. Build depot (800 wood) → beat 600 fires → large_carry_capacity offer queued.
//  7. Accept card → verify MaxCarry increased by 100.
func TestResourceDepotWorkflow(t *testing.T) {
	clock := game.NewFakeClock()
	g := game.NewWithClockAndRNG(clock, rand.New(rand.NewSource(42)))
	m := render.NewModelWithClock(g, clock)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(render.Model)

	// ── Phase 1: Trigger initial_log_storage beat ────────────────────────────
	// Beat 100 condition: player.Inventory[Wood] >= MaxCarry.
	// Move player off world-center first: findValidLocationNearPlayer walks from
	// the player toward the center, so the player must not be sitting on the center.
	announcePhase(m, "Phase 1: Trigger log storage beat")
	g.State.Player.X = 45
	g.State.Player.Y = 50
	g.State.Player.Inventory[game.Wood] = g.State.Player.MaxCarry
	tick(&m, clock) // Harvest (no-op) → beat 100 → foundation spawns

	if !g.State.World.HasStructureOfType(structures.FoundationLogStorage) {
		t.Fatal("phase 1: FoundationLogStorage not spawned after filling inventory")
	}

	// ── Phase 2: Build log storage ───────────────────────────────────────────
	announcePhase(m, "Phase 2: Build log storage")
	lsOrigin := firstOriginOf(g, structures.FoundationLogStorage)
	g.State.Player.X = lsOrigin.X - 1
	g.State.Player.Y = lsOrigin.Y
	g.State.Player.Inventory[game.Wood] = e2eLogStorageBuildCost
	g.State.Player.SetCooldown(game.Build, time.Time{})

	const maxLogStorageTicks = 60
	for i := range maxLogStorageTicks {
		tickDraining(&m, clock, g)
		if g.State.World.HasStructureOfType(structures.LogStorage) {
			break
		}
		if i == maxLogStorageTicks-1 {
			t.Fatalf("phase 2: log storage not built after %d ticks", maxLogStorageTicks)
		}
	}
	// Beat 200 fires on next tick → carry_capacity offer queued.
	tick(&m, clock)
	if !g.HasPendingOffer() {
		t.Fatal("phase 2: carry_capacity offer not queued after log storage built")
	}
	g.SelectCard(0) // carry_capacity: MaxCarry 20 → 100
	if g.State.Player.MaxCarry != 100 {
		t.Errorf("phase 2: MaxCarry = %d, want 100", g.State.Player.MaxCarry)
	}

	// ── Phase 3: Deposit 50 wood to trigger initial_house beat ───────────────
	// Beat 300 condition: stores.Total(Wood) >= 50.
	announcePhase(m, "Phase 3: Deposit 50 wood to trigger house beat")
	g.State.Player.X = lsOrigin.X - 1
	g.State.Player.Y = lsOrigin.Y
	g.State.Player.Inventory[game.Wood] = e2eHouseSpawnThreshold
	g.State.Player.SetCooldown(game.Deposit, time.Time{})

	const maxDepositTicks = 100
	for i := range maxDepositTicks {
		tickDraining(&m, clock, g)
		if g.Stores.Total(game.Wood) >= e2eHouseSpawnThreshold {
			break
		}
		if i == maxDepositTicks-1 {
			t.Fatalf("phase 3: stores only %d after %d ticks", g.Stores.Total(game.Wood), maxDepositTicks)
		}
	}
	// Beat 300 fires on next tick → house 1 foundation spawns near world center.
	tick(&m, clock)
	drainOffers(g)

	// ── Phases 4–7: Build houses 1 through 4 ────────────────────────────────
	// Each house auto-spawns after the previous one is built (houseDef.ShouldSpawn).
	// The new foundation appears on the tick AFTER FinalizeFoundation runs, so the
	// wait loop below gives the game one tick to spawn it.
	for houseNum := 1; houseNum <= 4; houseNum++ {
		announcePhase(m, fmt.Sprintf("Phase %d: Build house %d", houseNum+3, houseNum))

		// Wait for the house foundation to exist (may need 1 extra tick to spawn).
		const maxFoundWait = 20
		foundationFound := false
		for i := range maxFoundWait {
			if g.State.World.HasStructureOfType(structures.FoundationHouse) {
				foundationFound = true
				break
			}
			tickDraining(&m, clock, g)
			if i == maxFoundWait-1 {
				t.Fatalf("house %d: FoundationHouse not spawned after %d ticks", houseNum, maxFoundWait)
			}
		}
		if !foundationFound {
			t.Fatalf("house %d: FoundationHouse not spawned", houseNum)
		}

		hOrigin := firstOriginOf(g, structures.FoundationHouse)
		g.State.Player.X = hOrigin.X - 1
		g.State.Player.Y = hOrigin.Y
		g.State.Player.Inventory[game.Wood] = e2eHouseBuildCost
		g.State.Player.SetCooldown(game.Build, time.Time{})
		// Lock deposit cooldown so adjacent log storage doesn't consume build wood.
		g.State.Player.SetCooldown(game.Deposit, clock.Now().Add(time.Hour))

		housesBefore := g.State.World.CountStructureInstances(structures.House)
		const maxHouseBuildTicks = 150
		for i := range maxHouseBuildTicks {
			tickDraining(&m, clock, g)
			if g.State.World.CountStructureInstances(structures.House) > housesBefore {
				break
			}
			if i == maxHouseBuildTicks-1 {
				t.Fatalf("house %d: not built after %d ticks; deposited=%d",
					houseNum, maxHouseBuildTicks, g.State.FoundationDeposited[hOrigin])
			}
		}
		drainOffers(g)
	}

	if g.State.World.CountStructureInstances(structures.House) != 4 {
		t.Fatalf("expected 4 houses built, got %d", g.State.World.CountStructureInstances(structures.House))
	}

	// ── Phase 8: Verify depot foundation spawns after 4 houses ───────────────
	// Beat 500 condition: 4 houses AND no depot pending/built.
	// The beat fires during maybeAdvanceStory on the next tick.
	announcePhase(m, "Phase 8: Verify depot foundation spawns")
	const maxDepotFoundWait = 20
	for i := range maxDepotFoundWait {
		tickDraining(&m, clock, g)
		if g.State.World.HasStructureOfType(structures.FoundationResourceDepot) {
			break
		}
		if i == maxDepotFoundWait-1 {
			t.Fatalf("phase 8: FoundationResourceDepot not spawned after 4 houses built (ticks=%d)", maxDepotFoundWait)
		}
	}

	// ── Phase 9: Build depot (800 wood) ──────────────────────────────────────
	announcePhase(m, "Phase 9: Build resource depot (800 wood)")
	depotOrigin := firstOriginOf(g, structures.FoundationResourceDepot)

	// Give player enough wood to cover full build cost (extra buffer ensures
	// the build completes even if some ticks don't deposit due to transitions).
	g.State.Player.MaxCarry = e2eDepotBuildCost + 100
	g.State.Player.Inventory[game.Wood] = e2eDepotBuildCost + 100
	g.State.Player.X = depotOrigin.X - 1
	g.State.Player.Y = depotOrigin.Y
	g.State.Player.SetCooldown(game.Build, time.Time{})

	const maxDepotBuildTicks = 1200
	for i := range maxDepotBuildTicks {
		tickDraining(&m, clock, g)
		if g.State.World.HasStructureOfType(structures.ResourceDepot) {
			break
		}
		if i == maxDepotBuildTicks-1 {
			t.Fatalf("phase 9: depot not built after %d ticks; deposited=%d wood=%d",
				maxDepotBuildTicks, g.State.FoundationDeposited[depotOrigin],
				g.State.Player.Inventory[game.Wood])
		}
	}

	// ── Phase 10: Verify large_carry_capacity offer ───────────────────────────
	// Beat 600 fires during maybeAdvanceStory on the next tick after the depot tile
	// is registered. Use plain tick (not tickDraining) so the offer is preserved.
	announcePhase(m, "Phase 10: Verify and accept large_carry_capacity offer")
	const maxBeat600Wait = 5
	for i := range maxBeat600Wait {
		tick(&m, clock) // plain tick — do NOT drain; we want to inspect the offer
		if g.HasPendingOffer() {
			break
		}
		if i == maxBeat600Wait-1 {
			t.Fatal("phase 10: large_carry_capacity offer not queued after depot built")
		}
	}

	offer := g.CurrentOffer()
	if len(offer) != 1 || offer[0].ID() != "large_carry_capacity" {
		ids := make([]string, len(offer))
		for i, u := range offer {
			ids[i] = u.ID()
		}
		t.Fatalf("phase 10: expected offer [large_carry_capacity], got %v", ids)
	}

	maxCarryBefore := g.State.Player.MaxCarry
	g.SelectCard(0)

	if g.State.Player.MaxCarry != maxCarryBefore+100 {
		t.Errorf("phase 10: MaxCarry = %d, want %d (was %d + 100)",
			g.State.Player.MaxCarry, maxCarryBefore+100, maxCarryBefore)
	}

	announcePhase(m, fmt.Sprintf("Done — depot built, MaxCarry: %d → %d",
		maxCarryBefore, g.State.Player.MaxCarry))
}
