package e2e_tests

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
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

// tick advances the clock by GameTickInterval and fires one TickMsg.
func tick(m *render.Model, clock *game.FakeClock) {
	clock.Advance(game.GameTickInterval)
	updated, _ := m.Update(render.TickMsg(clock.Now()))
	*m = updated.(render.Model)
	renderFrame(*m, "")
}

// drainOffers accepts card 0 for every pending offer, unpausing the game.
// Use this inside tick loops that should not be interrupted by XP milestone cards.
func drainOffers(g *game.Game) {
	for g.HasPendingOffer() {
		g.SelectCard(0)
	}
}

// tickDraining ticks once then auto-accepts all pending card offers.
// Use inside harvest/deposit/build loops to skip XP milestone interruptions.
func tickDraining(m *render.Model, clock *game.FakeClock, g *game.Game) {
	tick(m, clock)
	drainOffers(g)
}

// moveDir moves the player one tile in the given direction by calling MoveSmooth
// in small steps until the tile position changes. This avoids float64 precision
// issues when computing an exact one-tile traversal duration.
func moveDir(m *render.Model, clock *game.FakeClock, g *game.Game, dir string) {
	p := g.State.Player
	startX, startY := p.TileX(), p.TileY()
	var dx, dy float64
	switch dir {
	case "w":
		dy = -1
	case "s":
		dy = 1
	case "a":
		dx = -1
	case "d":
		dx = 1
	}
	const (
		step     = 5 * time.Millisecond
		maxSteps = 200 // 1 second total; exit early if movement is blocked
	)
	for i := 0; i < maxSteps && p.TileX() == startX && p.TileY() == startY; i++ {
		clock.Advance(step)
		p.MoveSmooth(dx, dy, g.State.World, step)
	}
	renderFrame(*m, fmt.Sprintf("move %s → (%d, %d)", dir, p.TileX(), p.TileY()))
}

// moveSafe is an alias for moveDir. Previously it handled Forest→Grassland cooldown
// transitions, but MoveSmooth has no per-player move cooldown so the two are equivalent.
func moveSafe(m *render.Model, clock *game.FakeClock, g *game.Game, dir string) {
	moveDir(m, clock, g, dir)
}

// newTestGame creates a standard test game with seed-42 RNG and an 80×24 terminal.
// All e2e tests use this as their base setup.
func newTestGame() (*game.Game, *game.FakeClock, render.Model) {
	clock := game.NewFakeClock()
	g := game.NewWithClockAndRNG(clock, rand.New(rand.NewSource(42)))
	m := render.NewModelWithClock(g, clock)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	return g, clock, updated.(render.Model)
}

// loadFixture restores a game from a pre-built JSON fixture in e2e_tests/testdata/.
// The returned game uses a fresh FakeClock at time 0 and a seed-42 RNG.
// Regenerate fixtures with: go test -run TestGenerateFixtures -update-fixtures ./e2e_tests/
func loadFixture(t *testing.T, name string) (*game.Game, *game.FakeClock, render.Model) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name+".json"))
	if err != nil {
		t.Fatalf("loadFixture %q: %v", name, err)
	}
	var saveData game.SaveGameData
	if err := json.Unmarshal(raw, &saveData); err != nil {
		t.Fatalf("loadFixture %q: unmarshal: %v", name, err)
	}
	g, clock, m := newTestGame()
	if err := g.LoadFrom(saveData); err != nil {
		t.Fatalf("loadFixture %q: LoadFrom: %v", name, err)
	}
	return g, clock, m
}

// writeFixture serialises the current game state to e2e_tests/testdata/<name>.json.
// Used only by the fixture generator (TestGenerateFixtures).
func writeFixture(t *testing.T, name string, g *game.Game) {
	t.Helper()
	if err := os.MkdirAll("testdata", 0o755); err != nil {
		t.Fatalf("writeFixture %q: mkdir: %v", name, err)
	}
	data, err := json.MarshalIndent(g.SaveData(), "", "  ")
	if err != nil {
		t.Fatalf("writeFixture %q: marshal: %v", name, err)
	}
	path := filepath.Join("testdata", name+".json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("writeFixture %q: write: %v", name, err)
	}
	t.Logf("wrote %s (%d bytes)", path, len(data))
}
