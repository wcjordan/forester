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

// TickMsg is sent each tick interval to drive the game loop.
type TickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(game.GameTickInterval, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

var (
	playerStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))           // blue
	forestStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))            // green
	stumpStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))            // dark gray
	foundationStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))            // yellow (dim)
	logStorageStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true) // bold yellow
	houseStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true) // bold magenta
	villagerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))           // cyan
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
			switch msg.String() {
			case "1", "enter":
				m.game.State.SelectCard(0)
			case "2":
				m.game.State.SelectCard(1)
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

			isVillager := false
			for _, v := range m.game.State.Villagers {
				if worldX == v.X && worldY == v.Y {
					isVillager = true
					break
				}
			}
			if isVillager {
				sb.WriteString(villagerStyle.Render("v"))
				continue
			}

			tile := world.TileAt(worldX, worldY)
			if tile == nil {
				sb.WriteByte(' ')
				continue
			}

			// Structure overlays take priority over terrain.
			switch tile.Structure {
			case game.FoundationLogStorage, game.FoundationHouse:
				sb.WriteString(foundationStyle.Render("?"))
				continue
			case game.LogStorage:
				sb.WriteString(logStorageStyle.Render("L"))
				continue
			case game.House:
				sb.WriteString(houseStyle.Render("H"))
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

	logStored := m.game.Stores.Total(game.Wood)
	logCap := m.game.Stores.TotalCapacity(game.Wood)
	if logCap > 0 {
		status += fmt.Sprintf("  Log: %d/%d", logStored, logCap)
	}

	villagerCount := len(m.game.State.Villagers)
	houseCount := world.CountStructureInstances(game.House)
	if villagerCount > 0 || houseCount > 0 {
		status += fmt.Sprintf("  Villagers: %d/%d", villagerCount, houseCount)
	}

	if progress, ok := m.game.State.FoundationProgress(); ok {
		status += "  " + buildProgressBar(progress)
	}
	sb.WriteByte('\n')
	sb.WriteString(status)

	return sb.String()
}

// renderCardScreen renders the milestone card selection overlay centered in the terminal.
// When two cards are offered they are displayed side by side horizontally.
func (m Model) renderCardScreen() string {
	offer := m.game.State.CurrentOffer()
	if len(offer) == 0 {
		return ""
	}

	const (
		cardContent = 32 // chars between the two │ borders
		cardPad     = 2  // spaces between ║ border and card box
		gap         = 4  // spaces between two side-by-side cards
	)

	cardBoxWidth := cardContent + 2 // includes the │ borders

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

	cardFill := strings.Repeat("─", cardContent)
	emptyCard := "│" + strings.Repeat(" ", cardContent) + "│"

	// buildCardLines returns the box lines for one card (no outer ║ border).
	buildCardLines := func(card game.UpgradeDef, keyHint string) []string {
		var lines []string
		lines = append(lines, "┌"+cardFill+"┐")
		lines = append(lines, "│  "+padRight(strings.ToUpper(card.Name()), cardContent-2)+"│")
		lines = append(lines, emptyCard)
		for _, descLine := range strings.Split(card.Description(), "\n") {
			lines = append(lines, "│  "+padRight(descLine, cardContent-2)+"│")
		}
		lines = append(lines, emptyCard)
		lines = append(lines, "│"+centerIn("[ "+keyHint+" ] Accept", cardContent)+"│")
		lines = append(lines, "└"+cardFill+"┘")
		return lines
	}

	// insertEmptyCard inserts an empty card line at position pos.
	insertEmptyCard := func(lines []string, pos int) []string {
		result := make([]string, len(lines)+1)
		copy(result, lines[:pos])
		result[pos] = emptyCard
		copy(result[pos+1:], lines[pos:])
		return result
	}

	var cardLinesList [][]string
	var innerWidth int

	if len(offer) >= 2 {
		c1 := buildCardLines(offer[0], "1")
		c2 := buildCardLines(offer[1], "2")
		// Normalize card heights: insert empty lines before the accept line.
		for len(c1) < len(c2) {
			c1 = insertEmptyCard(c1, len(c1)-2)
		}
		for len(c2) < len(c1) {
			c2 = insertEmptyCard(c2, len(c2)-2)
		}
		cardLinesList = [][]string{c1, c2}
		innerWidth = cardPad + cardBoxWidth + gap + cardBoxWidth + cardPad
	} else {
		cardLinesList = [][]string{buildCardLines(offer[0], "1")}
		innerWidth = cardPad + cardBoxWidth + cardPad
	}

	outerWidth := innerWidth + 2
	outerFill := strings.Repeat("═", innerWidth)

	var lines []string
	lines = append(lines, "╔"+outerFill+"╗")
	lines = append(lines, "║"+centerIn("MILESTONE", innerWidth)+"║")
	lines = append(lines, "╠"+outerFill+"╣")
	lines = append(lines, "║"+strings.Repeat(" ", innerWidth)+"║")

	numCardLines := len(cardLinesList[0])
	for i := 0; i < numCardLines; i++ {
		var row string
		if len(cardLinesList) >= 2 {
			row = strings.Repeat(" ", cardPad) + cardLinesList[0][i] + strings.Repeat(" ", gap) + cardLinesList[1][i] + strings.Repeat(" ", cardPad)
		} else {
			row = strings.Repeat(" ", cardPad) + cardLinesList[0][i] + strings.Repeat(" ", cardPad)
		}
		lines = append(lines, "║"+row+"║")
	}

	lines = append(lines, "║"+strings.Repeat(" ", innerWidth)+"║")
	if len(offer) >= 2 {
		lines = append(lines, "║"+centerIn("Press 1 or 2 to choose an upgrade", innerWidth)+"║")
	} else {
		lines = append(lines, "║"+centerIn("Press 1 or ENTER to accept", innerWidth)+"║")
	}
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
