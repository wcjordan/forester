package e2e_tests

import (
	"math/rand"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"forester/game"
	"forester/game/geom"
	_ "forester/game/resources"
	"forester/game/structures"
	_ "forester/game/upgrades"
	"forester/render"
)

// TestFoundationProgressRGB verifies the shared amber→gold color helper that
// both the TUI and Ebiten renderers use for build-progress shading.
func TestFoundationProgressRGB(t *testing.T) {
	tests := []struct {
		progress            float64
		wantR, wantG, wantB uint8
	}{
		{0.0, 80, 60, 0},   // dark amber
		{1.0, 255, 215, 0}, // bright gold
		{0.5, 167, 137, 0}, // midpoint
	}
	for _, tt := range tests {
		r, g, b := render.FoundationProgressRGB(tt.progress)
		if r != tt.wantR || g != tt.wantG || b != tt.wantB {
			t.Errorf("FoundationProgressRGB(%.1f) = (%d,%d,%d), want (%d,%d,%d)",
				tt.progress, r, g, b, tt.wantR, tt.wantG, tt.wantB)
		}
	}
	// Confirm monotonically increasing brightness.
	r0, g0, _ := render.FoundationProgressRGB(0)
	r1, g1, _ := render.FoundationProgressRGB(1)
	if r1 <= r0 || g1 <= g0 {
		t.Errorf("color should brighten with progress: 0%%=(%d,%d) 100%%=(%d,%d)", r0, g0, r1, g1)
	}
}

// TestFoundationTUIShading verifies that foundation tiles render as '?' and
// that the character is stable across build-progress changes (colour changes,
// character does not).
//
// Setup mirrors TestLogStorageWorkflow (seed 42, player navigated to (48,45))
// so we know the foundation spawns at origin (48,46).  With terminal 80×24 the
// foundation origin maps to screen col=40, row=12.
func TestFoundationTUIShading(t *testing.T) {
	clock := game.NewFakeClock()
	g := game.NewWithClockAndRNG(clock, rand.New(rand.NewSource(42)))
	m := render.NewModelWithClock(g, clock)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(render.Model)

	// Navigate to (48,45) — same path as TestLogStorageWorkflow Phase 1.
	for _, dir := range []string{"a", "a", "w", "w", "w", "w", "w"} {
		moveDir(&m, clock, g, dir)
	}

	// Tick until the log-storage foundation appears (Phase 2).
	const maxTicks = 30
	for i := range maxTicks {
		tick(&m, clock)
		if g.State.World.HasStructureOfType(structures.FoundationLogStorage) {
			break
		}
		if i == maxTicks-1 {
			t.Fatal("foundation did not appear within expected ticks")
		}
	}

	// Locate the foundation origin in FoundationDeposited.
	var origin geom.Point
	for pt := range g.State.FoundationDeposited {
		origin = pt
		break
	}
	if origin == (geom.Point{}) {
		t.Fatal("no foundation deposit entry found after foundation spawned")
	}

	// Foundation origin (48,46) should map to screen col=40, row=12.
	// vpX = clamp(48-40, 0, max(0,100-80)) = 8
	// vpY = clamp(45-11, 0, max(0,100-23)) = 34
	const screenCol, screenRow = 40, 12

	// Verify the tile character is '?' at 0% and 100% progress.
	for _, deposited := range []int{0, 10, 20} {
		g.State.FoundationDeposited[origin] = deposited
		ch := charAtScreen(m, screenCol, screenRow)
		if ch != "?" {
			t.Errorf("foundation tile char at %d deposited = %q, want \"?\"", deposited, ch)
		}
	}

	// The stripped (no-ANSI) foundation row must be identical regardless of
	// progress — same '?' character, only the colour wrapping changes.
	g.State.FoundationDeposited[origin] = 0
	row0 := strings.Split(stripANSI(m.View()), "\n")[screenRow]
	g.State.FoundationDeposited[origin] = 20
	row100 := strings.Split(stripANSI(m.View()), "\n")[screenRow]
	if row0 != row100 {
		t.Errorf("stripped foundation row differs at 0%% vs 100%%: %q vs %q", row0, row100)
	}

	// FoundationProgressAt must return correct progress for tiles in the footprint.
	g.State.FoundationDeposited[origin] = 10
	progress, ok := g.State.FoundationProgressAt(origin)
	if !ok {
		t.Error("FoundationProgressAt(origin): ok = false after deposit")
	}
	if progress != 0.5 {
		t.Errorf("FoundationProgressAt(origin) = %v, want 0.5", progress)
	}
}
