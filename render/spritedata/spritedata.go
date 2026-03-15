// Package spritedata provides source crop rectangles and composition functions
// for building and forest sprites.
//
// Crop rectangles reference the following files (all gitignored; see README):
//
//	GrassRect   — assets/sprites/lpc-terrains/terrain-v7.png (1024×2048)
//	TrunkRect   — assets/sprites/lpc-trees/trees-green.png (1024×1024)
//	SaplingRect — assets/sprites/lpc-trees/trees-green.png
//	YoungRect   — assets/sprites/lpc-trees/trees-green.png
//	MatureRect  — assets/sprites/lpc-trees/trees-green.png
//	RoofRect    — assets/sprites/lpc-thatched-roof-cottage/thatched-roof.png (512×512)
//	WallRect    — assets/sprites/lpc-thatched-roof-cottage/cottage.png (512×512)
//	DoorRect    — assets/sprites/lpc-windows-doors-v2/windows-doors.png (1024×1024)
//	WindowRect  — assets/sprites/lpc-windows-doors-v2/windows-doors.png
//
// To tune crop coordinates, edit the vars here and rebuild. The sprite-preview
// tool (make sprite_preview) gives fast visual feedback without building the
// full game.
package spritedata

import (
	"image"

	ebiten "github.com/hajimehoshi/ebiten/v2"
)

// Terrain / grass.
var (
	// GrassRect crops the grassland tile from terrain-v7.png (32×32).
	GrassRect = image.Rect(224, 384, 224+32, 384+32)
)

// Forest (trees-green.png).
var (
	// TrunkRect crops the stump/trunk sprite (80×50).
	TrunkRect = image.Rect(36, 655, 36+80, 655+50)
	// SaplingRect crops the small sapling sprite (96×128).
	SaplingRect = image.Rect(64, 226, 64+96, 226+128)
	// YoungRect crops the mid-size tree sprite (128×128).
	YoungRect = image.Rect(256, 224, 256+128, 224+128)
	// MatureRect crops the large mature tree sprite (160×192).
	MatureRect = image.Rect(0, 512, 0+160, 512+192)
)

// House building (thatched-roof-cottage + windows-doors).
var (
	// RoofRect crops the yellow/wheat 3D cottage-top piece from thatched-roof.png (160×128).
	// x=32 skips the thin left edge strip so only the peaked roof shape is included.
	RoofRect = image.Rect(32, 0, 192, 128)
	// WallRect crops the mid-wall section from cottage.png (128×64).
	// 2:1 aspect ratio scales cleanly to the 64×32 screen wall face.
	WallRect = image.Rect(0, 64, 128, 128)
	// DoorRect crops a wooden door from windows-doors.png (64×96).
	DoorRect = image.Rect(0, 512, 64, 608)
	// WindowRect crops a brown-frame flower-box window from windows-doors.png (96×64).
	WindowRect = image.Rect(256, 64, 352, 128)
)

// BuildHouseImg composes the 64×96 house building image from pre-cropped source images.
//
// Layout: y=0..64 = thatched roof; y=64..96 = half-timber wall face with a
// centered door and flanking flower-box windows.
//
// The caller draws the result at the NW anchor tile of the 2×2 footprint with
// offsetY = -tileSize so the roof overflows one row above the footprint (same
// pattern as mature trees).
func BuildHouseImg(roofSrc, wallSrc, doorSrc, winSrc *ebiten.Image) *ebiten.Image {
	const bW, bH = 64, 96
	img := ebiten.NewImage(bW, bH)
	opts := &ebiten.DrawImageOptions{}

	// Roof: scale to fill top 64×64 px.
	rb := roofSrc.Bounds()
	opts.GeoM.Reset()
	opts.GeoM.Scale(float64(bW)/float64(rb.Dx()), float64(64)/float64(rb.Dy()))
	img.DrawImage(roofSrc, opts)

	// Wall: scale to fill 64×32 px at y=64 (south row of footprint).
	wb := wallSrc.Bounds()
	opts.GeoM.Reset()
	opts.GeoM.Scale(float64(bW)/float64(wb.Dx()), float64(32)/float64(wb.Dy()))
	opts.GeoM.Translate(0, 64)
	img.DrawImage(wallSrc, opts)

	// Door: 20×28 px, centered horizontally at y=68.
	db := doorSrc.Bounds()
	opts.GeoM.Reset()
	opts.GeoM.Scale(float64(20)/float64(db.Dx()), float64(28)/float64(db.Dy()))
	opts.GeoM.Translate(float64(bW/2-10), 68)
	img.DrawImage(doorSrc, opts)

	// Windows: 18×18 px, flanking the door.
	winb := winSrc.Bounds()
	wsX := float64(18) / float64(winb.Dx())
	wsY := float64(18) / float64(winb.Dy())
	for _, wx := range []float64{4, float64(bW) - 22} {
		opts.GeoM.Reset()
		opts.GeoM.Scale(wsX, wsY)
		opts.GeoM.Translate(wx, 66)
		img.DrawImage(winSrc, opts)
	}
	return img
}
