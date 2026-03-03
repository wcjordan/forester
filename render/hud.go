package render

import (
	"fmt"
	"image/color"
	"strings"

	ebiten "github.com/hajimehoshi/ebiten/v2"
	textv2 "github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"

	"forester/game"
	"forester/game/structures"
)

const hudHeight = 20

var (
	colorHUDBG        = color.RGBA{0, 0, 0, 200}
	colorHUDText      = color.RGBA{255, 255, 255, 255}
	colorOverlay      = color.RGBA{0, 0, 0, 180}
	colorCardBG       = color.RGBA{42, 42, 42, 255}
	colorCardBorder   = color.RGBA{212, 168, 64, 255}
	colorCardTitle    = color.RGBA{255, 255, 255, 255}
	colorCardDesc     = color.RGBA{204, 204, 204, 255}
	colorCardKeyHint  = color.RGBA{255, 215, 0, 255}
	colorTitlePanelBG = color.RGBA{42, 42, 42, 255}
)

// newHUDFace returns a GoXFace wrapping the built-in 7x13 bitmap font.
func newHUDFace() *textv2.GoXFace {
	return textv2.NewGoXFace(basicfont.Face7x13)
}

// drawHUD renders a semi-transparent status bar at the bottom of the screen.
func drawHUD(screen *ebiten.Image, g *game.Game, face *textv2.GoXFace, screenW, screenH int) {
	vector.FillRect(screen,
		0, float32(screenH-hudHeight),
		float32(screenW), float32(hudHeight),
		colorHUDBG, false)

	player := g.State.Player
	world := g.State.World

	status := fmt.Sprintf(" Player: (%d, %d)  Wood: %d/%d",
		player.X, player.Y, player.Inventory[game.Wood], player.MaxCarry)

	logStored := g.Stores.Total(game.Wood)
	logCap := g.Stores.TotalCapacity(game.Wood)
	if logCap > 0 {
		status += fmt.Sprintf("  Log: %d/%d", logStored, logCap)
	}

	villagerCount := g.Villagers.Count()
	houseCount := world.CountStructureInstances(structures.House)
	if villagerCount > 0 || houseCount > 0 {
		status += fmt.Sprintf("  Villagers: %d/%d", villagerCount, houseCount)
	}

	xp, nextMilestone := g.XPInfo()
	status += fmt.Sprintf("  XP: %d/%d", xp, nextMilestone)

	if progress, ok := g.State.FoundationProgress(); ok {
		status += "  " + hudProgressBar(progress)
	}

	op := &textv2.DrawOptions{}
	op.GeoM.Translate(8, float64(screenH-hudHeight)+4)
	op.ColorScale.ScaleWithColor(colorHUDText)
	textv2.Draw(screen, status, face, op)
}

// hudProgressBar renders a text progress bar, e.g. "Building: ████░░░░ 75%".
func hudProgressBar(progress float64) string {
	const width = 8
	filled := clamp(int(progress*width), 0, width)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("Building: %s %d%%", bar, int(progress*100))
}

// drawCardScreen renders a full-screen card selection overlay.
func drawCardScreen(screen *ebiten.Image, offer []game.UpgradeDef, face *textv2.GoXFace, screenW, screenH int) {
	// Dark overlay behind everything.
	vector.FillRect(screen, 0, 0, float32(screenW), float32(screenH), colorOverlay, false)

	// Title panel.
	const titleH = 80
	titleW := float32(screenW) * 0.6
	titleX := (float32(screenW) - titleW) / 2
	titleY := float32(40)
	vector.FillRect(screen, titleX, titleY, titleW, titleH, colorTitlePanelBG, false)
	vector.StrokeRect(screen, titleX, titleY, titleW, titleH, 2, colorCardBorder, false)

	drawCenteredText(screen, "MILESTONE REACHED!", face, colorCardTitle, screenW, int(titleY+titleH/2)-6)

	// Card panels.
	n := len(offer)
	if n == 0 {
		return
	}
	const cardW, cardH = 200, 180
	const cardSpacing = 20
	totalW := n*cardW + (n-1)*cardSpacing
	startX := (screenW - totalW) / 2
	cardY := (screenH - cardH) / 2

	for i, u := range offer {
		cx := float32(startX + i*(cardW+cardSpacing))
		cy := float32(cardY)
		vector.FillRect(screen, cx, cy, cardW, cardH, colorCardBG, false)
		vector.StrokeRect(screen, cx, cy, cardW, cardH, 2, colorCardBorder, false)

		// Card name.
		name := strings.ToUpper(u.Name())
		drawCenteredText(screen, name, face, colorCardTitle, int(cx)+cardW/2, int(cy)+16)

		// Description lines (word-wrap at ~26 chars).
		lines := wrapText(u.Description(), 26)
		for j, line := range lines {
			op := &textv2.DrawOptions{}
			op.GeoM.Translate(float64(cx)+8, float64(cy)+36+float64(j)*16)
			op.ColorScale.ScaleWithColor(colorCardDesc)
			textv2.Draw(screen, line, face, op)
		}

		// Key hint.
		hint := fmt.Sprintf("[ %d ] Accept", i+1)
		drawCenteredText(screen, hint, face, colorCardKeyHint, int(cx)+cardW/2, int(cy)+cardH-18)
	}
}

// drawCenteredText draws text centered horizontally at the given x-midpoint and y.
func drawCenteredText(screen *ebiten.Image, text string, face *textv2.GoXFace, clr color.Color, centerX, y int) {
	w, _ := textv2.Measure(text, face, 0)
	op := &textv2.DrawOptions{}
	op.GeoM.Translate(float64(centerX)-w/2, float64(y))
	op.ColorScale.ScaleWithColor(clr)
	textv2.Draw(screen, text, face, op)
}

// wrapText naively breaks text into lines no longer than maxLen characters.
func wrapText(text string, maxLen int) []string {
	words := strings.Fields(text)
	var lines []string
	var current strings.Builder
	for _, w := range words {
		if current.Len() > 0 && current.Len()+1+len(w) > maxLen {
			lines = append(lines, current.String())
			current.Reset()
		}
		if current.Len() > 0 {
			current.WriteByte(' ')
		}
		current.WriteString(w)
	}
	if current.Len() > 0 {
		lines = append(lines, current.String())
	}
	return lines
}
