package render

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"forester/game"
)

// Model is the bubbletea model for the game. It owns viewport dimensions
// and delegates all game logic to game.Game.
type Model struct {
	game       *game.Game
	termWidth  int
	termHeight int
}

// NewModel creates a Model wrapping the given game.
func NewModel(g *game.Game) Model {
	return Model{game: g}
}

// Init satisfies tea.Model. No initial commands needed.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and returns the updated model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "up", "w":
			m.game.State.Player.MovePlayer(0, -1, m.game.State.World)
		case "down", "s":
			m.game.State.Player.MovePlayer(0, 1, m.game.State.World)
		case "left", "a":
			m.game.State.Player.MovePlayer(-1, 0, m.game.State.World)
		case "right", "d":
			m.game.State.Player.MovePlayer(1, 0, m.game.State.World)
		}
	}

	return m, nil
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
				sb.WriteByte('@')
				continue
			}

			tile := world.TileAt(worldX, worldY)
			if tile == nil {
				sb.WriteByte(' ')
				continue
			}

			switch tile.Terrain {
			case game.Forest:
				sb.WriteByte('#')
			default:
				sb.WriteByte('.')
			}
		}
		if row < mapHeight-1 {
			sb.WriteByte('\n')
		}
	}

	// Status bar.
	status := fmt.Sprintf(" Player: (%d, %d)  Wood: %d",
		player.X, player.Y, player.Wood)
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
