// road-preview renders all 16 autotile mask combinations for road and
// trodden-path terrain using quadrant-composed tiles.
//
// To change a mapping, edit SoilComposed / GravelComposed in
// render/roads/roads.go and rerun:
//
//	go run ./tools/road-preview   (from repo root)
//
// Each cell shows a 3×3 context: road neighbours where the bitmask bit is set,
// grass everywhere else, and the composed tile at the centre.
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

	"forester/render/roads"
)

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
	grassImg       *ebiten.Image
	soilComposed   [16]*ebiten.Image
	gravelComposed [16]*ebiten.Image
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) || inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	return nil
}

// drawComposedSection renders a 4×4 grid of 3×3-context cells for a
// ComposedTile-based mapping.  Neighbour tiles are taken from the composed
// images for masks 4 (S-cap), 8 (W-cap), 1 (N-cap), 2 (E-cap).
func (g *Game) drawComposedSection(screen *ebiten.Image, tiles [16]*ebiten.Image, mapping [16]roads.ComposedTile, ox, oy int) {
	cellW := 3*tileSize + cellPad*2
	cellH := 3*tileSize + cellPad*2 + labelH

	nNeighbor := tiles[4] // S-cap
	eNeighbor := tiles[8] // W-cap
	sNeighbor := tiles[1] // N-cap
	wNeighbor := tiles[2] // E-cap

	var opts ebiten.DrawImageOptions
	for mask := 0; mask < 16; mask++ {
		cellX := ox + (mask%gridCols)*cellW
		cellY := oy + (mask/gridCols)*cellH

		center := tiles[mask]
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
					img = nNeighbor
				case cx == 2 && cy == 1 && eBit:
					img = eNeighbor
				case cx == 1 && cy == 2 && sBit:
					img = sNeighbor
				case cx == 0 && cy == 1 && wBit:
					img = wNeighbor
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

		c := mapping[mask]
		label := fmt.Sprintf("%d:%s TL{%d,%d}", mask, maskNames[mask], c.TL.X, c.TL.Y)
		ebitenutil.DebugPrintAt(screen, label, cellX+cellPad, cellY+cellPad+3*tileSize+2)
	}
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 0x22, G: 0x22, B: 0x22, A: 0xFF})

	cellW := 3*tileSize + cellPad*2
	sectionW := gridCols * cellW
	gap := 16

	ebitenutil.DebugPrintAt(screen, "SOIL composed (edit SoilComposed in roads.go)", 0, 0)
	g.drawComposedSection(screen, g.soilComposed, roads.SoilComposed, 0, 14)

	ebitenutil.DebugPrintAt(screen, "GRAVEL composed (edit GravelComposed in roads.go)", sectionW+gap, 0)
	g.drawComposedSection(screen, g.gravelComposed, roads.GravelComposed, sectionW+gap, 14)
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

	// Pre-compose all tiles from the quadrant mappings.
	var soilComposed [16]*ebiten.Image
	var gravelComposed [16]*ebiten.Image
	for i, c := range roads.SoilComposed {
		soilComposed[i] = roads.ComposeFromSheet(sheet, c)
	}
	for i, c := range roads.GravelComposed {
		gravelComposed[i] = roads.ComposeFromSheet(sheet, c)
	}

	cellH := 3*tileSize + cellPad*2 + labelH
	sectionW := gridCols * (3*tileSize + cellPad*2)
	gap := 16
	winW := sectionW*2 + gap + 8
	winH := gridRows*cellH + 14 + 8

	ebiten.SetWindowSize(winW, winH)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowTitle("Road Autotile Preview — Q to quit")

	g := &Game{
		grassImg:       grassImg,
		soilComposed:   soilComposed,
		gravelComposed: gravelComposed,
	}
	if err := ebiten.RunGame(g); err != nil && err != ebiten.Termination {
		panic(err)
	}
}
