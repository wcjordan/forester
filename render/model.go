package render

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"forester/game"
)

type tickMsg time.Time
type regrowTickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(game.HarvestTickInterval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func doRegrowTick() tea.Cmd {
	return tea.Tick(game.RegrowthTickInterval, func(t time.Time) tea.Msg {
		return regrowTickMsg(t)
	})
}

var (
	playerStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))           // blue
	forestStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))            // green
	stumpStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))            // dark gray
	ghostStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))            // yellow (dim)
	logStorageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true) // bold yellow
)

// DepositTickInterval is how often the player auto-deposits one wood when adjacent to Log Storage.
const DepositTickInterval = 500 * time.Millisecond

// Model is the bubbletea model for the game. It owns viewport dimensions
// and delegates all game logic to game.Game.
type Model struct {
	game            *game.Game
	termWidth       int
	termHeight      int
	lastMoveTime    time.Time
	depositCooldown time.Time
}

// NewModel creates a Model wrapping the given game.
func NewModel(g *game.Game) Model {
	return Model{game: g}
}

// Init satisfies tea.Model. Starts the harvest and regrowth tick loops.
func (m Model) Init() tea.Cmd {
	return tea.Batch(doTick(), doRegrowTick())
}

// Update handles messages and returns the updated model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height

	case tickMsg:
		m.game.State.Harvest()
		m.game.State.AdvanceBuild()
		if time.Now().After(m.depositCooldown) {
			before := m.game.State.LogStorageDeposited
			m.game.State.TickAdjacentStructures()
			if m.game.State.LogStorageDeposited > before {
				m.depositCooldown = time.Now().Add(DepositTickInterval)
			}
		}
		return m, doTick()

	case regrowTickMsg:
		m.game.State.Regrow()
		return m, doRegrowTick()

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "up", "w":
			if m.canMove() {
				m.game.State.Move(0, -1)
				m.lastMoveTime = time.Now()
			}
		case "down", "s":
			if m.canMove() {
				m.game.State.Move(0, 1)
				m.lastMoveTime = time.Now()
			}
		case "left", "a":
			if m.canMove() {
				m.game.State.Move(-1, 0)
				m.lastMoveTime = time.Now()
			}
		case "right", "d":
			if m.canMove() {
				m.game.State.Move(1, 0)
				m.lastMoveTime = time.Now()
			}
		}
	}

	return m, nil
}

// canMove returns true if enough time has elapsed since the last move,
// based on the terrain the player is currently standing on.
func (m Model) canMove() bool {
	p := m.game.State.Player
	tile := m.game.State.World.TileAt(p.X, p.Y)
	cooldown := game.DefaultMoveCooldown
	if tile != nil {
		if d, ok := game.MoveCooldowns[tile.Terrain]; ok {
			cooldown = d
		}
	}
	return time.Since(m.lastMoveTime) >= cooldown
}

// View renders the current game state to a string.
func (m Model) View() string {
	if m.termWidth == 0 || m.termHeight == 0 {
		return ""
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
			case game.GhostLogStorage:
				sb.WriteString(ghostStyle.Render("?"))
				continue
			case game.LogStorage:
				sb.WriteString(logStorageStyle.Render("L"))
				continue
			}

			switch tile.Terrain {
			case game.Forest:
				switch {
				case tile.TreeSize >= 7:
					sb.WriteString(forestStyle.Render("#"))
				case tile.TreeSize >= 4:
					sb.WriteString(forestStyle.Render("t"))
				default:
					sb.WriteString(forestStyle.Render(","))
				}
			case game.Stump:
				sb.WriteString(stumpStyle.Render("%"))
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
		player.X, player.Y, player.Wood, game.MaxWood)
	if b := m.game.State.Building; b != nil {
		status += "  " + buildProgressBar(b.Progress())
	}
	sb.WriteByte('\n')
	sb.WriteString(status)

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
	filled := int(progress * width)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("Building: %s %d%%", bar, int(progress*100))
}
