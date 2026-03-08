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

// SoilTileIDs maps each 4-bit neighbor bitmask to a Pos for trodden paths
// (terrain 14, Soil).
var SoilTileIDs = [16]Pos{
	{13, 10}, {13, 11}, {12, 10}, {14, 7}, //  0:isolated   1:N      2:E      3:N+E
	{13, 9}, {13, 10}, {14, 8}, {13, 10}, //  4:S          5:N+S    6:E+S    7:N+E+S
	{14, 10}, {13, 7}, {13, 10}, {13, 10}, //  8:W          9:N+W   10:E+W   11:N+E+W
	{13, 8}, {13, 10}, {13, 10}, {13, 10}, // 12:S+W       13:N+S+W 14:E+S+W 15:all
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
