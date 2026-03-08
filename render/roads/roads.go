// Package roads provides the tile position mappings and sheet-slicing helper
// for lpc-terrains road and trodden-path rendering.
//
// Bitmask convention: bit0=N  bit1=E  bit2=S  bit3=W.
// A bit is set when the cardinal neighbour in that direction is at the same
// road level or higher.
//
// To change which tile is shown for a given connectivity pattern, edit the
// relevant entry in SoilTileIDs or GravelTileIDs and rebuild.
//
// Pos is a tile grid position {col, row} in terrain-v7.png
// (1024×2048, 32×32 tiles, 32 columns).  Pixel origin = col*32, row*32.
//
// SubPos is a 16×16 quadrant address (each unit = 16 px).  A 32×32 tile at
// Pos{X, Y} has quadrants TL={X*2, Y*2}, TR={X*2+1, Y*2}, BL={X*2, Y*2+1},
// BR={X*2+1, Y*2+1}.  Use SubPos/ComposedTile to build tiles from mixed
// quadrant sources.
//
// Soil reference grid (terrain 14, rows 7–11, cols 12–14):
//
//	col 12     col 13     col 14
//	row  7:               {13,7}=NW  {14,7}=NE
//	row  8:               {13,8}=SW  {14,8}=SE
//	row  9:               {13,9}=S-cap
//	row 10:  {12,10}=E-cap {13,10}=fill {14,10}=W-cap
//	row 11:               {13,11}=N-cap
//
// Gravel reference grid (terrain 18, rows 7–11, cols 24–26):
//
//	col 24     col 25     col 26
//	row  7:               {25,7}=NW  {26,7}=NE
//	row  8:               {25,8}=SW  {26,8}=SE
//	row  9:               {25,9}=S-cap
//	row 10:  {24,10}=E-cap {25,10}=fill {26,10}=W-cap
//	row 11:               {25,11}=N-cap
package roads

import (
	"image"

	ebiten "github.com/hajimehoshi/ebiten/v2"
)

// Pos is a tile grid position {col, row} within terrain-v7.png.
type Pos struct{ X, Y int }

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

// SoilTileIDs maps each 4-bit neighbor bitmask to a Pos for trodden paths
// (terrain 14, Soil).
var SoilTileIDs = [16]Pos{
	{28, 25}, {28, 26}, {27, 23}, {29, 22}, //  0:isolated   1:N      2:E      3:N+E
	{28, 24}, {28, 25}, {29, 23}, {28, 25}, //  4:S          5:N+S    6:E+S    7:N+E+S
	{29, 25}, {28, 22}, {28, 25}, {28, 25}, //  8:W          9:N+W   10:E+W   11:N+E+W
	{28, 23}, {28, 25}, {28, 25}, {28, 26}, // 12:S+W       13:N+S+W 14:E+S+W 15:all
}

// GravelTileIDs maps each 4-bit neighbor bitmask to a Pos for roads
// (terrain 18, Gravel_1).
var GravelTileIDs = [16]Pos{
	{25, 10}, {25, 11}, {24, 10}, {26, 7}, //  0:isolated   1:N      2:E      3:N+E
	{25, 9}, {25, 10}, {26, 8}, {25, 10}, //  4:S          5:N+S    6:E+S    7:N+E+S
	{26, 10}, {25, 7}, {25, 10}, {25, 10}, //  8:W          9:N+W   10:E+W   11:N+E+W
	{25, 8}, {25, 10}, {25, 10}, {25, 10}, // 12:S+W       13:N+S+W 14:E+S+W 15:all
}

// TileFromSheet returns the 32×32 SubImage at position p from a terrain-v7.png
// sheet (32×32 tiles, 32 columns per row).
func TileFromSheet(sheet *ebiten.Image, p Pos) *ebiten.Image {
	x, y := p.X*32, p.Y*32
	return sheet.SubImage(image.Rect(x, y, x+32, y+32)).(*ebiten.Image)
}

// PosToComposed converts a 32×32 tile address into a ComposedTile whose four
// quadrants all point to the same source tile (default / no compositing).
func PosToComposed(p Pos) ComposedTile {
	return ComposedTile{
		TL: SubPos{p.X * 2, p.Y * 2},
		TR: SubPos{p.X*2 + 1, p.Y * 2},
		BL: SubPos{p.X * 2, p.Y*2 + 1},
		BR: SubPos{p.X*2 + 1, p.Y*2 + 1},
	}
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
