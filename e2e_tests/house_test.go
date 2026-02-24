// Package e2e_tests contains end-to-end tests that drive the full game stack
// through the bubbletea model — the same code path as a real user.
package e2e_tests

import (
	"fmt"
	"math/rand"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"forester/game"
	"forester/render"
)

// moveSafe advances the clock by the greater of the current tile's move cooldown
// or the remaining move cooldown, then sends the directional key. This correctly
// handles Forest→Grassland transitions: after moving off a Forest tile (300ms
// cooldown set), the subsequent Grassland move (only 150ms) would fail with
// moveDir because the previous 300ms cooldown hasn't expired.
func moveSafe(m *render.Model, clock *game.FakeClock, g *game.Game, dir string) {
	p := g.State.Player
	tile := g.State.World.TileAt(p.X, p.Y)
	needed := game.MoveCooldownFor(tile)
	remaining := p.Cooldowns[game.Move].Sub(clock.Now())
	if remaining > needed {
		clock.Advance(remaining)
	} else {
		clock.Advance(needed)
	}
	sendKey(m, dir)
	renderFrame(*m, fmt.Sprintf("moveSafe %s → (%d,%d)", dir, g.State.Player.X, g.State.Player.Y))
}

// TestHouseWorkflow is a full end-to-end scenario for the house building path:
//
//  1. Navigate to harvest position (48,45) adjacent to forest.
//  2. Tick until log storage is built and accept the MaxCarry upgrade.
//  3. Harvest wood going north in 5 steps to fill MaxCarry (100 wood).
//  4. Return south to (48,45) adjacent to the log storage.
//  5. Deposit 50 wood into storage to trigger the house foundation spawn.
//  6. Navigate east×4, south×5, west×1, south×1 to (51,51).
//  7. Tick until the house is built (50 wood deposited into the foundation).
//  8. Accept the house's 2-card upgrade offer and verify the effect.
//
// World layout facts for seed 42:
//   - Player starts at (50,50). Log storage foundation spawns at (48,46)–(51,49).
//   - After Phase 3 north harvest: fresh trees at y=43–39 provide 100+ wood total.
//   - After Phase 5 deposits 50 into log storage, player.Wood = 50.
//   - Phase 7 tick 1: house foundation spawns at (49,51)–(50,52), the closest valid
//     2×2 grassland area to spawn (50,50) with a full 1-tile gap from the log storage.
//     Player at (50,50) is adjacent to (50,51) and drives the per-tick deposits.
//   - The 2-card offer after house completion: card 0 = "Faster Construction"
//     (build_speed, reduces BuildInterval by 10%), card 1 = "Faster Depositing"
//     (deposit_speed, reduces DepositInterval by 10%).
func TestHouseWorkflow(t *testing.T) {
	// ── Setup ────────────────────────────────────────────────────────────────
	clock := game.NewFakeClock()
	g := game.NewWithClockAndRNG(clock, rand.New(rand.NewSource(42)))
	m := render.NewModelWithClock(g, clock)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(render.Model)

	// ── Phase 1: Navigate to harvest position (48,45) ────────────────────────
	// west×2 from (50,50) → (48,50), north×5 → (48,45).
	// All tiles are Grassland or Forest; moveSafe handles any terrain transition.
	announcePhase(m, "Phase 1: Navigate to harvest position (48,45)")
	for _, dir := range []string{"a", "a", "w", "w", "w", "w", "w"} {
		moveSafe(&m, clock, g, dir)
	}
	if g.State.Player.X != 48 || g.State.Player.Y != 45 {
		t.Fatalf("phase 1: expected player at (48,45), got (%d,%d)",
			g.State.Player.X, g.State.Player.Y)
	}

	// ── Phase 2: Build log storage + accept MaxCarry upgrade ─────────────────
	// Player harvests from the 4-tile arc facing north and auto-deposits into
	// the foundation. Log storage completes after 20 deposits; offer queued.
	announcePhase(m, "Phase 2: Build log storage")
	const maxLogStorageTicks = 200
	for i := range maxLogStorageTicks {
		tick(&m, clock)
		if g.State.HasStructureOfType(game.LogStorage) {
			break
		}
		if i == maxLogStorageTicks-1 {
			t.Fatal("phase 2: log storage not built within expected ticks")
		}
	}
	if !g.State.HasPendingOffer() {
		t.Fatal("phase 2: expected carry upgrade offer after log storage built")
	}
	g.State.SelectCard(0) // carry_capacity: MaxCarry 20 → 100
	if g.State.Player.MaxCarry != 100 {
		t.Errorf("phase 2: MaxCarry = %d, want 100 after accepting carry upgrade",
			g.State.Player.MaxCarry)
	}

	// ── Phase 3: Harvest wood going north to fill MaxCarry ───────────────────
	// Walk north 5 steps (48,45)→(48,44)→…→(48,40), ticking exhaustively at
	// each position. The fresh arc at each step provides 16–29 new wood.
	// With player.Wood≈17 at start and ~102 fresh wood available across 5 stops,
	// MaxCarry (100) is reached partway through step 5.
	announcePhase(m, "Phase 3: Harvest wood (north route to fill MaxCarry=100)")
	const stepsNorth = 5
	const ticksPerStep = 40
	for range stepsNorth {
		moveSafe(&m, clock, g, "w") // north
		for range ticksPerStep {
			tick(&m, clock)
			if g.State.Player.Wood >= g.State.Player.MaxCarry {
				break
			}
		}
	}
	if g.State.Player.Wood < game.HouseBuildCost {
		t.Fatalf("phase 3: Wood = %d after north harvest, need at least %d for house build",
			g.State.Player.Wood, game.HouseBuildCost)
	}
	// Player should be at (48,40) after 5 north steps from (48,44 start).
	if g.State.Player.X != 48 || g.State.Player.Y != 40 {
		t.Fatalf("phase 3: expected player at (48,40), got (%d,%d)",
			g.State.Player.X, g.State.Player.Y)
	}

	// ── Phase 4: Return south to deposit position (48,45) ───────────────────
	announcePhase(m, "Phase 4: Return south to (48,45)")
	for range stepsNorth {
		moveSafe(&m, clock, g, "s") // south
	}
	if g.State.Player.X != 48 || g.State.Player.Y != 45 {
		t.Fatalf("phase 4: expected player at (48,45), got (%d,%d)",
			g.State.Player.X, g.State.Player.Y)
	}

	// ── Phase 5: Deposit 50 wood to trigger house foundation ─────────────────
	// Player.Wood ≈ 100 (MaxCarry). Adjacent to log storage to the south.
	// Each tick deposits 1 wood into the storage (no nearby trees, so no harvest).
	// After 50 deposits: stores = 50 ≥ HouseSpawnThreshold → foundation will
	// spawn on the NEXT tick's Harvest phase (not this tick's, since Harvest
	// runs before TickAdjacentStructures).
	// Player.Wood at break = 100 - 50 = 50 (enough to build the house).
	announcePhase(m, "Phase 5: Deposit 50 wood to trigger house foundation")
	const maxDepositTicks = 200
	for i := range maxDepositTicks {
		tick(&m, clock)
		if g.Stores.Total(game.Wood) >= game.HouseSpawnThreshold {
			break
		}
		if i == maxDepositTicks-1 {
			t.Fatalf("phase 5: stores only %d after %d ticks; player.Wood=%d",
				g.Stores.Total(game.Wood), maxDepositTicks, g.State.Player.Wood)
		}
	}
	woodForHouse := g.State.Player.Wood
	if woodForHouse < game.HouseBuildCost {
		t.Fatalf("phase 5: need %d wood to build house, only have %d (stores=%d)",
			game.HouseBuildCost, woodForHouse, g.Stores.Total(game.Wood))
	}
	// House foundation should NOT have spawned yet (spawns on Phase 7 tick 1).
	if g.State.HasStructureOfType(game.FoundationHouse) {
		t.Error("phase 5: house foundation appeared before navigation; expected it to spawn on Phase 7 tick 1")
	}

	// ── Phase 6: Navigate to (51,51) ─────────────────────────────────────────
	// Route: east×4 through mixed terrain to (52,45), south×5 to (52,50) —
	// x=52 bypasses the log storage footprint (x=48–51) — west×1 to (51,50),
	// south×1 to (51,51).
	// At (51,51) the player is adjacent west to (50,51) [house foundation] but
	// is NOT adjacent to the log storage (nearest tile (51,49) is 2 tiles north).
	// This ensures Phase 7 ticks deposit only into the house foundation.
	// moveSafe handles the Forest→Grassland cooldown transitions along this route.
	// No ticks during navigation: player.Wood stays at woodForHouse.
	announcePhase(m, "Phase 6: Navigate to (51,51)")
	for range 4 {
		moveSafe(&m, clock, g, "d") // east
	}
	for range 5 {
		moveSafe(&m, clock, g, "s") // south
	}
	moveSafe(&m, clock, g, "a") // west
	moveSafe(&m, clock, g, "s") // south
	if g.State.Player.X != 51 || g.State.Player.Y != 51 {
		t.Fatalf("phase 6: expected player at (51,51), got (%d,%d)",
			g.State.Player.X, g.State.Player.Y)
	}

	// ── Phase 7: Build house ──────────────────────────────────────────────────
	// First tick fires Harvest → maybeSpawnFoundation sees stores≥50 → spawns
	// house foundation at (49,51) (closest valid 2×2 to spawn with 1-tile gap from
	// log storage). Within the same tick, TickAdjacentStructures sees (50,51)
	// adjacent to player at (51,51), fires OnPlayerInteraction, deposits 1 wood.
	// After 50 total ticks: HouseBuildCost deposits complete → house built → 2-card offer.
	announcePhase(m, "Phase 7: Build house (50 wood deposits)")
	const maxHouseBuildTicks = 150
	for i := range maxHouseBuildTicks {
		tick(&m, clock)
		if g.State.HasStructureOfType(game.House) || g.State.HasPendingOffer() {
			break
		}
		if i == maxHouseBuildTicks-1 {
			t.Fatalf("phase 7: house not built after %d ticks; Wood=%d foundationDeposited=%v hasFoundation=%v",
				maxHouseBuildTicks, g.State.Player.Wood,
				g.State.FoundationDeposited, g.State.HasStructureOfType(game.FoundationHouse))
		}
	}

	// ── Phase 8: Verify house + accept upgrade card ───────────────────────────
	announcePhase(m, "Phase 8: Accept house upgrade card")

	// House must be built.
	if !g.State.HasStructureOfType(game.House) {
		t.Fatal("phase 8: House structure not found after build loop")
	}
	if g.State.HasStructureOfType(game.FoundationHouse) {
		t.Error("phase 8: FoundationHouse tiles should be gone after house is built")
	}

	// House tile at (49,51) (origin of the foundation) must show House structure.
	houseTile := g.State.World.TileAt(49, 51)
	if houseTile == nil || houseTile.Structure != game.House {
		t.Errorf("phase 8: expected House at (49,51), got %v", houseTile)
	}

	// Offer must be pending with exactly 2 cards.
	if !g.State.HasPendingOffer() {
		t.Fatal("phase 8: expected 2-card offer after house built")
	}
	offer := g.State.CurrentOffer()
	if len(offer) != 2 {
		t.Fatalf("phase 8: expected 2-card offer, got %d card(s)", len(offer))
	}

	// Selecting card 0 (build_speed) must reduce BuildInterval by 10%.
	buildIntervalBefore := g.State.Player.BuildInterval
	g.State.SelectCard(0)
	if g.State.Player.BuildInterval >= buildIntervalBefore {
		t.Errorf("phase 8: BuildInterval should decrease; was %v, got %v",
			buildIntervalBefore, g.State.Player.BuildInterval)
	}
	if g.State.HasPendingOffer() {
		t.Error("phase 8: offer should be cleared after SelectCard")
	}

	announcePhase(m, fmt.Sprintf("Done — House built! BuildInterval: %v → %v",
		buildIntervalBefore, g.State.Player.BuildInterval))
}
