package render

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"forester/game"
)

// TickMsg is sent each harvest tick interval to drive the game loop.
type TickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(game.HarvestTickInterval, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

var (
	playerStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))           // blue
	forestStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))            // green
	stumpStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))            // dark gray
	foundationStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))            // yellow (dim)
	logStorageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true) // bold yellow
)

// Model is the bubbletea model for the game. It owns viewport dimensions
// and delegates all game logic to game.Game.
type Model struct {
	game       *game.Game
	termWidth  int
	termHeight int
	clock      game.Clock
}

// NewModel creates a Model wrapping the given game using the system clock.
func NewModel(g *game.Game) Model {
	return NewModelWithClock(g, game.RealClock{})
}

// NewModelWithClock creates a Model with the given clock. Use in tests to
// inject a FakeClock for deterministic time control.
func NewModelWithClock(g *game.Game, clock game.Clock) Model {
	return Model{game: g, clock: clock}
}

// Init satisfies tea.Model. Starts the harvest tick loop.
func (m Model) Init() tea.Cmd {
	return doTick()
}

// Update handles messages and returns the updated model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height

	case TickMsg:
		m.game.Tick()
		return m, doTick()

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}

		if m.game.State.HasPendingOffer() {
			if msg.String() == "1" || msg.String() == "enter" {
				m.game.State.SelectCard(0)
			}
			return m, nil
		}

		switch msg.String() {
		case "up", "w":
			m.game.State.Player.Move(0, -1, m.game.State.World, m.clock.Now())
		case "down", "s":
			m.game.State.Player.Move(0, 1, m.game.State.World, m.clock.Now())
		case "left", "a":
			m.game.State.Player.Move(-1, 0, m.game.State.World, m.clock.Now())
		case "right", "d":
			m.game.State.Player.Move(1, 0, m.game.State.World, m.clock.Now())
		}
	}

	return m, nil
}

// View renders the current game state to a string.
func (m Model) View() string {
	if m.termWidth == 0 || m.termHeight == 0 {
		return ""
	}

	if m.game.State.HasPendingOffer() {
		return m.renderCardScreen()
	}

	// Status bar occupies the last line; map gets the rest.
	mapHeight := m.termHeight - 1
	mapWidth := m.termWidth

	world := m.game.State.World
	player := m.game.State.Player

	// Top-left corner of the viewport in world coordinates.
	// Center the viewport on the player, clamped to world bounds.
	vpX := clamp(player.X-mapWidth/2, 0, max(0, world.Width-mapWidth))
	vpY := clamp(player.Y-mapHeight/2, 0, max(0, world.Height-mapHeight))

	var sb strings.Builder
	for row := 0; row < mapHeight; row++ {
		for col := 0; col < mapWidth; col++ {
			worldX := vpX + col
			worldY := vpY + row

			if worldX == player.X && worldY == player.Y {
				sb.WriteString(playerStyle.Render("@"))
				continue
			}

			tile := world.TileAt(worldX, worldY)
			if tile == nil {
				sb.WriteByte(' ')
				continue
			}

			// Structure overlays take priority over terrain.
			switch tile.Structure {
			case game.FoundationLogStorage:
				sb.WriteString(foundationStyle.Render("?"))
				continue
			case game.LogStorage:
				sb.WriteString(logStorageStyle.Render("L"))
				continue
			}

			switch tile.Terrain {
			case game.Forest:
				switch {
				case tile.TreeSize == 0:
					sb.WriteString(stumpStyle.Render("%"))
				case tile.TreeSize >= 7:
					sb.WriteString(forestStyle.Render("#"))
				case tile.TreeSize >= 4:
					sb.WriteString(forestStyle.Render("t"))
				default:
					sb.WriteString(forestStyle.Render(","))
				}
			default:
				sb.WriteByte('.')
			}
		}
		if row < mapHeight-1 {
			sb.WriteByte('\n')
		}
	}

	// Status bar.
	status := fmt.Sprintf(" Player: (%d, %d)  Wood: %d/%d",
		player.X, player.Y, player.Wood, player.MaxCarry)
	for _, deposited := range m.game.State.FoundationDeposited {
		progress := float64(deposited) / float64(game.LogStorageBuildCost)
		status += "  " + buildProgressBar(progress)
		break // show at most one foundation progress bar
	}
	sb.WriteByte('\n')
	sb.WriteString(status)

	return sb.String()
}

// renderCardScreen renders the milestone card selection overlay centered in the terminal.
func (m Model) renderCardScreen() string {
	offer := m.game.State.CurrentOffer()
	if len(offer) == 0 {
		return ""
	}
	card := offer[0]

	const (
		outerWidth  = 44
		innerWidth  = 42 // chars between the two ║ borders
		cardContent = 36 // chars between the two │ borders
	)

	// padRight pads s to exactly width rune-columns using trailing spaces.
	padRight := func(s string, width int) string {
		n := utf8.RuneCountInString(s)
		if n >= width {
			return s
		}
		return s + strings.Repeat(" ", width-n)
	}

	// centerIn centers s within a field of width rune-columns.
	centerIn := func(s string, width int) string {
		n := utf8.RuneCountInString(s)
		if n >= width {
			return s
		}
		left := (width - n) / 2
		right := width - n - left
		return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
	}

	outerFill := strings.Repeat("═", innerWidth)
	cardFill := strings.Repeat("─", cardContent)
	emptyCard := "║  │" + strings.Repeat(" ", cardContent) + "│  ║"

	var lines []string
	lines = append(lines, "╔"+outerFill+"╗")
	lines = append(lines, "║"+centerIn("MILESTONE", innerWidth)+"║")
	lines = append(lines, "╠"+outerFill+"╣")
	lines = append(lines, "║"+strings.Repeat(" ", innerWidth)+"║")
	lines = append(lines, "║  ┌"+cardFill+"┐  ║")
	lines = append(lines, "║  │  "+padRight(strings.ToUpper(card.Name()), cardContent-2)+"│  ║")
	lines = append(lines, emptyCard)
	for _, descLine := range strings.Split(card.Description(), "\n") {
		lines = append(lines, "║  │  "+padRight(descLine, cardContent-2)+"│  ║")
	}
	lines = append(lines, emptyCard)
	lines = append(lines, "║  │"+centerIn("[ 1 ] Accept", cardContent)+"│  ║")
	lines = append(lines, "║  └"+cardFill+"┘  ║")
	lines = append(lines, "║"+strings.Repeat(" ", innerWidth)+"║")
	lines = append(lines, "║"+centerIn("Press 1 or ENTER to accept", innerWidth)+"║")
	lines = append(lines, "╚"+outerFill+"╝")

	boxHeight := len(lines)

	leftPad := (m.termWidth - outerWidth) / 2
	if leftPad < 0 {
		leftPad = 0
	}
	prefix := strings.Repeat(" ", leftPad)

	topPad := (m.termHeight - boxHeight) / 2
	if topPad < 0 {
		topPad = 0
	}

	var sb strings.Builder
	for i := 0; i < topPad; i++ {
		sb.WriteByte('\n')
	}
	for i, line := range lines {
		sb.WriteString(prefix + line)
		if i < len(lines)-1 {
			sb.WriteByte('\n')
		}
	}
	for i := topPad + boxHeight; i < m.termHeight; i++ {
		sb.WriteByte('\n')
	}

	return sb.String()
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// buildProgressBar renders a text progress bar for a build operation.
// e.g. "Building: ████░░░░ 75%"
func buildProgressBar(progress float64) string {
	const width = 8
	filled := clamp(int(progress*width), 0, width)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("Building: %s %d%%", bar, int(progress*100))
}
