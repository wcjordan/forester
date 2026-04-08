package e2e_tests

import (
	"fmt"
	"regexp"
	"strings"

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

// moveDir advances the clock by the current tile's move cooldown, then sends the key.
// It reads the player's current tile cooldown from the game state directly.
func moveDir(m *render.Model, clock *game.FakeClock, g *game.Game, dir string) {
	p := g.State.Player
	tile := g.State.World.TileAt(p.TileX(), p.TileY())
	cooldown := game.MoveCooldownFor(tile)
	clock.Advance(cooldown)
	sendKey(m, dir)
	renderFrame(*m, fmt.Sprintf("move %s → (%d, %d)", dir, g.State.Player.TileX(), g.State.Player.TileY()))
}

// moveSafe advances the clock by the greater of the current tile's move cooldown
// or the remaining move cooldown, then sends the directional key. This correctly
// handles Forest→Grassland transitions: after moving off a Forest tile (300ms
// cooldown set), the subsequent Grassland move (only 150ms) would fail with
// moveDir because the previous 300ms cooldown hasn't expired.
func moveSafe(m *render.Model, clock *game.FakeClock, g *game.Game, dir string) {
	p := g.State.Player
	tile := g.State.World.TileAt(p.TileX(), p.TileY())
	needed := game.MoveCooldownFor(tile)
	remaining := p.Cooldowns[game.Move].Sub(clock.Now())
	if remaining > needed {
		clock.Advance(remaining)
	} else {
		clock.Advance(needed)
	}
	sendKey(m, dir)
	renderFrame(*m, fmt.Sprintf("moveSafe %s → (%d,%d)", dir, g.State.Player.TileX(), g.State.Player.TileY()))
}
