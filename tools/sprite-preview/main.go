// sprite-preview displays building and forest sprites for visual tuning.
//
// Run from repo root:
//
//	go run ./tools/sprite-preview
//
// The tool renders each sprite on a 2×2 grass-tile background at the same
// scale and offset used by the game. The house building section reflects the
// crop coordinates defined in render/sprites.go — edit them here to tune, then
// copy the final values back to sprites.go.
//
// Q or Escape to quit.
package main

import (
	"image"
	"image/color"
	_ "image/png"
	"os"

	ebiten "github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"forester/render/spritedata"
)

// ─────────────────────────────────────────────────────────────────────────────
// Sprite crop coordinates are defined in render/spritedata.
// To tune crops, edit the vars in that package and rebuild.
// ─────────────────────────────────────────────────────────────────────────────

const tileSize = 32

func loadSheet(path string) *ebiten.Image {
	f, err := os.Open(path)
	if err != nil {
		panic("run from repo root: " + err.Error())
	}
	defer func() { _ = f.Close() }()
	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}
	return ebiten.NewImageFromImage(img)
}

func cropImg(sheet *ebiten.Image, x, y, w, h int) *ebiten.Image {
	return ebiten.NewImageFromImage(sheet.SubImage(image.Rect(x, y, x+w, y+h)))
}

func cropRect(sheet *ebiten.Image, r image.Rectangle) *ebiten.Image {
	return ebiten.NewImageFromImage(sheet.SubImage(r))
}

// ─────────────────────────────────────────────────────────────────────────────

const (
	cellPad     = 10
	headerH     = 14
	overflowTop = 32                 // pixels above grass background reserved for upward-overflowing sprites
	bgTiles     = 2                  // grass background side in tiles
	bgPx        = bgTiles * tileSize // 64
	labelH      = 14
	sectionGap  = 24
)

// item describes one sprite to display.
// offX / offY are pixel offsets from the cell's grass-top-left corner, matching
// the drawArgs.offsetX / offsetY semantics used in the game renderer.
type item struct {
	label string
	img   *ebiten.Image
	scale float64
	offX  float64
	offY  float64 // negative = overflow above grass background
}

// Game implements ebiten.Game.
type Game struct {
	grass      *ebiten.Image
	forest     []item
	builds     []item
	logW, logH int
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	return nil
}

func (g *Game) drawSection(screen *ebiten.Image, title string, items []item, ox, oy int) {
	ebitenutil.DebugPrintAt(screen, title, ox, oy)

	var opts ebiten.DrawImageOptions
	grassTop := oy + headerH + overflowTop
	x := ox

	for _, it := range items {
		// 2×2 grass background
		for ty := 0; ty < bgTiles; ty++ {
			for tx := 0; tx < bgTiles; tx++ {
				opts.GeoM.Reset()
				opts.GeoM.Translate(float64(x+tx*tileSize), float64(grassTop+ty*tileSize))
				screen.DrawImage(g.grass, &opts)
			}
		}

		// Sprite overlay (same offset semantics as the game renderer)
		if it.img != nil {
			opts.GeoM.Reset()
			opts.GeoM.Scale(it.scale, it.scale)
			opts.GeoM.Translate(float64(x)+it.offX, float64(grassTop)+it.offY)
			screen.DrawImage(it.img, &opts)
		}

		// Label below the grass background
		ebitenutil.DebugPrintAt(screen, it.label, x, grassTop+bgPx+2)

		x += bgPx + cellPad*2
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 0x22, G: 0x22, B: 0x22, A: 0xFF})
	forestW := len(g.forest) * (bgPx + cellPad*2)
	g.drawSection(screen, "FOREST", g.forest, cellPad, cellPad)
	g.drawSection(screen, "BUILDINGS", g.builds, cellPad+forestW+sectionGap, cellPad)
}

func (g *Game) Layout(_, _ int) (w, h int) { return g.logW, g.logH }

// ─────────────────────────────────────────────────────────────────────────────

func main() {
	terrain := loadSheet("assets/sprites/lpc-terrains/terrain-v7.png")
	trees := loadSheet("assets/sprites/lpc-trees/trees-green.png")
	roofSheet := loadSheet("assets/sprites/lpc-thatched-roof-cottage/thatched-roof.png")
	cottageSheet := loadSheet("assets/sprites/lpc-thatched-roof-cottage/cottage.png")
	winDoorSheet := loadSheet("assets/sprites/lpc-windows-doors-v2/windows-doors.png")

	grass := cropImg(terrain, spritedata.GrassRect.Min.X, spritedata.GrassRect.Min.Y,
		spritedata.GrassRect.Dx(), spritedata.GrassRect.Dy())

	houseImg := spritedata.BuildHouseImg(
		cropRect(roofSheet, spritedata.RoofRect),
		cropRect(cottageSheet, spritedata.WallRect),
		cropRect(winDoorSheet, spritedata.DoorRect),
		cropRect(winDoorSheet, spritedata.WindowRect),
	)

	// Forest sprites — offsets match render/sprites.go drawArgs exactly.
	forest := []item{
		{
			label: "stump",
			img:   cropRect(trees, spritedata.TrunkRect),
			scale: 1.0 / 3.0,
		},
		{
			label: "sapling",
			img:   cropRect(trees, spritedata.SaplingRect),
			scale: 0.25,
		},
		{
			label: "young",
			img:   cropRect(trees, spritedata.YoungRect),
			scale: 0.4,
		},
		{
			label: "mature",
			img:   cropRect(trees, spritedata.MatureRect),
			scale: 1.0 / 3.0,
			offY:  -float64(tileSize),
		},
	}

	// Building sprites — offsets match render/sprites.go houseOverlays() exactly.
	builds := []item{
		{
			label: "house (2×2)",
			img:   houseImg,
			scale: 1.0,
			offY:  -float64(tileSize),
		},
	}

	forestW := len(forest) * (bgPx + cellPad*2)
	buildsW := len(builds) * (bgPx + cellPad*2)
	logW := cellPad + forestW + sectionGap + buildsW + cellPad
	logH := cellPad + headerH + overflowTop + bgPx + labelH + cellPad

	ebiten.SetWindowSize(logW*2, logH*2)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("Sprite Preview — Q to quit")

	g := &Game{
		grass:  grass,
		forest: forest,
		builds: builds,
		logW:   logW,
		logH:   logH,
	}
	if err := ebiten.RunGame(g); err != nil && err != ebiten.Termination {
		panic(err)
	}
}
