package render

import (
	"image"

	ebiten "github.com/hajimehoshi/ebiten/v2"

	"forester/assets"
	"forester/game"
	"forester/game/structures"
)

// drawArgs bundles a sub-image and pre-scaled draw options for a sprite.
// The caller must add position translation before drawing.
type drawArgs struct {
	img  *ebiten.Image
	opts *ebiten.DrawImageOptions
}

// Pre-sliced sprite frames cached at package scope to avoid repeated SubImage calls
// in the per-tile draw loop.
var (
	// Structures / terrain
	dirtFoundationImg  = assets.Dirt.SubImage(image.Rect(32, 64, 32+32, 64+32)).(*ebiten.Image)
	barrelLogStorageImg = assets.Barrel.SubImage(image.Rect(0, 0, 0+64, 0+64)).(*ebiten.Image)
	houseImg           = assets.House.SubImage(image.Rect(0, 0, 0+96, 0+96)).(*ebiten.Image)
	trunkSmallImg      = assets.Trunk.SubImage(image.Rect(0, 0, 0+96, 0+96)).(*ebiten.Image)
	treetopMatureImg   = assets.Treetop.SubImage(image.Rect(96, 112, 96+96, 112+112)).(*ebiten.Image)
	treetopYoungImg    = assets.Treetop.SubImage(image.Rect(0, 0, 0+96, 0+112)).(*ebiten.Image)
	grassForestImg     = assets.Grass.SubImage(image.Rect(0, 0, 0+32, 0+32)).(*ebiten.Image)
	grassTileImg       = assets.GrassTile.SubImage(image.Rect(0, 0, 0+32, 0+32)).(*ebiten.Image)

	// Characters
	playerImg   = assets.Player.SubImage(image.Rect(0, 128, 0+64, 128+64)).(*ebiten.Image)
	villagerImg = assets.Villager.SubImage(image.Rect(0, 128, 0+64, 128+64)).(*ebiten.Image)
)

// scaledSprite returns drawArgs for a pre-sliced sprite image with GeoM pre-scaled.
func scaledSprite(img *ebiten.Image, scale float64) drawArgs {
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(scale, scale)
	return drawArgs{img: img, opts: opts}
}

// spriteForTile returns the drawArgs for a world tile (terrain + structure).
func spriteForTile(tile *game.Tile) drawArgs {
	switch tile.Structure {
	case structures.FoundationLogStorage, structures.FoundationHouse:
		return scaledSprite(dirtFoundationImg, 1.0)
	case structures.LogStorage:
		return scaledSprite(barrelLogStorageImg, 0.5)
	case structures.House:
		return scaledSprite(houseImg, 1.0/3.0)
	}

	switch tile.Terrain {
	case game.Forest:
		switch {
		case tile.TreeSize == 0:
			return scaledSprite(trunkSmallImg, 1.0/3.0)
		case tile.TreeSize >= 7:
			return scaledSprite(treetopMatureImg, 1.0/3.0)
		case tile.TreeSize >= 4:
			return scaledSprite(treetopYoungImg, 1.0/3.0)
		default:
			return scaledSprite(grassForestImg, 1.0)
		}
	default:
		return scaledSprite(grassTileImg, 1.0)
	}
}

// spriteForPlayer returns drawArgs for the player character.
func spriteForPlayer() drawArgs {
	return scaledSprite(playerImg, 0.5)
}

// spriteForVillager returns drawArgs for a villager character.
func spriteForVillager() drawArgs {
	return scaledSprite(villagerImg, 0.5)
}
