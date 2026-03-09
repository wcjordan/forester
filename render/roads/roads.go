// Package roads provides quadrant-composed autotile mappings and the sheet
// compositor for lpc-terrains road and trodden-path rendering.
//
// Bitmask convention: bit0=N  bit1=E  bit2=S  bit3=W.
// A bit is set when the cardinal neighbour in that direction has any road or
// a building.
//
// Each 32×32 road tile is assembled from four 16×16 quadrants sourced from
// terrain-v7.png (1024×2048, 32-column sheet).  Quadrant addresses are given
// as SubPos{X, Y} where each unit equals 16 pixels.
//
// To change a tile shape, edit the relevant entry in SoilComposed or
// GravelComposed and rebuild.  The encoding rules are documented on those
// vars; the source regions are:
//
//	Soil   — SubPos x:54–59, y:46–51  (32px tile cols 27–29, rows 23–25)
//	Gravel — SubPos x:48–53, y:46–51  (32px tile cols 24–26, rows 23–25)
package roads

import (
	"image"

	ebiten "github.com/hajimehoshi/ebiten/v2"
)

// SubPos is a 16×16 quadrant address within terrain-v7.png.
// Each unit equals 16 pixels; pixel origin = X*16, Y*16.
type SubPos struct{ X, Y int }

// ComposedTile describes a 32×32 tile assembled from four 16×16 quadrants
// sourced from (potentially different) locations in terrain-v7.png.
// Edit the SubPos values in SoilComposed / GravelComposed to mix quadrants.
type ComposedTile struct{ TL, TR, BL, BR SubPos }

// SoilComposed maps each 4-bit bitmask to a ComposedTile for trodden paths.
//
// Quadrants are drawn from the 6×6 SubPos grid at x:54–59, y:46–51 (Soil).
// Edge coordinates encode connectivity:
//
//	x-left:   W=0,E=0→54  W=0,E=1→54  W=1,E=0→58  W=1,E=1→56
//	x-right:  W=0,E=0→59  W=0,E=1→55  W=1,E=0→59  W=1,E=1→57
//	y-top:    N=0,S=0→46  N=0,S=1→46  N=1,S=0→50  N=1,S=1→49
//	y-bottom: N=0,S=0→51  N=0,S=1→47  N=1,S=0→51  N=1,S=1→48
var SoilComposed = [16]ComposedTile{
	/* 0  isolated */ {TL: SubPos{54, 46}, TR: SubPos{59, 46}, BL: SubPos{54, 51}, BR: SubPos{59, 51}},
	/* 1  N        */ {TL: SubPos{54, 50}, TR: SubPos{59, 50}, BL: SubPos{54, 51}, BR: SubPos{59, 51}},
	/* 2  E        */ {TL: SubPos{54, 46}, TR: SubPos{55, 46}, BL: SubPos{54, 51}, BR: SubPos{55, 51}},
	/* 3  N+E      */ {TL: SubPos{54, 50}, TR: SubPos{55, 50}, BL: SubPos{54, 51}, BR: SubPos{55, 51}},
	/* 4  S        */ {TL: SubPos{54, 46}, TR: SubPos{59, 46}, BL: SubPos{54, 47}, BR: SubPos{59, 47}},
	/* 5  N+S      */ {TL: SubPos{54, 49}, TR: SubPos{59, 49}, BL: SubPos{54, 48}, BR: SubPos{59, 48}},
	/* 6  E+S      */ {TL: SubPos{54, 46}, TR: SubPos{55, 46}, BL: SubPos{54, 47}, BR: SubPos{55, 47}},
	/* 7  N+E+S    */ {TL: SubPos{54, 49}, TR: SubPos{55, 49}, BL: SubPos{54, 48}, BR: SubPos{55, 48}},
	/* 8  W        */ {TL: SubPos{58, 46}, TR: SubPos{59, 46}, BL: SubPos{58, 51}, BR: SubPos{59, 51}},
	/* 9  N+W      */ {TL: SubPos{58, 50}, TR: SubPos{59, 50}, BL: SubPos{58, 51}, BR: SubPos{59, 51}},
	/* 10 E+W      */ {TL: SubPos{56, 46}, TR: SubPos{57, 46}, BL: SubPos{56, 51}, BR: SubPos{57, 51}},
	/* 11 N+E+W    */ {TL: SubPos{56, 50}, TR: SubPos{57, 50}, BL: SubPos{56, 51}, BR: SubPos{57, 51}},
	/* 12 S+W      */ {TL: SubPos{58, 46}, TR: SubPos{59, 46}, BL: SubPos{58, 47}, BR: SubPos{59, 47}},
	/* 13 N+S+W    */ {TL: SubPos{58, 49}, TR: SubPos{59, 49}, BL: SubPos{58, 48}, BR: SubPos{59, 48}},
	/* 14 E+S+W    */ {TL: SubPos{56, 46}, TR: SubPos{57, 46}, BL: SubPos{56, 47}, BR: SubPos{57, 47}},
	/* 15 all      */ {TL: SubPos{56, 49}, TR: SubPos{57, 49}, BL: SubPos{56, 48}, BR: SubPos{57, 48}},
}

// GravelComposed maps each 4-bit bitmask to a ComposedTile for roads.
//
// Same edge-connectivity encoding as SoilComposed, applied to the Gravel
// 6×6 SubPos grid at x:48–53, y:46–51 (offset −6 in x, 0 in y from Soil).
var GravelComposed = [16]ComposedTile{
	/* 0  isolated */ {TL: SubPos{48, 46}, TR: SubPos{53, 46}, BL: SubPos{48, 51}, BR: SubPos{53, 51}},
	/* 1  N        */ {TL: SubPos{48, 50}, TR: SubPos{53, 50}, BL: SubPos{48, 51}, BR: SubPos{53, 51}},
	/* 2  E        */ {TL: SubPos{48, 46}, TR: SubPos{49, 46}, BL: SubPos{48, 51}, BR: SubPos{49, 51}},
	/* 3  N+E      */ {TL: SubPos{48, 50}, TR: SubPos{49, 50}, BL: SubPos{48, 51}, BR: SubPos{49, 51}},
	/* 4  S        */ {TL: SubPos{48, 46}, TR: SubPos{53, 46}, BL: SubPos{48, 47}, BR: SubPos{53, 47}},
	/* 5  N+S      */ {TL: SubPos{48, 49}, TR: SubPos{53, 49}, BL: SubPos{48, 48}, BR: SubPos{53, 48}},
	/* 6  E+S      */ {TL: SubPos{48, 46}, TR: SubPos{49, 46}, BL: SubPos{48, 47}, BR: SubPos{49, 47}},
	/* 7  N+E+S    */ {TL: SubPos{48, 49}, TR: SubPos{49, 49}, BL: SubPos{48, 48}, BR: SubPos{49, 48}},
	/* 8  W        */ {TL: SubPos{52, 46}, TR: SubPos{53, 46}, BL: SubPos{52, 51}, BR: SubPos{53, 51}},
	/* 9  N+W      */ {TL: SubPos{52, 50}, TR: SubPos{53, 50}, BL: SubPos{52, 51}, BR: SubPos{53, 51}},
	/* 10 E+W      */ {TL: SubPos{50, 46}, TR: SubPos{51, 46}, BL: SubPos{50, 51}, BR: SubPos{51, 51}},
	/* 11 N+E+W    */ {TL: SubPos{50, 50}, TR: SubPos{51, 50}, BL: SubPos{50, 51}, BR: SubPos{51, 51}},
	/* 12 S+W      */ {TL: SubPos{52, 46}, TR: SubPos{53, 46}, BL: SubPos{52, 47}, BR: SubPos{53, 47}},
	/* 13 N+S+W    */ {TL: SubPos{52, 49}, TR: SubPos{53, 49}, BL: SubPos{52, 48}, BR: SubPos{53, 48}},
	/* 14 E+S+W    */ {TL: SubPos{50, 46}, TR: SubPos{51, 46}, BL: SubPos{50, 47}, BR: SubPos{51, 47}},
	/* 15 all      */ {TL: SubPos{50, 49}, TR: SubPos{51, 49}, BL: SubPos{50, 48}, BR: SubPos{51, 48}},
}

// ComposeFromSheet builds a new 32×32 image by blitting four 16×16 quadrants
// from sheet at the SubPos addresses given by c.
func ComposeFromSheet(sheet *ebiten.Image, c ComposedTile) *ebiten.Image {
	out := ebiten.NewImage(32, 32)
	quads := [4]struct {
		pos    SubPos
		ox, oy int
	}{
		{c.TL, 0, 0},
		{c.TR, 16, 0},
		{c.BL, 0, 16},
		{c.BR, 16, 16},
	}
	var opts ebiten.DrawImageOptions
	for _, q := range quads {
		x, y := q.pos.X*16, q.pos.Y*16
		sub := sheet.SubImage(image.Rect(x, y, x+16, y+16)).(*ebiten.Image)
		opts.GeoM.Reset()
		opts.GeoM.Translate(float64(q.ox), float64(q.oy))
		out.DrawImage(sub, &opts)
	}
	return out
}
