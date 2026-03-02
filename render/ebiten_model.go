package render

import (
	"image/color"
	"time"

	ebiten "github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"forester/game"
	"forester/game/geom"
	"forester/game/structures"
)

const (
	tileSize     = 32
	screenWidth  = 1280
	screenHeight = 720
)

// Color palette for the Ebitengine renderer.
var (
	colorBackground    = color.RGBA{R: 0x1A, G: 0x1A, B: 0x1A, A: 0xFF}
	colorGrassland     = color.RGBA{R: 0x7E, G: 0xC8, B: 0x50, A: 0xFF}
	colorForestDense   = color.RGBA{R: 0x2D, G: 0x6A, B: 0x2D, A: 0xFF}
	colorForestMid     = color.RGBA{R: 0x4A, G: 0x8A, B: 0x4A, A: 0xFF}
	colorForestSapling = color.RGBA{R: 0x6D, G: 0xAA, B: 0x6D, A: 0xFF}
	colorStump         = color.RGBA{R: 0x6B, G: 0x5A, B: 0x3E, A: 0xFF}
	colorFoundation    = color.RGBA{R: 0xD4, G: 0xA8, B: 0x40, A: 0xFF}
	colorLogStorage    = color.RGBA{R: 0xC8, G: 0x92, B: 0x0A, A: 0xFF}
	colorHouse         = color.RGBA{R: 0xA0, G: 0x40, B: 0xC0, A: 0xFF}
	colorPlayer        = color.RGBA{R: 0x40, G: 0x80, B: 0xFF, A: 0xFF}
	colorVillager      = color.RGBA{R: 0x40, G: 0xC0, B: 0xC0, A: 0xFF}
)

// EbitenGame implements ebiten.Game and renders the world using solid-color rectangles.
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

// Draw renders the world as a grid of colored rectangles.
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

	for row := 0; row < viewH; row++ {
		for col := 0; col < viewW; col++ {
			worldX := vpX + col
			worldY := vpY + row
			c := e.tileColor(worldX, worldY, world, player, villagerPos)
			x := float32(col * tileSize)
			y := float32(row * tileSize)
			vector.FillRect(screen, x, y, tileSize, tileSize, c, false)
		}
	}
}

// tileColor returns the color for the tile at (worldX, worldY) using the same
// priority ordering as render/model.go: player > villager > structure > terrain.
func (e *EbitenGame) tileColor(
	worldX, worldY int,
	world *game.World,
	player *game.Player,
	villagerPos map[geom.Point]struct{},
) color.Color {
	if worldX == player.X && worldY == player.Y {
		return colorPlayer
	}

	if _, ok := villagerPos[geom.Point{X: worldX, Y: worldY}]; ok {
		return colorVillager
	}

	tile := world.TileAt(worldX, worldY)
	if tile == nil {
		return colorBackground
	}

	switch tile.Structure {
	case structures.FoundationLogStorage, structures.FoundationHouse:
		return colorFoundation
	case structures.LogStorage:
		return colorLogStorage
	case structures.House:
		return colorHouse
	}

	switch tile.Terrain {
	case game.Forest:
		switch {
		case tile.TreeSize == 0:
			return colorStump
		case tile.TreeSize >= 7:
			return colorForestDense
		case tile.TreeSize >= 4:
			return colorForestMid
		default:
			return colorForestSapling
		}
	default:
		return colorGrassland
	}
}

// Layout returns the logical screen dimensions.
func (e *EbitenGame) Layout(_, _ int) (w, h int) {
	return screenWidth, screenHeight
}
