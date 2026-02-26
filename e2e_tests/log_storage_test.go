// Package e2e_tests contains end-to-end tests that drive the full game stack
// through the bubbletea model — the same code path as a real user.
package e2e_tests

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"forester/game"
	_ "forester/game/upgrades"
	"forester/render"
)

// ansiRE matches ANSI escape sequences (e.g. colour codes from lipgloss).
var ansiRE = regexp.MustCompile(`\x1b\[[0-9;]*m`)

// stripANSI removes ANSI escape sequences so assertions work on plain text.
func stripANSI(s string) string { return ansiRE.ReplaceAllString(s, "") }

// viewLines returns the stripped View() output split into lines.
func viewLines(m render.Model) []string {
	return strings.Split(stripANSI(m.View()), "\n")
}

// statusBar returns the last line of the view (the status bar).
func statusBar(m render.Model) string {
	lines := viewLines(m)
	return lines[len(lines)-1]
}

// charAtScreen returns the single character at (col, row) in the stripped view.
func charAtScreen(m render.Model, col, row int) string {
	lines := viewLines(m)
	if row >= len(lines) || col >= len(lines[row]) {
		return ""
	}
	return string(lines[row][col])
}

// sendKey fires a direction key ('w','a','s','d') through model.Update.
func sendKey(m *render.Model, key string) {
	msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	updated, _ := m.Update(msg)
	*m = updated.(render.Model)
}

// tick advances the clock by GameTickInterval and fires one TickMsg.
func tick(m *render.Model, clock *game.FakeClock) {
	clock.Advance(game.GameTickInterval)
	updated, _ := m.Update(render.TickMsg(clock.Now()))
	*m = updated.(render.Model)
	renderFrame(*m, "")
}

// moveDir advances the clock by the current tile's move cooldown, then sends the key.
// It reads the player's current tile cooldown from the game state directly.
func moveDir(m *render.Model, clock *game.FakeClock, g *game.Game, dir string) {
	p := g.State.Player
	tile := g.State.World.TileAt(p.X, p.Y)
	cooldown := game.MoveCooldownFor(tile)
	clock.Advance(cooldown)
	sendKey(m, dir)
	renderFrame(*m, fmt.Sprintf("move %s → (%d, %d)", dir, g.State.Player.X, g.State.Player.Y))
}

// TestLogStorageWorkflow is a full end-to-end scenario:
//
//  1. Navigate to a harvest position (48, 45) adjacent to forest.
//  2. Tick until enough wood is cut and player has full inventory (20 wood).
//  3. Verify foundation blocks southward movement from (48,45).
//  4. Tick until the foundation completes via resource deposits (20 wood × 500ms each).
//  5. Move north to (48,44) to restock wood from y=43 trees, return south, deposit.
//  6. Verify player position, LogStorage tile renders as 'L', and built storage registered.
//
// World layout facts for seed 42 with circular clearing radius 5:
//   - Harvest arc from (48,45) facing north: (48,44), (47,44), (49,44) — all Forest.
//   - Foundation spawns at 4×4 footprint (48,46)–(51,49), immediately south of harvest pos.
//   - Player at (48,45) is already adjacent to foundation's north edge.
//   - Foundation blocks southward movement: player stays at (48,45).
//   - With terminal 80×24 and player at (48,45): vpX=8, vpY=34.
//     LogStorage corner (48,46) maps to screen (col=40, row=12).
func TestLogStorageWorkflow(t *testing.T) {
	// ── Setup ────────────────────────────────────────────────────────────────
	clock := game.NewFakeClock()
	g := game.NewWithClockAndRNG(clock, rand.New(rand.NewSource(42)))
	m := render.NewModelWithClock(g, clock)
	// Set terminal size so View() renders.
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(render.Model)

	// ── Phase 1: Navigate to harvest position (48, 45) ───────────────────────
	// Path: west×2 → (48,50), then north×5 → (48,45).
	announcePhase(m, "Phase 1: Navigate to harvest position")
	for _, dir := range []string{"a", "a", "w", "w", "w", "w", "w"} {
		moveDir(&m, clock, g, dir)
	}
	if g.State.Player.X != 48 || g.State.Player.Y != 45 {
		t.Fatalf("phase 1: expected player at (48,45), got (%d,%d)",
			g.State.Player.X, g.State.Player.Y)
	}

	// ── Phase 2: Harvest until foundation appears ────────────────────────────
	// Forward arc faces north: (48,44) size=8, (47,44) size=10, (49,44) size=9.
	// 3 wood/tick → player.Wood reaches InitialCarryingCapacity (20) after ~7 ticks; foundation spawns automatically.
	// Foundation spawns at (48,46)–(51,49), all within clearing radius so guaranteed Grassland.
	// With a fast deposit interval the foundation may complete before this loop ends; accept either.
	announcePhase(m, "Phase 2: Harvest wood until foundation log storage appears")
	const maxHarvestTicks = 30
	for i := range maxHarvestTicks {
		tick(&m, clock)
		if g.State.HasStructureOfType(game.FoundationLogStorage) || g.State.HasStructureOfType(game.LogStorage) {
			break
		}
		if i == maxHarvestTicks-1 {
			t.Fatal("phase 2: foundation did not appear within expected ticks")
		}
	}

	// ── Phase 3: Verify foundation blocks south movement ─────────────────────
	// Player at (48,45) is already adjacent to foundation's north edge at (48,46).
	// Attempting to move south into the foundation should be blocked.
	announcePhase(m, "Phase 3: Verify foundation blocks south movement")
	moveDir(&m, clock, g, "s")
	if g.State.Player.X != 48 || g.State.Player.Y != 45 {
		t.Errorf("phase 3: expected player at (48,45) (foundation blocks south), got (%d,%d)",
			g.State.Player.X, g.State.Player.Y)
	}

	// ── Phase 4: Deposit wood to complete the foundation ─────────────────────
	// Player at (48,45) is adjacent to foundation. Each tick fires TickAdjacentStructures.
	// Foundation completes after LogStorageBuildCost (20) deposits; may already be done from phase 2.
	announcePhase(m, "Phase 4: Build foundation via resource deposit")
	const maxBuildTicks = 120
	for i := range maxBuildTicks {
		tick(&m, clock)
		if g.State.HasStructureOfType(game.LogStorage) {
			break
		}
		if i == maxBuildTicks-1 {
			t.Fatal("phase 4: foundation did not complete within expected ticks")
		}
	}

	// Foundation tiles should be gone; LogStorage should exist.
	if g.State.HasStructureOfType(game.FoundationLogStorage) {
		t.Error("phase 4: FoundationLogStorage tiles should be gone after build completes")
	}
	lsTile := g.State.World.TileAt(49, 48)
	if lsTile == nil || lsTile.Structure != game.LogStorage {
		t.Fatalf("phase 4: expected LogStorage at (49,48), got %v", lsTile)
	}

	// ── Phase 5: Accept the upgrade card ─────────────────────────────────────
	// When the Log Storage is completed, a card offer is queued. The game pauses
	// until the player accepts. Extra ticks during this phase should not change
	// player.Wood (harvest is blocked while paused).
	announcePhase(m, "Phase 5: Accept upgrade card")
	if !g.State.HasPendingOffer() {
		t.Fatal("phase 5: expected a pending card offer after building log storage")
	}
	woodBeforePause := g.State.Player.Wood
	tick(&m, clock) // game should be paused — wood should not change
	if g.State.Player.Wood != woodBeforePause {
		t.Errorf("phase 5: game should be paused during card selection; Wood changed from %d to %d",
			woodBeforePause, g.State.Player.Wood)
	}
	g.State.SelectCard(0)
	if g.State.Player.MaxCarry != 100 {
		t.Errorf("phase 5: MaxCarry = %d, want 100 after accepting carry upgrade", g.State.Player.MaxCarry)
	}
	if g.State.HasPendingOffer() {
		t.Error("phase 5: offer should be gone after SelectCard")
	}

	// Phase 4 leaves player.Wood == 1 (a harvest tick re-stocked 1 wood mid-build).
	// Move north to (48,44) to face north and harvest more from trees at y=43
	// (sizes 9, 4, 4 — untouched by Phase 2). At least 2 ticks are required
	// before the return move: (48,44) is a cut tree (150ms cooldown) so a single
	// tick isn't enough to let the "w" move cooldown (300ms) expire.
	// Return south to (48,45), adjacent to LogStorage, then wait for deposit.
	announcePhase(m, "Phase 6: Restock wood and deposit into built log storage")
	moveDir(&m, clock, g, "w") // Move to (48,44), face north
	const maxRestockTicks = 20
	for i := range maxRestockTicks {
		tick(&m, clock)
		// Require at least 2 ticks so the move cooldown from the "w" step
		// (Forest 300ms) expires before moveDir("s") advances only 150ms.
		if i >= 1 && g.State.Player.Wood > 0 {
			break
		}
		if i == maxRestockTicks-1 {
			t.Fatal("phase 5: could not harvest wood for storage deposit")
		}
	}
	moveDir(&m, clock, g, "s") // Return to (48,45), adjacent to LogStorage
	storedBefore := g.Stores.Total(game.Wood)
	const maxDepositTicks = 30
	for i := range maxDepositTicks {
		tick(&m, clock)
		if g.Stores.Total(game.Wood) > storedBefore {
			break
		}
		if i == maxDepositTicks-1 {
			t.Fatal("phase 5: wood was not deposited into built log storage")
		}
	}

	// ── Assertions ────────────────────────────────────────────────────────────

	// 1. Player position: at (48,45) — foundation blocked south throughout.
	if g.State.Player.X != 48 || g.State.Player.Y != 45 {
		t.Errorf("player position: got (%d,%d), want (48,45)",
			g.State.Player.X, g.State.Player.Y)
	}

	// 2. Status bar shows the correct player position and wood count.
	bar := statusBar(m)
	wantPos := "Player: (48, 45)"
	if !strings.Contains(bar, wantPos) {
		t.Errorf("status bar %q does not contain player position %q", bar, wantPos)
	}
	currentWood := g.State.Player.Wood
	wantWood := fmt.Sprintf("Wood: %d/%d", currentWood, g.State.Player.MaxCarry)
	if !strings.Contains(bar, wantWood) {
		t.Errorf("status bar %q does not contain wood count %q", bar, wantWood)
	}

	// 3. Log storage tile displays as 'L' at its screen position.
	//    Player (48,45) with 80×24 terminal → vpX=8, vpY=34.
	//    LogStorage corner (48,46) → screen col=40, row=12.
	lsChar := charAtScreen(m, 40, 12)
	if lsChar != "L" {
		t.Errorf("expected 'L' (LogStorage) at screen(col=40,row=12), got %q", lsChar)
	}

	// 4. Log storage contains deposited wood.
	if g.Stores.Total(game.Wood) == 0 {
		t.Error("expected wood to be stored in log storage after deposit")
	}

	announcePhase(m, fmt.Sprintf("Done — LogStorage built and %d wood deposited", g.Stores.Total(game.Wood)))
}
