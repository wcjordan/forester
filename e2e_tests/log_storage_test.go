// Package e2e_tests contains end-to-end tests that drive the full game stack
// through the bubbletea model — the same code path as a real user.
package e2e_tests

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"forester/game"
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

// tick advances the clock by HarvestTickInterval and fires one TickMsg.
func tick(m *render.Model, clock *game.FakeClock) {
	clock.Advance(game.HarvestTickInterval)
	updated, _ := m.Update(render.TickMsg(clock.Now()))
	*m = updated.(render.Model)
}

// moveDir advances the clock by the current tile's move cooldown, then sends the key.
// It reads the player's current tile cooldown from the game state directly.
func moveDir(m *render.Model, clock *game.FakeClock, g *game.Game, dir string) {
	p := g.State.Player
	tile := g.State.World.TileAt(p.X, p.Y)
	cooldown := game.MoveCooldownFor(tile)
	clock.Advance(cooldown)
	sendKey(m, dir)
}

// TestLogStorageWorkflow is a full end-to-end scenario:
//
//  1. Navigate to a harvest position (48, 45) adjacent to forest.
//  2. Tick until enough wood is cut to trigger the ghost log storage.
//  3. Walk onto the ghost tile to start the build, which nudges the player to (49, 47).
//  4. Tick until build completes (30 ticks).
//  5. Deposit wood into the completed log storage.
//  6. Verify player position, UI tile icons, and status bar text.
//
// World layout facts for seed 42 used in assertions:
//   - Harvest arc from (48,45) facing north: (48,44), (47,44), (49,44) — all Forest.
//   - Ghost spawns at 4×4 footprint (49,48)–(52,51).
//   - After ghost contact at (49,48) player is nudged to (49,47).
//   - With terminal 80×24 and player at (49,47): vpX=9, vpY=36.
//     LogStorage corner (49,48) maps to screen (col=40, row=12).
//     Harvested trees (47–49, 44) map to screen (col=38–40, row=8).
func TestLogStorageWorkflow(t *testing.T) {
	// ── Setup ────────────────────────────────────────────────────────────────
	clock := game.NewFakeClock()
	g := game.NewWithClock(clock)
	m := render.NewModelWithClock(g, clock)
	// Set terminal size so View() renders.
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(render.Model)

	// ── Phase 1: Navigate to harvest position (48, 45) ───────────────────────
	// Path: west×2 → (48,50), then north×5 → (48,45).
	// Tiles along path: all Grassland except the last two northward steps (Forest).
	for _, dir := range []string{"a", "a", "w", "w", "w", "w", "w"} {
		moveDir(&m, clock, g, dir)
	}
	if g.State.Player.X != 48 || g.State.Player.Y != 45 {
		t.Fatalf("phase 1: expected player at (48,45), got (%d,%d)",
			g.State.Player.X, g.State.Player.Y)
	}

	// ── Phase 2: Harvest until ghost log storage appears ─────────────────────
	// Forward arc faces north: (48,44) size=8, (47,44) size=10, (49,44) size=9.
	// 3 wood/tick → TotalWoodCut≥10 after ~4 ticks; ghost spawns automatically.
	const maxHarvestTicks = 30
	for i := range maxHarvestTicks {
		tick(&m, clock)
		if g.State.HasStructureOfType(game.GhostLogStorage) {
			break
		}
		if i == maxHarvestTicks-1 {
			t.Fatal("phase 2: ghost log storage did not appear after harvesting")
		}
	}

	// ── Phase 3: Step onto ghost tile to trigger build ────────────────────────
	// From (48,45): south×3 → (48,48), then east → (49,48) which is the ghost.
	// checkGhostContact fires; build starts; player is nudged to (49,47).
	for _, dir := range []string{"s", "s", "s", "d"} {
		moveDir(&m, clock, g, dir)
	}
	if g.State.Building == nil {
		t.Fatal("phase 3: build should have started after stepping on ghost tile")
	}
	if g.State.Player.X != 49 || g.State.Player.Y != 47 {
		t.Errorf("phase 3: expected player nudged to (49,47), got (%d,%d)",
			g.State.Player.X, g.State.Player.Y)
	}

	// ── Phase 4: Complete build (30 ticks) ────────────────────────────────────
	const maxBuildTicks = 40
	for i := range maxBuildTicks {
		tick(&m, clock)
		if g.State.Building == nil {
			break
		}
		if i == maxBuildTicks-1 {
			t.Fatal("phase 4: build did not complete within expected ticks")
		}
	}

	// LogStorage footprint starts at (49,48).
	lsTile := g.State.World.TileAt(49, 48)
	if lsTile == nil || lsTile.Structure != game.LogStorage {
		t.Fatalf("phase 4: expected LogStorage at (49,48), got %v", lsTile)
	}

	// ── Phase 5: Deposit wood into the log storage ────────────────────────────
	// Player is at (49,47), directly adjacent to LogStorage at (49,48).
	// Note: one deposit already fires in the same Tick() that completes the build
	// (AdvanceBuild runs before TryDeposit). storedBefore captures whatever is
	// already stored so we can verify at least one more deposit occurs here.
	if g.State.Player.Wood == 0 {
		t.Fatal("phase 5: player has no wood to deposit")
	}
	storedBefore := g.State.TotalStored(game.Wood)
	// Advance clock past the deposit cooldown so the next TryDeposit call fires.
	clock.Advance(game.DepositTickInterval + time.Millisecond)
	const maxDepositTicks = 30
	for i := range maxDepositTicks {
		tick(&m, clock)
		if g.State.TotalStored(game.Wood) > storedBefore {
			break
		}
		if i == maxDepositTicks-1 {
			t.Fatal("phase 5: wood was not deposited into log storage")
		}
	}

	// ── Assertions ────────────────────────────────────────────────────────────

	// 1. Player position after the full workflow.
	if g.State.Player.X != 49 || g.State.Player.Y != 47 {
		t.Errorf("player position: got (%d,%d), want (49,47)",
			g.State.Player.X, g.State.Player.Y)
	}

	// 2. Status bar shows the correct player position and wood count.
	bar := statusBar(m)
	wantPos := "Player: (49, 47)"
	if !strings.Contains(bar, wantPos) {
		t.Errorf("status bar %q does not contain player position %q", bar, wantPos)
	}
	// Snapshot wood count at assertion time — harvest and deposit both run each
	// tick so the exact value varies; we verify the format is present and correct.
	currentWood := g.State.Player.Wood
	wantWood := fmt.Sprintf("Wood: %d/20", currentWood)
	if !strings.Contains(bar, wantWood) {
		t.Errorf("status bar %q does not contain wood count %q", bar, wantWood)
	}

	// 3. Log storage tile displays as 'L' at its screen position.
	//    Player (49,47) with 80×24 terminal → vpX=9, vpY=36.
	//    LogStorage corner (49,48) → screen col=40, row=12.
	lsChar := charAtScreen(m, 40, 12)
	if lsChar != "L" {
		t.Errorf("expected 'L' (LogStorage) at screen(col=40,row=12), got %q", lsChar)
	}

	// 4. Partially-harvested trees north of harvest position display as 't'.
	//    Initial sizes: (47,44)=10, (48,44)=8, (49,44)=9. After 4 harvest ticks
	//    (1 per tick, 3 tiles) sizes are 6, 4, 5 — all in the 't' range (4–6).
	//    Player hit MaxWood before fully cutting them.
	//    Viewport: vpX=9, vpY=36 → trees at screen row=8, cols=38,39,40.
	for _, col := range []int{38, 39, 40} {
		ch := charAtScreen(m, col, 8)
		if ch != "t" {
			t.Errorf("expected 't' (partially harvested tree) at screen(col=%d,row=8), got %q", col, ch)
		}
	}

	// 5. Log storage contains the deposited wood.
	if g.State.TotalStored(game.Wood) == 0 {
		t.Error("expected wood to be stored in log storage after deposit")
	}
}
