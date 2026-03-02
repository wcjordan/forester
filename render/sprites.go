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

// scaledSprite extracts a sub-region from src and returns drawArgs with GeoM pre-scaled.
func scaledSprite(src *ebiten.Image, sx, sy, sw, sh int, scale float64) drawArgs {
	subImg := src.SubImage(image.Rect(sx, sy, sx+sw, sy+sh)).(*ebiten.Image)
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(scale, scale)
	return drawArgs{img: subImg, opts: opts}
}

// spriteForTile returns the drawArgs for a world tile (terrain + structure).
func spriteForTile(tile *game.Tile) drawArgs {
	switch tile.Structure {
	case structures.FoundationLogStorage, structures.FoundationHouse:
		return scaledSprite(assets.Dirt, 32, 64, 32, 32, 1.0)
	case structures.LogStorage:
		return scaledSprite(assets.Barrel, 0, 0, 64, 64, 0.5)
	case structures.House:
		return scaledSprite(assets.House, 0, 0, 96, 96, 1.0/3.0)
	}

	switch tile.Terrain {
	case game.Forest:
		switch {
		case tile.TreeSize == 0:
			return scaledSprite(assets.Trunk, 0, 0, 96, 96, 1.0/3.0)
		case tile.TreeSize >= 7:
			return scaledSprite(assets.Treetop, 96, 112, 96, 112, 1.0/3.0)
		case tile.TreeSize >= 4:
			return scaledSprite(assets.Treetop, 0, 0, 96, 112, 1.0/3.0)
		default:
			return scaledSprite(assets.Grass, 0, 0, 32, 32, 1.0)
		}
	default:
		return scaledSprite(assets.GrassTile, 0, 0, 32, 32, 1.0)
	}
}

// spriteForPlayer returns drawArgs for the player character.
func spriteForPlayer() drawArgs {
	return scaledSprite(assets.Player, 0, 128, 64, 64, 0.5)
}

// spriteForVillager returns drawArgs for a villager character.
func spriteForVillager() drawArgs {
	return scaledSprite(assets.Villager, 0, 128, 64, 64, 0.5)
}
