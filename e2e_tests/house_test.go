// Package e2e_tests contains end-to-end tests that drive the full game stack
// through the bubbletea model — the same code path as a real user.
package e2e_tests

import (
	"fmt"
	"math/rand"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"forester/game"
	_ "forester/game/resources"
	_ "forester/game/structures"
	_ "forester/game/upgrades"
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
//  6. Navigate west×2, south×6 to (46,51).
//  7. Tick until the house is built (50 wood deposited into the foundation).
//  8. Accept the house's 2-card upgrade offer and verify the effect.
//
// World layout facts for seed 42:
//   - Player starts at (50,50). Log storage foundation spawns at (48,46)–(51,49).
//   - After Phase 3 north harvest: fresh trees at y=43–39 provide 100+ wood total.
//   - After Phase 5 deposits 50 into log storage, player.Wood = 50.
//   - Phase 7 tick 1: house foundation spawns at (47,51)–(48,52), the first valid
//     2×2 grassland area found by spiralSearchDo from anchor (49,49) with a 1-tile
//     gap from the log storage. Player at (46,51) is adjacent east to (47,51) and
//     drives the per-tick deposits.
//   - The 2-card offer after house completion: card 0 = "Faster Construction"
//     (build_speed, reduces BuildInterval by 10%), card 1 = "Faster Depositing"
//     (deposit_speed, reduces DepositInterval by 10%).
//
// houseBuildCost and houseSpawnThreshold mirror the values in game/structures/house.go.
// They are duplicated here (rather than exported) because game/structures' constants
// are intentionally package-private. Update both if the gameplay values change.
const (
	houseBuildCost      = 50
	houseSpawnThreshold = 50
)

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
	// the foundation. Log storage completes after 20 deposits. One extra tick is
	// then needed for the first_log_storage_built story beat to queue the offer.
	announcePhase(m, "Phase 2: Build log storage")
	const maxLogStorageTicks = 200
	for i := range maxLogStorageTicks {
		tick(&m, clock)
		if g.State.World.HasStructureOfType(game.LogStorage) {
			break
		}
		if i == maxLogStorageTicks-1 {
			t.Fatal("phase 2: log storage not built within expected ticks")
		}
	}
	// Extra tick: first_log_storage_built story beat fires → carry upgrade offer queued.
	tick(&m, clock)
	if !g.HasPendingOffer() {
		t.Fatal("phase 2: expected carry upgrade offer after log storage built")
	}
	g.SelectCard(0) // carry_capacity: MaxCarry 20 → 100
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
			if g.State.Player.Inventory[game.Wood] >= g.State.Player.MaxCarry {
				break
			}
		}
	}
	if g.State.Player.Inventory[game.Wood] < houseBuildCost {
		t.Fatalf("phase 3: Wood = %d after north harvest, need at least %d for house build",
			g.State.Player.Inventory[game.Wood], houseBuildCost)
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
		if g.Stores.Total(game.Wood) >= houseSpawnThreshold {
			break
		}
		if i == maxDepositTicks-1 {
			t.Fatalf("phase 5: stores only %d after %d ticks; player.Wood=%d",
				g.Stores.Total(game.Wood), maxDepositTicks, g.State.Player.Inventory[game.Wood])
		}
	}
	woodForHouse := g.State.Player.Inventory[game.Wood]
	if woodForHouse < houseBuildCost {
		t.Fatalf("phase 5: need %d wood to build house, only have %d (stores=%d)",
			houseBuildCost, woodForHouse, g.Stores.Total(game.Wood))
	}
	// House foundation should NOT have spawned yet (spawns on Phase 7 tick 1).
	if g.State.World.HasStructureOfType(game.FoundationHouse) {
		t.Error("phase 5: house foundation appeared before navigation; expected it to spawn on Phase 7 tick 1")
	}

	// ── Phase 6: Navigate to (46,51) ─────────────────────────────────────────
	// Route: west×2 to (46,45), south×6 to (46,51).
	// At (46,51) the player is adjacent east to (47,51) [house foundation] but
	// is NOT adjacent to the log storage (nearest log storage tile (48,49) is
	// 2 tiles east and 2 tiles north). This ensures Phase 7 ticks deposit only into the house
	// foundation. moveSafe handles any terrain cooldown transitions.
	// No ticks during navigation: player.Wood stays at woodForHouse.
	announcePhase(m, "Phase 6: Navigate to (46,51)")
	for range 2 {
		moveSafe(&m, clock, g, "a") // west
	}
	for range 6 {
		moveSafe(&m, clock, g, "s") // south
	}
	if g.State.Player.X != 46 || g.State.Player.Y != 51 {
		t.Fatalf("phase 6: expected player at (46,51), got (%d,%d)",
			g.State.Player.X, g.State.Player.Y)
	}

	// ── Phase 7: Build house ──────────────────────────────────────────────────
	// First tick fires Harvest → maybeAdvanceStory fires initial_house story beat
	// (stores≥50) → spawns house foundation at (47,51) (first valid 2×2 found by
	// spiralSearchDo from anchor (49,49) with 1-tile gap from log storage). Within
	// the same tick, TickAdjacentStructures sees (47,51) adjacent to player at
	// (46,51), fires OnPlayerInteraction, deposits 1 wood.
	// After 50 total build ticks the house is complete. One extra tick then fires the
	// first_house_built story beat → 2-card offer; world condition also spawns the
	// next house foundation since built≥1 and no foundation is pending yet.
	announcePhase(m, "Phase 7: Build house (50 wood deposits)")
	const maxHouseBuildTicks = 150
	for i := range maxHouseBuildTicks {
		tick(&m, clock)
		if g.State.World.HasStructureOfType(game.House) {
			break
		}
		if i == maxHouseBuildTicks-1 {
			t.Fatalf("phase 7: house not built after %d ticks; Wood=%d foundationDeposited=%v hasFoundation=%v",
				maxHouseBuildTicks, g.State.Player.Inventory[game.Wood],
				g.State.FoundationDeposited, g.State.World.HasStructureOfType(game.FoundationHouse))
		}
	}
	// Extra tick: first_house_built story beat fires → 2-card offer queued.
	// The world condition (houseDef.ShouldSpawn) also spawns a new house foundation
	// on this tick since built≥1 and no pending foundation exists yet.
	tick(&m, clock)

	// ── Phase 8: Verify house + accept upgrade card ───────────────────────────
	announcePhase(m, "Phase 8: Accept house upgrade card")

	// House must be built.
	if !g.State.World.HasStructureOfType(game.House) {
		t.Fatal("phase 8: House structure not found after build loop")
	}

	// House tile at (47,51) (origin of the foundation) must show House structure.
	houseTile := g.State.World.TileAt(47, 51)
	if houseTile == nil || houseTile.Structure != game.House {
		t.Errorf("phase 8: expected House at (47,51), got %v", houseTile)
	}

	// Offer must be pending with exactly 2 cards.
	if !g.HasPendingOffer() {
		t.Fatal("phase 8: expected 2-card offer after house built")
	}
	offer := g.CurrentOffer()
	if len(offer) != 2 {
		t.Fatalf("phase 8: expected 2-card offer, got %d card(s)", len(offer))
	}

	// Selecting card 0 (build_speed) must reduce BuildInterval by 10%.
	buildIntervalBefore := g.State.Player.BuildInterval
	g.SelectCard(0)
	if g.State.Player.BuildInterval >= buildIntervalBefore {
		t.Errorf("phase 8: BuildInterval should decrease; was %v, got %v",
			buildIntervalBefore, g.State.Player.BuildInterval)
	}
	if g.HasPendingOffer() {
		t.Error("phase 8: offer should be cleared after SelectCard")
	}

	// ── Phase 9: Verify 1st villager + harvest wood for 2nd house ────────────
	// After the 1st house is built, 1 villager must have spawned from OnBuilt.
	// Harvest from the fresh forest belt at x=53 (completely untouched — Phase 3
	// only swept x=47–49). Route: north×1→(46,50), east×7→(53,50), north×6→(53,44),
	// then sweep north 3 more positions ticking 15 times each.
	// Fresh tile count: 4+3×3=13 tiles × min tree size 4 = 52 wood guaranteed.
	announcePhase(m, "Phase 9: Verify 1st villager, harvest wood for 2nd house")
	if g.Villagers.Count() != 1 {
		t.Errorf("phase 9: expected 1 villager after 1st house built, got %d", g.Villagers.Count())
	}

	moveSafe(&m, clock, g, "w") // north → (46,50): clear of 1st house (y=51+)
	for range 7 {
		moveSafe(&m, clock, g, "d") // east → (53,50): below log storage (y=46–49) and house
	}
	for range 6 {
		moveSafe(&m, clock, g, "w") // north → (53,44): east of log storage (x=48–51)
	}
	if g.State.Player.X != 53 || g.State.Player.Y != 44 {
		t.Fatalf("phase 9: expected player at (53,44), got (%d,%d)",
			g.State.Player.X, g.State.Player.Y)
	}
	// Harvest at (53,44) then sweep 3 more positions north; 15 ticks each position.
	const woodFor2ndHouse = houseBuildCost*9/10 + 1 // 46 (>90% of 50)
	for range 15 {
		tick(&m, clock)
	}
	for range 3 {
		moveSafe(&m, clock, g, "w") // north to next fresh arc
		for range 15 {
			tick(&m, clock)
		}
	}
	// Player is now at (53,41) after 3 north moves from (53,44).
	if g.State.Player.X != 53 || g.State.Player.Y != 41 {
		t.Fatalf("phase 9: expected player at (53,41), got (%d,%d)",
			g.State.Player.X, g.State.Player.Y)
	}
	if g.State.Player.Inventory[game.Wood] < woodFor2ndHouse {
		t.Fatalf("phase 9: only harvested %d wood; need %d (≥46)",
			g.State.Player.Inventory[game.Wood], woodFor2ndHouse)
	}

	// ── Phase 10: Navigate to 2nd house foundation deposit position (49,51) ──────
	// The 2nd house foundation spawned at (50,51) on the extra tick after the 1st house was
	// built (houseDef.ShouldSpawn: built≥1 && pending==0; spiral from anchor (49,49) finds
	// (50,51) as the first valid 2×2 not bordering the log storage or 1st house).
	// Deposit position (49,51) is directly west of the origin (50,51): adjacent to the
	// foundation but NOT adjacent to the log storage (nearest storage tile (49,49) is 2 north).
	// Route from (53,41): south×9→(53,50), west×4→(49,50), south×1→(49,51).
	announcePhase(m, "Phase 10: Navigate to 2nd foundation deposit position (49,51)")
	for range 9 {
		moveSafe(&m, clock, g, "s") // south
	}
	for range 4 {
		moveSafe(&m, clock, g, "a") // west
	}
	moveSafe(&m, clock, g, "s") // south → (49,51)
	if g.State.Player.X != 49 || g.State.Player.Y != 51 {
		t.Fatalf("phase 10: expected player at (49,51), got (%d,%d)",
			g.State.Player.X, g.State.Player.Y)
	}
	// Verify the 2nd foundation is directly east of the player at (50,51).
	if tile := g.State.World.TileAt(50, 51); tile == nil || tile.Structure != game.FoundationHouse {
		t.Fatalf("phase 10: expected FoundationHouse at (50,51), got %v", tile)
	}

	// ── Phase 11: Player deposits >90% into 2nd foundation, then stops ─────────
	// Player at (49,51) is cardinally adjacent east to the 2nd foundation origin (50,51).
	// Each tick deposits 1 wood via OnPlayerInteraction (Build cooldown ≈ 90ms < GameTickInterval).
	// Break once foundation progress exceeds 90% (>45 of 50 deposited). The villager may also
	// contribute during these ticks, which only accelerates completion.
	announcePhase(m, "Phase 11: Player deposits >90% into 2nd foundation, then stops")
	const maxBuild2Ticks = 120
	for i := range maxBuild2Ticks {
		tick(&m, clock)
		// Check whether the foundation is still in progress and past the 90% threshold.
		// Note: FoundationProgress returns (0, false) before the first deposit; isBuilt
		// guards against that case so we never break prematurely on an untouched foundation.
		isBuilt := g.State.World.CountStructureInstances(game.House) >= 2
		progress, hasPending := g.State.FoundationProgress()
		if isBuilt || (hasPending && progress > 0.9) {
			break
		}
		if i == maxBuild2Ticks-1 {
			t.Fatalf("phase 11: foundation not >90%% complete after %d ticks; progress=%.2f",
				maxBuild2Ticks, progress)
		}
	}

	// ── Phase 12: Player steps back; villager delivers the remaining wood ─────────
	// If the 2nd house isn't fully built yet (player stopped after >90%), move the player
	// north to (49,50) — no longer adjacent to the foundation — and let the villager fetch
	// the remaining wood from the log storage to complete the build.
	if g.State.World.CountStructureInstances(game.House) < 2 {
		announcePhase(m, "Phase 12: Player steps back to (49,50); villager completes the 2nd house")
		moveSafe(&m, clock, g, "w") // north → (49,50); adjacent to log storage, not foundation
		if g.State.Player.X != 49 || g.State.Player.Y != 50 {
			t.Fatalf("phase 12: expected player at (49,50), got (%d,%d)",
				g.State.Player.X, g.State.Player.Y)
		}
		const maxVillagerBuildTicks = 200
		for i := range maxVillagerBuildTicks {
			tick(&m, clock)
			if g.State.World.CountStructureInstances(game.House) >= 2 {
				break
			}
			if i == maxVillagerBuildTicks-1 {
				t.Fatalf("phase 12: villager did not complete 2nd house after %d ticks", maxVillagerBuildTicks)
			}
		}
	}

	// ── Phase 13: Verify 2nd house built + 2nd villager spawned ───────────────
	announcePhase(m, "Phase 13: Verify 2nd house built and 2nd villager spawned")
	if g.State.World.CountStructureInstances(game.House) < 2 {
		t.Fatal("phase 13: 2nd house not built")
	}
	if g.Villagers.Count() != 2 {
		t.Errorf("phase 13: expected 2 villagers after 2nd house built, got %d", g.Villagers.Count())
	}

	announcePhase(m, fmt.Sprintf("Done — 2 houses built, 2 villagers spawned! BuildInterval: %v",
		g.State.Player.BuildInterval))
}
