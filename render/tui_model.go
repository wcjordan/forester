//go:build !js

package render

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"forester/game"
	"forester/game/geom"
	"forester/game/structures"
)

// foundationProgressStyle returns a lipgloss style whose foreground is the
// same amber→gold color used by the Ebiten progress bar overlay.
func foundationProgressStyle(progress float64) lipgloss.Style {
	r, g, b := FoundationProgressRGB(progress)
	hex := fmt.Sprintf("#%02X%02X%02X", r, g, b)
	return lipgloss.NewStyle().Foreground(lipgloss.Color(hex))
}

// TickMsg is sent each tick interval to drive the game loop.
type TickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(game.GameTickInterval, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

var (
	playerStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))            // blue
	forestStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))             // green
	stumpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))             // dark gray
	logStorageStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Bold(true)  // bold yellow
	houseStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Bold(true)  // bold magenta
	resourceDepotStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true) // bold cyan
	villagerStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))            // cyan
	troddenStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("130"))           // brown/amber
	roadStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))             // dark gray
)

// Model is the bubbletea model for the game. It owns viewport dimensions
// and delegates all game logic to game.Game.
type Model struct {
	game             *game.Game
	termWidth        int
	termHeight       int
	clock            game.Clock
	debugVillager    bool // whether the debug bar is visible
	debugVillagerIdx int  // currently selected villager index
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
		case "ctrl+s":
			m.game.Save()
			return m, nil
		case "ctrl+l":
			m.game.Load()
			return m, nil
		case "ctrl+n":
			m.game.Reset()
			return m, nil
		}

		if m.game.HasPendingOffer() {
			switch msg.String() {
			case "1", "enter":
				m.game.SelectCard(0)
			case "2":
				m.game.SelectCard(1)
			case "3":
				m.game.SelectCard(2)
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
		case "\\":
			m.debugVillager = !m.debugVillager
		case "[":
			if n := m.game.Villagers.Count(); n > 0 {
				m.debugVillagerIdx = (m.debugVillagerIdx - 1 + n) % n
			}
		case "]":
			if n := m.game.Villagers.Count(); n > 0 {
				m.debugVillagerIdx = (m.debugVillagerIdx + 1) % n
			}
		}
	}

	return m, nil
}

// View renders the current game state to a string.
func (m Model) View() string {
	if m.termWidth == 0 || m.termHeight == 0 {
		return ""
	}

	if m.game.HasPendingOffer() {
		return m.renderCardScreen()
	}

	// Status bar occupies the last line; map gets the rest.
	// Debug bar (when visible) and save/load status each take one additional line.
	statusMsg := saveStatusText(m.game.Status.Code)
	statusActive := statusMsg != "" && m.clock.Now().Before(m.game.Status.SetAt.Add(statusDuration))
	mapHeight := m.termHeight - 1
	if m.debugVillager {
		mapHeight--
	}
	if statusActive {
		mapHeight--
	}
	mapWidth := m.termWidth

	world := m.game.State.World
	player := m.game.State.Player

	// Top-left corner of the viewport in world coordinates.
	// Center the viewport on the player, clamped to world bounds.
	vpX := clamp(player.X-mapWidth/2, 0, max(0, world.Width-mapWidth))
	vpY := clamp(player.Y-mapHeight/2, 0, max(0, world.Height-mapHeight))

	// Build a position set for O(1) villager lookup during rendering.
	villagerPos := make(map[geom.Point]struct{}, m.game.Villagers.Count())
	for _, v := range m.game.Villagers.Villagers {
		villagerPos[geom.Point{X: v.X, Y: v.Y}] = struct{}{}
	}

	var sb strings.Builder
	for row := 0; row < mapHeight; row++ {
		for col := 0; col < mapWidth; col++ {
			worldX := vpX + col
			worldY := vpY + row

			if worldX == player.X && worldY == player.Y {
				sb.WriteString(playerStyle.Render("@"))
				continue
			}

			if _, isVillager := villagerPos[geom.Point{X: worldX, Y: worldY}]; isVillager {
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
			case structures.FoundationLogStorage, structures.FoundationHouse, structures.FoundationResourceDepot:
				progress, _ := m.game.State.FoundationProgressAt(geom.Point{X: worldX, Y: worldY})
				sb.WriteString(foundationProgressStyle(progress).Render("?"))
				continue
			case structures.LogStorage:
				sb.WriteString(logStorageStyle.Render("L"))
				continue
			case structures.House:
				sb.WriteString(houseStyle.Render("H"))
				continue
			case structures.ResourceDepot:
				sb.WriteString(resourceDepotStyle.Render("D"))
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
				switch game.RoadLevelFor(tile) {
				case 2:
					sb.WriteString(roadStyle.Render("="))
				case 1:
					sb.WriteString(troddenStyle.Render(":"))
				default:
					sb.WriteByte('.')
				}
			}
		}
		if row < mapHeight-1 {
			sb.WriteByte('\n')
		}
	}

	// Status bar.
	status := fmt.Sprintf(" Player: (%d, %d)  Wood: %d/%d",
		player.X, player.Y, player.Inventory[game.Wood], player.MaxCarry)

	logStored := m.game.Stores.Total(game.Wood)
	logCap := m.game.Stores.TotalCapacity(game.Wood)
	if logCap > 0 {
		status += fmt.Sprintf("  Log: %d/%d", logStored, logCap)
	}

	villagerCount := m.game.Villagers.Count()
	houseCount := world.CountStructureInstances(structures.House)
	if villagerCount > 0 || houseCount > 0 {
		status += fmt.Sprintf("  Villagers: %d/%d", villagerCount, houseCount)
	}

	xp, nextMilestone := m.game.XPInfo()
	status += fmt.Sprintf("  XP: %d/%d", xp, nextMilestone)

	sb.WriteByte('\n')
	sb.WriteString(status)

	if statusActive {
		sb.WriteByte('\n')
		sb.WriteString(" " + statusMsg)
	}
	if m.debugVillager {
		sb.WriteByte('\n')
		sb.WriteString(m.villagerDebugBar())
	}

	return sb.String()
}

// villagerDebugBar returns a one-line debug string for the selected villager.
func (m Model) villagerDebugBar() string {
	villagers := m.game.Villagers.Villagers
	n := len(villagers)
	if n == 0 {
		return " Debug: no villagers"
	}
	idx := clamp(m.debugVillagerIdx, 0, n-1)
	v := villagers[idx]
	return fmt.Sprintf(" Debug V%d/%d  Pos: (%d,%d)  Task: %s  Target: (%d,%d)  Wood: %d/%d",
		idx+1, n, v.X, v.Y, v.Task, v.TargetX, v.TargetY, v.Wood, game.VillagerMaxCarry)
}

// renderCardScreen renders the milestone card selection overlay centered in the terminal.
// When two cards are offered they are displayed side by side horizontally.
func (m Model) renderCardScreen() string {
	offer := m.game.CurrentOffer()
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

	switch {
	case len(offer) >= 3:
		c1 := buildCardLines(offer[0], "1")
		c2 := buildCardLines(offer[1], "2")
		c3 := buildCardLines(offer[2], "3")
		// Normalize all card heights: insert empty lines before the accept line.
		maxLen := max(len(c1), max(len(c2), len(c3)))
		for len(c1) < maxLen {
			c1 = insertEmptyCard(c1, len(c1)-2)
		}
		for len(c2) < maxLen {
			c2 = insertEmptyCard(c2, len(c2)-2)
		}
		for len(c3) < maxLen {
			c3 = insertEmptyCard(c3, len(c3)-2)
		}
		cardLinesList = [][]string{c1, c2, c3}
		innerWidth = cardPad + cardBoxWidth + gap + cardBoxWidth + gap + cardBoxWidth + cardPad
	case len(offer) >= 2:
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
	default:
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
		switch len(cardLinesList) {
		case 3:
			row = strings.Repeat(" ", cardPad) + cardLinesList[0][i] + strings.Repeat(" ", gap) + cardLinesList[1][i] + strings.Repeat(" ", gap) + cardLinesList[2][i] + strings.Repeat(" ", cardPad)
		case 2:
			row = strings.Repeat(" ", cardPad) + cardLinesList[0][i] + strings.Repeat(" ", gap) + cardLinesList[1][i] + strings.Repeat(" ", cardPad)
		default:
			row = strings.Repeat(" ", cardPad) + cardLinesList[0][i] + strings.Repeat(" ", cardPad)
		}
		lines = append(lines, "║"+row+"║")
	}

	lines = append(lines, "║"+strings.Repeat(" ", innerWidth)+"║")
	switch len(cardLinesList) {
	case 3:
		lines = append(lines, "║"+centerIn("Press 1, 2, or 3 to choose an upgrade", innerWidth)+"║")
	case 2:
		lines = append(lines, "║"+centerIn("Press 1 or 2 to choose an upgrade", innerWidth)+"║")
	default:
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
