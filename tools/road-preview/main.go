// road-preview renders all 16 autotile mask combinations for road and
// trodden-path terrain.  Edit the mappings below and run:
//
//	go run ./tools/road-preview   (from repo root)
//
// Each cell shows a 3×3 context: road neighbours where the bitmask bit is set,
// grass everywhere else, and the mapped tile at the centre.
// Q or Escape to quit.
package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"os"

	ebiten "github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// ── Edit these mappings to experiment ────────────────────────────────────────
//
// Bitmask layout: bit0=N  bit1=E  bit2=S  bit3=W
//
// Tile ID → pixel in terrain-v7.png (1024×2048, 32×32 tiles, 32 columns):
//
//	x = (id % 32) * 32
//	y = (id / 32) * 32
//
// Handy Soil tile IDs (terrain 14):
//
//	333=fill  365=N-cap  332=E-cap  301=S-cap  334=W-cap
//	237=NW    238=NE     269=SW     270=SE
//
// Handy Gravel tile IDs (terrain 18):
//
//	345=fill  377=N-cap  344=E-cap  313=S-cap  346=W-cap
//	249=NW    250=NE     281=SW     282=SE
var soilMapping = [16]int{
	333, 365, 332, 238, //  0:isolated   1:N      2:E      3:N+E
	301, 333, 270, 333, //  4:S          5:N+S    6:E+S    7:N+E+S
	334, 237, 333, 333, //  8:W          9:N+W   10:E+W   11:N+E+W
	269, 333, 333, 333, // 12:S+W       13:N+S+W 14:E+S+W 15:all
}

var gravelMapping = [16]int{
	345, 377, 344, 250, //  0:isolated   1:N      2:E      3:N+E
	313, 345, 282, 345, //  4:S          5:N+S    6:E+S    7:N+E+S
	346, 249, 345, 345, //  8:W          9:N+W   10:E+W   11:N+E+W
	281, 345, 345, 345, // 12:S+W       13:N+S+W 14:E+S+W 15:all
}

// ─────────────────────────────────────────────────────────────────────────────

const (
	tileSize = 32
	cellPad  = 6
	labelH   = 14
	gridCols = 4
	gridRows = 4
)

var maskNames = [16]string{
	"isolated", "N", "E", "N+E",
	"S", "N+S", "E+S", "N+E+S",
	"W", "N+W", "E+W", "N+E+W",
	"S+W", "N+S+W", "E+S+W", "all",
}

// Game implements ebiten.Game.
type Game struct {
	sheet    *ebiten.Image
	grassImg *ebiten.Image
}

// tileAt returns a SubImage for the given tile ID from the terrain sheet.
func tileAt(sheet *ebiten.Image, id int) *ebiten.Image {
	x, y := (id%32)*tileSize, (id/32)*tileSize
	return sheet.SubImage(image.Rect(x, y, x+tileSize, y+tileSize)).(*ebiten.Image)
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	return nil
}

// drawSection renders a 4×4 grid of 3×3-context cells for one mapping.
func (g *Game) drawSection(screen *ebiten.Image, mapping [16]int, ox, oy int) {
	cellW := 3*tileSize + cellPad*2
	cellH := 3*tileSize + cellPad*2 + labelH

	// Use mask-15 tile (all neighbours road) as the fill for road neighbours in context.
	fillTile := tileAt(g.sheet, mapping[15])

	var opts ebiten.DrawImageOptions
	for mask := 0; mask < 16; mask++ {
		cellX := ox + (mask%gridCols)*cellW
		cellY := oy + (mask/gridCols)*cellH

		center := tileAt(g.sheet, mapping[mask])
		nBit := (mask>>0)&1 == 1
		eBit := (mask>>1)&1 == 1
		sBit := (mask>>2)&1 == 1
		wBit := (mask>>3)&1 == 1

		for cy := 0; cy < 3; cy++ {
			for cx := 0; cx < 3; cx++ {
				var img *ebiten.Image
				switch {
				case cx == 1 && cy == 1:
					img = center
				case cx == 1 && cy == 0 && nBit:
					img = fillTile
				case cx == 2 && cy == 1 && eBit:
					img = fillTile
				case cx == 1 && cy == 2 && sBit:
					img = fillTile
				case cx == 0 && cy == 1 && wBit:
					img = fillTile
				default:
					img = g.grassImg
				}
				opts.GeoM.Reset()
				opts.GeoM.Translate(
					float64(cellX+cellPad+cx*tileSize),
					float64(cellY+cellPad+cy*tileSize),
				)
				screen.DrawImage(img, &opts)
			}
		}

		label := fmt.Sprintf("%d:%s t%d", mask, maskNames[mask], mapping[mask])
		ebitenutil.DebugPrintAt(screen, label, cellX+cellPad, cellY+cellPad+3*tileSize+2)
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 0x22, G: 0x22, B: 0x22, A: 0xFF})

	cellW := 3*tileSize + cellPad*2
	sectionW := gridCols * cellW
	gap := 16

	ebitenutil.DebugPrintAt(screen, "SOIL — trodden path (level 1)", 0, 0)
	g.drawSection(screen, soilMapping, 0, 14)

	ebitenutil.DebugPrintAt(screen, "GRAVEL — road (level 2)", sectionW+gap, 0)
	g.drawSection(screen, gravelMapping, sectionW+gap, 14)
}

func (g *Game) Layout(outsideW, outsideH int) (w, h int) { return outsideW, outsideH }

func main() {
	f, err := os.Open("assets/sprites/lpc-terrains/terrain-v7.png")
	if err != nil {
		panic("run from repo root: " + err.Error())
	}
	defer func() { _ = f.Close() }()

	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}
	sheet := ebiten.NewImageFromImage(img)

	grassImg := ebiten.NewImage(tileSize, tileSize)
	grassImg.Fill(color.RGBA{R: 0x7E, G: 0xC8, B: 0x50, A: 0xFF})

	cellW := 3*tileSize + cellPad*2
	cellH := 3*tileSize + cellPad*2 + labelH
	sectionW := gridCols * cellW
	winW := sectionW*2 + 16 + 8
	winH := gridRows*cellH + 14 + 8

	ebiten.SetWindowSize(winW, winH)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("Road Autotile Preview — Q to quit")

	g := &Game{sheet: sheet, grassImg: grassImg}
	if err := ebiten.RunGame(g); err != nil && err != ebiten.Termination {
		panic(err)
	}
}
