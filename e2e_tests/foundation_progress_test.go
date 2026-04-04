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

// TestFoundationTUIShading verifies that foundation tiles render with different
// colors at different build progress levels.
//
// Setup mirrors TestLogStorageWorkflow (seed 42, player navigated to (48,45))
// so we know the foundation spawns at origin (48,46).  With terminal 80×24 the
// foundation origin maps to screen col=40, row=12.
//
// The test checks:
//  1. The character at the foundation tile is always '?'.
//  2. The raw (ANSI-inclusive) view differs between 0% and 100% progress —
//     proving the colour changes with build progress.
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

	// Verify the tile character is '?' in the stripped view.
	ch := charAtScreen(m, screenCol, screenRow)
	if ch != "?" {
		t.Errorf("foundation tile char = %q, want \"?\"", ch)
	}

	// Capture raw view at 0% progress.
	g.State.FoundationDeposited[origin] = 0
	view0 := m.View()

	// Capture raw view at 100% progress.
	const logStorageBuildCost = 20
	g.State.FoundationDeposited[origin] = logStorageBuildCost
	view100 := m.View()

	// The character must still be '?' at full progress.
	ch = charAtScreen(m, screenCol, screenRow)
	if ch != "?" {
		t.Errorf("foundation tile char at 100%% = %q, want \"?\"", ch)
	}

	// The stripped (no-ANSI) views at the foundation row must be identical —
	// same character, just different colours.
	lines0 := strings.Split(stripANSI(view0), "\n")
	lines100 := strings.Split(stripANSI(view100), "\n")
	if screenRow < len(lines0) && screenRow < len(lines100) {
		if lines0[screenRow] != lines100[screenRow] {
			t.Errorf("stripped foundation row differs: %q vs %q", lines0[screenRow], lines100[screenRow])
		}
	}

	// The raw views must differ — confirming the colour changes with progress.
	if view0 == view100 {
		t.Error("raw view unchanged between 0%% and 100%% progress: foundation shading not applied")
	}
}
