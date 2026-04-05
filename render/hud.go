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

const (
	foundationBarHeight  = 4
	foundationBarPadding = 2 // pixels above the top edge of the tile
	foundationBarInset   = 2 // pixels inset from left and right edges
)

// drawFoundationOverlays draws a colored progress bar above each active
// foundation that is visible in the current viewport. The bar floats
// foundationBarPadding pixels above the top edge of the foundation footprint
// and spans the full footprint width (minus foundationBarInset on each side).
// Color transitions from dark amber (0%) to bright gold (100%) using the same
// progression as the TUI shading.
func drawFoundationOverlays(screen *ebiten.Image, g *game.Game, camX, camY, zoom float64) {
	scaledTile := float64(tileSize) * zoom
	vpX := int(camX)
	vpY := int(camY)
	fracX := camX - float64(vpX)
	fracY := camY - float64(vpY)
	for _, fi := range g.State.AllFoundationsProgress() {
		sx := float32((float64(fi.Origin.X-vpX)-fracX)*scaledTile) + foundationBarInset
		sy := float32((float64(fi.Origin.Y-vpY)-fracY)*scaledTile) - foundationBarHeight - foundationBarPadding
		if sy < 0 {
			continue // bar would be above the viewport
		}
		barW := float32(float64(fi.Width)*scaledTile) - 2*foundationBarInset
		fillW := barW * float32(fi.Progress)

		// Background.
		vector.FillRect(screen, sx, sy, barW, foundationBarHeight, colorFoundationBarBG, false)

		// Fill using shared amber→gold progression.
		cr, cg, cb := FoundationProgressRGB(fi.Progress)
		if fillW > 0 {
			vector.FillRect(screen, sx, sy, fillW, foundationBarHeight, color.RGBA{R: cr, G: cg, B: cb, A: 220}, false)
		}
	}
}

var (
	colorHUDBG           = color.RGBA{0, 0, 0, 200}
	colorFoundationBarBG = color.RGBA{0, 0, 0, 160}
	colorHUDText         = color.RGBA{255, 255, 255, 255}
	colorOverlay         = color.RGBA{0, 0, 0, 180}
	colorCardBG          = color.RGBA{42, 42, 42, 255}
	colorCardBorder      = color.RGBA{212, 168, 64, 255}
	colorCardTitle       = color.RGBA{255, 255, 255, 255}
	colorCardDesc        = color.RGBA{204, 204, 204, 255}
	colorCardHint        = color.RGBA{255, 215, 0, 255}
	// colorTitlePanelBG intentionally shares the card background color.
	colorTitlePanelBG = colorCardBG
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

	op := &textv2.DrawOptions{}
	op.GeoM.Translate(8, float64(screenH-hudHeight)+4)
	op.ColorScale.ScaleWithColor(colorHUDText)
	textv2.Draw(screen, status, face, op)
}

// drawStatusBar renders a brief save/load status message above the main HUD.
func drawStatusBar(screen *ebiten.Image, msg string, face *textv2.GoXFace, screenW, screenH int) {
	barY := screenH - hudHeight*2
	vector.FillRect(screen,
		0, float32(barY),
		float32(screenW), float32(hudHeight),
		colorHUDBG, false)
	op := &textv2.DrawOptions{}
	op.GeoM.Translate(8, float64(barY)+4)
	op.ColorScale.ScaleWithColor(colorHUDText)
	textv2.Draw(screen, " "+msg, face, op)
}

// drawVillagerDebugBar renders a second HUD row above the main status bar showing
// debug info for the selected villager.
func drawVillagerDebugBar(screen *ebiten.Image, g *game.Game, face *textv2.GoXFace, screenW, screenH, idx int) {
	barY := screenH - hudHeight*2
	vector.FillRect(screen,
		0, float32(barY),
		float32(screenW), float32(hudHeight),
		colorHUDBG, false)

	villagers := g.Villagers.Villagers
	n := len(villagers)
	var text string
	if n == 0 {
		text = " Debug: no villagers"
	} else {
		i := clamp(idx, 0, n-1)
		v := villagers[i]
		text = fmt.Sprintf(" Debug V%d/%d  Pos: (%d,%d)  Task: %s  Target: (%d,%d)  Wood: %d/%d",
			i+1, n, v.X, v.Y, v.Task, v.TargetX, v.TargetY, v.Wood, game.VillagerMaxCarry)
	}

	op := &textv2.DrawOptions{}
	op.GeoM.Translate(8, float64(barY)+4)
	op.ColorScale.ScaleWithColor(colorHUDText)
	textv2.Draw(screen, text, face, op)
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

		// Description lines — respect explicit \n breaks, then word-wrap each paragraph.
		lines := wrapDescription(u.Description(), 26)
		for j, line := range lines {
			op := &textv2.DrawOptions{}
			op.GeoM.Translate(float64(cx)+8, float64(cy)+36+float64(j)*16)
			op.ColorScale.ScaleWithColor(colorCardDesc)
			textv2.Draw(screen, line, face, op)
		}

		// Key hint.
		hint := fmt.Sprintf("[ %d ] Accept", i+1)
		drawCenteredText(screen, hint, face, colorCardHint, int(cx)+cardW/2, int(cy)+cardH-18)
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

// wrapDescription splits text on explicit \n paragraph breaks, then word-wraps
// each paragraph to maxLen characters, preserving intentional line breaks.
func wrapDescription(text string, maxLen int) []string {
	paragraphs := strings.Split(text, "\n")
	var lines []string
	for _, para := range paragraphs {
		lines = append(lines, wrapParagraph(para, maxLen)...)
	}
	return lines
}

// wrapParagraph breaks a single paragraph into lines no longer than maxLen characters.
func wrapParagraph(text string, maxLen int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}
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
