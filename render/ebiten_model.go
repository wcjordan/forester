package render

import (
	"image/color"
	"time"

	ebiten "github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"forester/game"
	"forester/game/geom"
)

const (
	tileSize     = 32
	screenWidth  = 1280
	screenHeight = 720
)

var colorBackground = color.RGBA{R: 0x1A, G: 0x1A, B: 0x1A, A: 0xFF}

// EbitenGame implements ebiten.Game and renders the world using LPC sprites.
type EbitenGame struct {
	game     *game.Game
	clock    game.Clock
	lastTick time.Time
}

// NewEbitenGame creates an EbitenGame wrapping the given game using the system clock.
func NewEbitenGame(g *game.Game) *EbitenGame {
	return &EbitenGame{game: g, clock: game.RealClock{}}
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

	// Movement: hold-to-move; player's 150ms cooldown throttles actual movement.
	player := e.game.State.Player
	world := e.game.State.World
	switch {
	case ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyArrowUp):
		player.Move(0, -1, world, now)
	case ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyArrowDown):
		player.Move(0, 1, world, now)
	case ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft):
		player.Move(-1, 0, world, now)
	case ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyArrowRight):
		player.Move(1, 0, world, now)
	}

	// Game tick at GameTickInterval cadence.
	if now.Sub(e.lastTick) >= game.GameTickInterval {
		e.game.Tick()
		e.lastTick = now
	}

	return nil
}

// Draw renders the world as a grid of LPC sprites.
func (e *EbitenGame) Draw(screen *ebiten.Image) {
	screen.Fill(colorBackground)

	world := e.game.State.World
	player := e.game.State.Player

	viewW := screenWidth / tileSize
	viewH := screenHeight / tileSize
	vpX := clamp(player.X-viewW/2, 0, max(0, world.Width-viewW))
	vpY := clamp(player.Y-viewH/2, 0, max(0, world.Height-viewH))

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
		opts.GeoM.Translate(screenX, screenY)
		screen.DrawImage(da.img, &opts)
	}

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

			drawSprite(spriteForTile(tile), screenX, screenY)

			if _, ok := villagerPos[geom.Point{X: worldX, Y: worldY}]; ok {
				drawSprite(spriteForVillager(), screenX, screenY)
			}

			if worldX == player.X && worldY == player.Y {
				drawSprite(spriteForPlayer(), screenX, screenY)
			}
		}
	}
}

// Layout returns the logical screen dimensions.
func (e *EbitenGame) Layout(_, _ int) (w, h int) {
	return screenWidth, screenHeight
}
