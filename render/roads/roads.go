// Package roads provides the tile ID mappings and sheet-slicing helper for
// lpc-terrains road and trodden-path rendering.
//
// Bitmask convention: bit0=N  bit1=E  bit2=S  bit3=W.
// A bit is set when the cardinal neighbour in that direction is at the same
// road level or higher.
//
// To change which tile is shown for a given connectivity pattern, edit the
// relevant entry in SoilTileIDs or GravelTileIDs and rebuild.
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
package roads

import (
	"image"

	ebiten "github.com/hajimehoshi/ebiten/v2"
)

// SoilTileIDs maps each 4-bit neighbor bitmask to a terrain-v7.png tile ID
// for trodden paths (terrain 14, Soil).
var SoilTileIDs = [16]int{
	333, 365, 332, 238, //  0:isolated   1:N      2:E      3:N+E
	301, 333, 270, 333, //  4:S          5:N+S    6:E+S    7:N+E+S
	334, 237, 333, 333, //  8:W          9:N+W   10:E+W   11:N+E+W
	269, 333, 333, 333, // 12:S+W       13:N+S+W 14:E+S+W 15:all
}

// GravelTileIDs maps each 4-bit neighbor bitmask to a terrain-v7.png tile ID
// for roads (terrain 18, Gravel_1).
var GravelTileIDs = [16]int{
	345, 377, 344, 250, //  0:isolated   1:N      2:E      3:N+E
	313, 345, 282, 345, //  4:S          5:N+S    6:E+S    7:N+E+S
	346, 249, 345, 345, //  8:W          9:N+W   10:E+W   11:N+E+W
	281, 345, 345, 345, // 12:S+W       13:N+S+W 14:E+S+W 15:all
}

// TileFromSheet returns the 32×32 SubImage for the given tile ID from a
// terrain-v7.png sheet (1024×2048, 32×32 tiles, 32 columns per row).
func TileFromSheet(sheet *ebiten.Image, id int) *ebiten.Image {
	x, y := (id%32)*32, (id/32)*32
	return sheet.SubImage(image.Rect(x, y, x+32, y+32)).(*ebiten.Image)
}
