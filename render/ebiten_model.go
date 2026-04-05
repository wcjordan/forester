package render

import (
	"image/color"
	"math"
	"time"

	ebiten "github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	textv2 "github.com/hajimehoshi/ebiten/v2/text/v2"

	"forester/game"
	"forester/game/geom"
)

const tileSize = 32

const (
	zoomMin     = 0.25
	zoomMax     = 4.0
	zoomKeyStep = 0.02 // per-frame zoom delta while key held (~1.2× per second at 60 fps)
)

var colorBackground = color.RGBA{R: 0x1A, G: 0x1A, B: 0x1A, A: 0xFF}

// EbitenGame implements ebiten.Game and renders the world using LPC sprites.
type EbitenGame struct {
	game             *game.Game
	clock            game.Clock
	lastTick         time.Time
	camX             float64
	camY             float64
	screenW          int
	screenH          int
	zoom             float64
	prevPinchDist    float64 // distance between two touch points last frame; 0 = no active pinch
	hudFace          *textv2.GoXFace
	debugVillager    bool
	debugVillagerIdx int
	playerMoving     bool      // true while any movement key is held
	animTick         int       // increments each Update while playerMoving; resets to 0 when idle
	slashCycleStart  time.Time // wall-clock start of the current slash cycle; zero = inactive
	thrustCycleStart time.Time // wall-clock start of the current thrust cycle; zero = inactive
}

// NewEbitenGame creates an EbitenGame wrapping the given game using the system clock.
func NewEbitenGame(g *game.Game) *EbitenGame {
	zoom := g.ZoomLevel
	if zoom == 0 {
		zoom = 1.0
	}
	return &EbitenGame{
		game:    g,
		clock:   game.RealClock{},
		screenW: 1280,
		screenH: 720,
		zoom:    zoom,
		hudFace: newHUDFace(),
	}
}

// applyZoom multiplies the current zoom by delta and clamps to [zoomMin, zoomMax].
func (e *EbitenGame) applyZoom(delta float64) {
	e.zoom = clampF(e.zoom*delta, zoomMin, zoomMax)
}

// saveGame syncs zoom into game state then saves.
func (e *EbitenGame) saveGame() {
	e.game.ZoomLevel = e.zoom
	e.game.Save()
}

// loadGame loads game state then reads zoom back (0 → 1.0 default).
func (e *EbitenGame) loadGame() {
	e.game.Load()
	if e.game.ZoomLevel == 0 {
		e.zoom = 1.0
	} else {
		e.zoom = e.game.ZoomLevel
	}
}

// Update is called every frame (~60/s) by Ebitengine. Handles input and game ticks.
func (e *EbitenGame) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		return ebiten.Termination
	}

	now := e.clock.Now()

	if e.game.HasPendingOffer() {
		switch {
		case inpututil.IsKeyJustPressed(ebiten.Key1):
			e.game.SelectCard(0)
		case inpututil.IsKeyJustPressed(ebiten.Key2):
			e.game.SelectCard(1)
		case inpututil.IsKeyJustPressed(ebiten.Key3):
			e.game.SelectCard(2)
		case inpututil.IsKeyJustPressed(ebiten.KeyEnter):
			e.game.SelectCard(0)
		}
		return nil
	}

	// Ctrl shortcuts: save, load, new game. Return early to skip movement.
	if ebiten.IsKeyPressed(ebiten.KeyControl) {
		switch {
		case inpututil.IsKeyJustPressed(ebiten.KeyS):
			e.saveGame()
		case inpututil.IsKeyJustPressed(ebiten.KeyL):
			e.loadGame()
		case inpututil.IsKeyJustPressed(ebiten.KeyN):
			e.game.Reset()
		}
		return nil
	}

	// Zoom: scroll wheel, +/- keys, and two-finger pinch.
	if _, dy := ebiten.Wheel(); dy != 0 {
		e.applyZoom(math.Pow(1.1, dy))
	}
	if ebiten.IsKeyPressed(ebiten.KeyEqual) || ebiten.IsKeyPressed(ebiten.KeyKPAdd) {
		e.applyZoom(1 + zoomKeyStep)
	}
	if ebiten.IsKeyPressed(ebiten.KeyMinus) || ebiten.IsKeyPressed(ebiten.KeyKPSubtract) {
		e.applyZoom(1 - zoomKeyStep)
	}
	touchIDs := ebiten.AppendTouchIDs(nil)
	if len(touchIDs) == 2 {
		x0, y0 := ebiten.TouchPosition(touchIDs[0])
		x1, y1 := ebiten.TouchPosition(touchIDs[1])
		dx := float64(x1 - x0)
		dy := float64(y1 - y0)
		dist := math.Sqrt(dx*dx + dy*dy)
		if e.prevPinchDist > 0 && dist > 0 {
			e.applyZoom(dist / e.prevPinchDist)
		}
		e.prevPinchDist = dist
	} else {
		e.prevPinchDist = 0
	}

	// Movement: hold-to-move; player's 150ms cooldown throttles actual movement.
	// Also tracks whether any movement key is held for walk animation.
	player := e.game.State.Player
	world := e.game.State.World
	e.playerMoving = false
	switch {
	case ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyArrowUp):
		player.Move(0, -1, world, now)
		e.playerMoving = true
	case ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyArrowDown):
		player.Move(0, 1, world, now)
		e.playerMoving = true
	case ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft):
		player.Move(-1, 0, world, now)
		e.playerMoving = true
	case ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyArrowRight):
		player.Move(1, 0, world, now)
		e.playerMoving = true
	}
	if e.playerMoving {
		e.animTick++
	} else {
		e.animTick = 0
	}

	// Debug villager bar toggle and selection.
	if inpututil.IsKeyJustPressed(ebiten.KeyBackslash) {
		e.debugVillager = !e.debugVillager
	}
	if n := e.game.Villagers.Count(); n > 0 {
		if inpututil.IsKeyJustPressed(ebiten.KeyBracketLeft) {
			e.debugVillagerIdx = (e.debugVillagerIdx - 1 + n) % n
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyBracketRight) {
			e.debugVillagerIdx = (e.debugVillagerIdx + 1) % n
		}
	}

	// Update camera with lerp toward player.
	viewW := e.screenW / tileSize
	viewH := e.screenH / tileSize
	targetCamX := clampF(float64(player.X)-float64(viewW)/2, 0, float64(max(0, world.Width-viewW)))
	targetCamY := clampF(float64(player.Y)-float64(viewH)/2, 0, float64(max(0, world.Height-viewH)))
	e.camX += (targetCamX - e.camX) * 0.12
	e.camY += (targetCamY - e.camY) * 0.12

	// Game tick at GameTickInterval cadence.
	if now.Sub(e.lastTick) >= game.GameTickInterval {
		e.game.Tick()
		e.lastTick = now
	}

	// Advance animation cycle anchors, looping seamlessly while the action continues.
	// If the action fired during the current cycle, restart at the cycle boundary so
	// there is no idle gap. If the action fired after the cycle ended, restart from
	// that action time. Looping stops naturally once no new action occurs in a cycle.
	if ha := player.LastHarvestAt; !ha.IsZero() {
		cycleEnd := e.slashCycleStart.Add(slashAnimDuration)
		if ha.After(cycleEnd) {
			e.slashCycleStart = ha
		} else if ha.After(e.slashCycleStart) && now.After(cycleEnd) {
			e.slashCycleStart = cycleEnd
		}
	}
	if ta := player.LastThrustAt; !ta.IsZero() {
		cycleEnd := e.thrustCycleStart.Add(thrustAnimDuration)
		if ta.After(cycleEnd) {
			e.thrustCycleStart = ta
		} else if ta.After(e.thrustCycleStart) && now.After(cycleEnd) {
			e.thrustCycleStart = cycleEnd
		}
	}

	return nil
}

// Draw renders the world as a grid of LPC sprites.
func (e *EbitenGame) Draw(screen *ebiten.Image) {
	screen.Fill(colorBackground)

	if e.game.HasPendingOffer() {
		drawCardScreen(screen, e.game.CurrentOffer(), e.hudFace, e.screenW, e.screenH)
		return
	}

	world := e.game.State.World
	player := e.game.State.Player

	// Compute player animation frame for this draw call.
	playerDir := dirFrom(player.FacingDX, player.FacingDY)
	now := e.clock.Now()
	playerBaseRow, playerFrame, playerSlash128 := playerAnimFrame(e.slashCycleStart, e.thrustCycleStart, now, e.playerMoving, e.animTick)

	vpX := int(e.camX)
	vpY := int(e.camY)
	viewW := e.screenW / tileSize
	viewH := e.screenH / tileSize

	// Build villager position set for O(1) lookup.
	villagerPos := make(map[geom.Point]struct{}, e.game.Villagers.Count())
	for _, v := range e.game.Villagers.Villagers {
		villagerPos[geom.Point{X: v.X, Y: v.Y}] = struct{}{}
	}

	// Single opts reused for every DrawImage call — reset before each use.
	var opts ebiten.DrawImageOptions

	drawSprite := func(da drawArgs, screenX, screenY float64) {
		opts.GeoM.Reset()
		opts.GeoM.Scale(da.scale, da.scale)
		opts.GeoM.Translate(screenX+da.offsetX, screenY+da.offsetY)
		screen.DrawImage(da.img, &opts)
	}

	// Pass 1: terrain bases. All base tiles are painted before any sprite overlay
	// so that overflowing sprites (e.g. mature tree canopy) are never masked by a
	// neighbouring tile's ground layer.
	for row := 0; row < viewH; row++ {
		for col := 0; col < viewW; col++ {
			worldX := vpX + col
			worldY := vpY + row
			tile := world.TileAt(worldX, worldY)
			if tile == nil {
				continue
			}
			base, _ := spriteForTile(tile, world, worldX, worldY)
			drawSprite(base, float64(col*tileSize), float64(row*tileSize))
		}
	}

	// Pass 2: sprite overlays (trees, structures, villagers, player).
	for row := 0; row < viewH; row++ {
		for col := 0; col < viewW; col++ {
			worldX := vpX + col
			worldY := vpY + row
			tile := world.TileAt(worldX, worldY)
			if tile == nil {
				continue
			}

			screenX := float64(col * tileSize)
			screenY := float64(row * tileSize)

			_, overlays := spriteForTile(tile, world, worldX, worldY)
			for _, da := range overlays {
				drawSprite(da, screenX, screenY)
			}

			if _, ok := villagerPos[geom.Point{X: worldX, Y: worldY}]; ok {
				drawSprite(spriteForVillager(), screenX, screenY)
			}

			if worldX == player.X && worldY == player.Y {
				drawSprite(spriteForPlayer(playerBaseRow, playerDir, playerFrame, playerSlash128), screenX, screenY)
			}
		}
	}

	drawFoundationOverlays(screen, e.game, vpX, vpY)
	drawHUD(screen, e.game, e.hudFace, e.screenW, e.screenH)
	if e.debugVillager {
		drawVillagerDebugBar(screen, e.game, e.hudFace, e.screenW, e.screenH, e.debugVillagerIdx)
	}
	if msg := saveStatusText(e.game.Status.Code); msg != "" && now.Before(e.game.Status.SetAt.Add(statusDuration)) {
		drawStatusBar(screen, msg, e.hudFace, e.screenW, e.screenH)
	}
}

// Layout stores the current window dimensions and returns them as the logical screen size.
func (e *EbitenGame) Layout(outsideW, outsideH int) (w, h int) {
	e.screenW = outsideW
	e.screenH = outsideH
	return outsideW, outsideH
}
