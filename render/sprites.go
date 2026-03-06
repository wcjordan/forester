package render

import (
	"image"

	ebiten "github.com/hajimehoshi/ebiten/v2"

	"forester/assets"
	"forester/game"
	"forester/game/structures"
)

// drawArgs bundles a pre-sliced sprite image and its display scale.
// Draw() owns a single ebiten.DrawImageOptions that it resets per use.
type drawArgs struct {
	img   *ebiten.Image
	scale float64
}

// Pre-sliced sprite frames cached at package scope to avoid repeated SubImage calls
// in the per-tile draw loop.
var (
	// Structures / terrain
	dirtFoundationImg   = assets.Dirt.SubImage(image.Rect(32, 64, 32+32, 64+32)).(*ebiten.Image)
	barrelLogStorageImg = assets.Barrel.SubImage(image.Rect(0, 0, 0+64, 0+64)).(*ebiten.Image)
	houseImg            = assets.House.SubImage(image.Rect(0, 0, 0+96, 0+96)).(*ebiten.Image)
	trunkSmallImg       = assets.Trunk.SubImage(image.Rect(0, 0, 0+96, 0+96)).(*ebiten.Image)
	treetopMatureImg    = assets.Treetop.SubImage(image.Rect(96, 112, 96+96, 112+112)).(*ebiten.Image)
	treetopYoungImg     = assets.Treetop.SubImage(image.Rect(0, 0, 0+96, 0+112)).(*ebiten.Image)
	grassForestImg      = assets.Grass.SubImage(image.Rect(0, 0, 0+32, 0+32)).(*ebiten.Image)
	grassTileImg        = assets.GrassTile.SubImage(image.Rect(0, 0, 0+32, 0+32)).(*ebiten.Image)
	troddenPathImg      = assets.TroddenPath.SubImage(image.Rect(0, 0, 32, 32)).(*ebiten.Image)
	roadImg             = assets.Road.SubImage(image.Rect(0, 0, 32, 32)).(*ebiten.Image)

	// Characters
	villagerImg = assets.Villager.SubImage(image.Rect(0, 128, 0+64, 128+64)).(*ebiten.Image)
)

// Universal LPC spritesheet constants (64×64 px per frame).
// Row groups each have 4 rows: +0=up, +1=left, +2=down, +3=right.
const (
	lpcFrameSize     = 64
	lpcWalkBaseRow   = 8  // rows 8–11
	lpcThrustBaseRow = 4  // rows 4–7
	lpcSlashBaseRow  = 12 // rows 12–15
)

// dirFrom converts a facing vector to a direction index for spritesheet row selection.
// Returns 0=up, 1=left, 2=down, 3=right. Defaults to down for a zero vector.
func dirFrom(dx, dy int) int {
	switch {
	case dy < 0:
		return 0 // up
	case dx < 0:
		return 1 // left
	case dx > 0:
		return 3 // right
	default:
		return 2 // down
	}
}

// spriteForTile returns the drawArgs for a world tile (terrain + structure).
func spriteForTile(tile *game.Tile) drawArgs {
	switch tile.Structure {
	case structures.FoundationLogStorage, structures.FoundationHouse:
		return drawArgs{dirtFoundationImg, 1.0}
	case structures.LogStorage:
		return drawArgs{barrelLogStorageImg, 0.5}
	case structures.House:
		return drawArgs{houseImg, 1.0 / 3.0}
	}

	switch tile.Terrain {
	case game.Forest:
		switch {
		case tile.TreeSize == 0:
			return drawArgs{trunkSmallImg, 1.0 / 3.0}
		case tile.TreeSize >= 7:
			return drawArgs{treetopMatureImg, 1.0 / 3.0}
		case tile.TreeSize >= 4:
			return drawArgs{treetopYoungImg, 1.0 / 3.0}
		default:
			return drawArgs{grassForestImg, 1.0}
		}
	default:
		switch game.RoadLevelFor(tile) {
		case 2:
			return drawArgs{roadImg, 1.0}
		case 1:
			return drawArgs{troddenPathImg, 1.0}
		default:
			return drawArgs{grassTileImg, 1.0}
		}
	}
}

// spriteForPlayer returns drawArgs for the player, selecting the correct frame
// from the Universal LPC spritesheet.
// baseRow selects the animation group (lpcWalkBaseRow, lpcSlashBaseRow, etc.);
// dir selects the row within that group (0=up,1=left,2=down,3=right);
// frame selects the column (0-based).
func spriteForPlayer(baseRow, dir, frame int) drawArgs {
	row := baseRow + dir
	x := frame * lpcFrameSize
	y := row * lpcFrameSize
	img := assets.PlayerSheet.SubImage(image.Rect(x, y, x+lpcFrameSize, y+lpcFrameSize)).(*ebiten.Image)
	return drawArgs{img, 0.5}
}

// spriteForVillager returns drawArgs for a villager character.
func spriteForVillager() drawArgs {
	return drawArgs{villagerImg, 0.5}
}
