// road-preview renders all 16 autotile mask combinations for road and
// trodden-path terrain.  To change a mapping, edit render/autotile/roads.go
// and rerun:
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
	sheet    *ebiten.Image
	grassImg *ebiten.Image
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

	// Each cardinal neighbour shows the dead-end tile facing back toward the center:
	//   N-neighbour has only a S-connection (bitmask 4)
	//   E-neighbour has only a W-connection (bitmask 8)
	//   S-neighbour has only a N-connection (bitmask 1)
	//   W-neighbour has only a E-connection (bitmask 2)
	nNeighbor := roads.TileFromSheet(g.sheet, mapping[4]) // S-cap
	eNeighbor := roads.TileFromSheet(g.sheet, mapping[8]) // W-cap
	sNeighbor := roads.TileFromSheet(g.sheet, mapping[1]) // N-cap
	wNeighbor := roads.TileFromSheet(g.sheet, mapping[2]) // E-cap

	var opts ebiten.DrawImageOptions
	for mask := 0; mask < 16; mask++ {
		cellX := ox + (mask%gridCols)*cellW
		cellY := oy + (mask/gridCols)*cellH

		center := roads.TileFromSheet(g.sheet, mapping[mask])
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
	g.drawSection(screen, roads.SoilTileIDs, 0, 14)

	ebitenutil.DebugPrintAt(screen, "GRAVEL — road (level 2)", sectionW+gap, 0)
	g.drawSection(screen, roads.GravelTileIDs, sectionW+gap, 14)
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
