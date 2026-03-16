// Package spritedata provides source crop rectangles and composition functions
// for building and forest sprites.
//
// Crop rectangles reference the following files (all gitignored; see README):
//
//	GrassRect              — assets/sprites/lpc-terrains/terrain-v7.png (1024×2048)
//	TrunkRect              — assets/sprites/lpc-trees/trees-green.png (1024×1024)
//	SaplingRect            — assets/sprites/lpc-trees/trees-green.png
//	YoungRect              — assets/sprites/lpc-trees/trees-green.png
//	MatureRect             — assets/sprites/lpc-trees/trees-green.png
//	Roof*Rect              — assets/sprites/lpc-thatched-roof-cottage/thatched-roof.png (512×512)
//	Wall*Rect              — assets/sprites/lpc-thatched-roof-cottage/cottage.png (512×512)
//	DoorRect               — assets/sprites/lpc-windows-doors-v2/windows-doors.png (1024×1024)
//	WindowRect             — assets/sprites/lpc-windows-doors-v2/windows-doors.png
//	WindowTopRect          — assets/sprites/lpc-windows-doors-v2/windows-doors.png
//	Container*Rect         — assets/sprites/container-v4_2/container.png (512×2048)
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
	// Roof*Rect crops the yellow/wheat 3D cottage-top piece from thatched-roof.png.
	RoofTopLeftRect     = image.Rect(192, 126, 192+24, 128+32)
	RoofTopRightRect    = image.Rect(262, 126, 262+26, 128+32)
	RoofBottomLeftRect  = image.Rect(192, 126+66, 192+24, 128+66+32)
	RoofBottomRightRect = image.Rect(262, 126+66, 262+26, 128+66+32)
	RoofBottomRect      = image.Rect(120, 64, 120+50, 64+44)
	RoofLeftRect        = image.Rect(88, 0, 88+40, 0+110)
	RoofRightRect       = image.Rect(162, 0, 162+42, 0+110)
	// Wall*Rect crops the building front wall sections from cottage.png (64x96) each.
	WallLeftRect  = image.Rect(0, 0, 0+64, 0+96)
	WallRightRect = image.Rect(32, 0, 32+64, 0+96)
	// DoorRect crops a wooden door from windows-doors.png (32x54).
	DoorRect = image.Rect(16, 768, 16+32, 768+54)
	// WindowRect crops a brown-frame flower-box window from windows-doors.png.
	WindowRect    = image.Rect(480, 32, 480+32, 32+42)
	WindowTopRect = image.Rect(354, 8, 354+28, 8+6)
)

// Container yard (container-v4_2/container.png, 512×2048).
var (
	// ContainerBarrelStackRect crops a 3-barrel pyramid (59×54).
	ContainerBarrelStackRect = image.Rect(195, 3, 254, 57)
	// ContainerBarrelLargeRect crops 2 large barrels side by side (63×61).
	ContainerBarrelLargeRect = image.Rect(193, 67, 256, 128)
	// ContainerBarrelMedRect crops 3 medium barrels grouped (56×64).
	ContainerBarrelMedRect = image.Rect(68, 62, 124, 126)
	// ContainerBarrelClusterRect crops 2 barrels side by side (64×58).
	ContainerBarrelClusterRect = image.Rect(1, 5, 65, 63)
	// ContainerChestRect crops a single wooden chest (64×49).
	ContainerChestRect = image.Rect(0, 256, 64, 305)
)

// BuildLogStorageImg composes the 128×128 log-storage yard image from the
// container sheet. Five container groups are scattered across the 4×4 tile
// footprint with visible ground between them.
//
// The caller draws the result at the NW anchor tile of the 4×4 footprint at
// scale 1.0 with no offset (the image fits exactly within the footprint).
func BuildLogStorageImg(containerSheet *ebiten.Image) *ebiten.Image {
	const w, h = 128, 128
	img := ebiten.NewImage(w, h)
	opts := &ebiten.DrawImageOptions{}

	barrelStack := containerSheet.SubImage(ContainerBarrelStackRect).(*ebiten.Image)
	barrelLarge := containerSheet.SubImage(ContainerBarrelLargeRect).(*ebiten.Image)
	barrelMed := containerSheet.SubImage(ContainerBarrelMedRect).(*ebiten.Image)
	barrelCluster := containerSheet.SubImage(ContainerBarrelClusterRect).(*ebiten.Image)
	chest := containerSheet.SubImage(ContainerChestRect).(*ebiten.Image)

	// Staggered 4-column layout across 4 rows, cycling through the 5 sprite types.
	// Columns at x≈3,35,68,98; rows at y≈3,33,63,96.

	// Row 0
	draw := func(spr *ebiten.Image, sc, tx, ty float64) {
		opts.GeoM.Reset()
		opts.GeoM.Scale(sc, sc)
		opts.GeoM.Translate(tx, ty)
		img.DrawImage(spr, opts)
	}
	draw(barrelStack, 0.45, 3, 3)
	draw(barrelCluster, 0.42, 35, 2)
	draw(barrelLarge, 0.40, 68, 3)
	draw(chest, 0.42, 98, 3)

	// Row 1
	draw(barrelMed, 0.38, 4, 33)
	draw(barrelStack, 0.44, 36, 32)
	draw(chest, 0.42, 68, 32)
	draw(barrelCluster, 0.42, 98, 33)

	// Row 2
	draw(barrelCluster, 0.43, 3, 63)
	draw(barrelLarge, 0.40, 36, 62)
	draw(barrelMed, 0.38, 70, 63)
	draw(barrelStack, 0.44, 99, 62)

	// Row 3 (two sprites, offset for asymmetry)
	draw(barrelLarge, 0.40, 18, 96)
	draw(chest, 0.42, 82, 95)

	return img
}

// BuildHouseImg composes the 64×96 house building image from the three source sheets.
//
// Layout: y=0..48 = thatched roof; y=48..96 = half-timber wall face with a
// centered door and flanking flower-box windows.
//
// The caller draws the result at the NW anchor tile of the 2×2 footprint with
// offsetY = -tileSize so the roof overflows one row above the footprint (same
// pattern as mature trees).
func BuildHouseImg(roofSheet, wallSheet, winDoorSheet *ebiten.Image) *ebiten.Image {
	const bW, bH = 64, 96
	const scale = 0.5
	const windowHOffset, wallHOffset = 55, 48
	const roofTopY = 16.0 // y where roof bottom-half pieces start (roofLeft/Right height * scale / 2)
	img := ebiten.NewImage(bW, bH)
	opts := &ebiten.DrawImageOptions{}

	wallLeftSrc := wallSheet.SubImage(WallLeftRect).(*ebiten.Image)
	wallRightSrc := wallSheet.SubImage(WallRightRect).(*ebiten.Image)
	roofTopLeftSrc := roofSheet.SubImage(RoofTopLeftRect).(*ebiten.Image)
	roofTopRightSrc := roofSheet.SubImage(RoofTopRightRect).(*ebiten.Image)
	roofBottomLeftSrc := roofSheet.SubImage(RoofBottomLeftRect).(*ebiten.Image)
	roofBottomRightSrc := roofSheet.SubImage(RoofBottomRightRect).(*ebiten.Image)
	roofBottomSrc := roofSheet.SubImage(RoofBottomRect).(*ebiten.Image)
	roofLeftSrc := roofSheet.SubImage(RoofLeftRect).(*ebiten.Image)
	roofRightSrc := roofSheet.SubImage(RoofRightRect).(*ebiten.Image)
	doorSrc := winDoorSheet.SubImage(DoorRect).(*ebiten.Image)
	winSrc := winDoorSheet.SubImage(WindowRect).(*ebiten.Image)
	winTopSrc := winDoorSheet.SubImage(WindowTopRect).(*ebiten.Image)

	// Wall left: 32×48 px at x=0, y=wallHOffset.
	opts.GeoM.Reset()
	opts.GeoM.Scale(scale, scale)
	opts.GeoM.Translate(0, wallHOffset)
	img.DrawImage(wallLeftSrc, opts)

	// Wall right: 32×48 px at x=32, y=wallHOffset.
	opts.GeoM.Reset()
	opts.GeoM.Scale(scale, scale)
	opts.GeoM.Translate(32, wallHOffset)
	img.DrawImage(wallRightSrc, opts)

	// Roof top left
	roofTopLeftOffsetX := float64(RoofLeftRect.Dx()) * scale
	opts.GeoM.Reset()
	opts.GeoM.Scale(scale, scale)
	opts.GeoM.Translate(roofTopLeftOffsetX, 0)
	img.DrawImage(roofTopLeftSrc, opts)

	// Roof top right
	roofTopRightOffsetX := roofTopLeftOffsetX + float64(RoofTopLeftRect.Dx())*scale
	opts.GeoM.Reset()
	opts.GeoM.Scale(scale, scale)
	opts.GeoM.Translate(roofTopRightOffsetX, 0)
	img.DrawImage(roofTopRightSrc, opts)

	// Roof bottom left
	opts.GeoM.Reset()
	opts.GeoM.Scale(scale, scale)
	opts.GeoM.Translate(roofTopLeftOffsetX, roofTopY)
	img.DrawImage(roofBottomLeftSrc, opts)

	// Roof bottom right
	opts.GeoM.Reset()
	opts.GeoM.Scale(scale, scale)
	opts.GeoM.Translate(roofTopRightOffsetX, roofTopY)
	img.DrawImage(roofBottomRightSrc, opts)

	// Roof bottom
	opts.GeoM.Reset()
	opts.GeoM.Scale(scale, scale)
	opts.GeoM.Translate(roofTopLeftOffsetX, 32)
	img.DrawImage(roofBottomSrc, opts)

	// Roof left
	opts.GeoM.Reset()
	opts.GeoM.Scale(scale, scale)
	img.DrawImage(roofLeftSrc, opts)

	// Roof right
	roofRightOffsetX := roofTopRightOffsetX + float64(RoofTopRightRect.Dx())*scale
	opts.GeoM.Reset()
	opts.GeoM.Scale(scale, scale)
	opts.GeoM.Translate(roofRightOffsetX, 0)
	img.DrawImage(roofRightSrc, opts)

	// Door: 16×27 px at x=24, y=69.
	opts.GeoM.Reset()
	opts.GeoM.Scale(scale, scale)
	opts.GeoM.Translate(24, 69)
	img.DrawImage(doorSrc, opts)

	// Windows: flanking the door, each with a decorative top strip.
	windowTopH := float64(WindowTopRect.Dy()) * scale
	for _, wx := range []float64{6, float64(bW) - 22} {
		opts.GeoM.Reset()
		opts.GeoM.Scale(scale, scale)
		opts.GeoM.Translate(wx+1, windowHOffset)
		img.DrawImage(winTopSrc, opts)

		opts.GeoM.Reset()
		opts.GeoM.Scale(scale, scale)
		opts.GeoM.Translate(wx, windowHOffset+windowTopH)
		img.DrawImage(winSrc, opts)
	}
	return img
}
