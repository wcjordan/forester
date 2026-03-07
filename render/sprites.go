package render

import (
	"image"
	"time"

	ebiten "github.com/hajimehoshi/ebiten/v2"

	"forester/assets"
	"forester/game"
	"forester/game/structures"
)

// drawArgs bundles a pre-sliced sprite image, its display scale, and optional
// pixel offsets applied after scaling (for oversized frames that extend beyond a tile).
// Draw() owns a single ebiten.DrawImageOptions that it resets per use.
type drawArgs struct {
	img     *ebiten.Image
	scale   float64
	offsetX float64
	offsetY float64
}

// terrainTile slices a single 32×32 tile from the lpc-terrains sheet by tile ID.
// The sheet is 32 tiles wide; tile pixel origin = (id%32 * 32, id/32 * 32).
func terrainTile(id int) *ebiten.Image {
	x, y := (id%32)*32, (id/32)*32
	return assets.TerrainSheet.SubImage(image.Rect(x, y, x+32, y+32)).(*ebiten.Image)
}

// Pre-sliced sprite frames cached at package scope to avoid repeated SubImage calls
// in the per-tile draw loop.
var (
	// Structures / terrain
	dirtFoundationImg   = assets.Dirt.SubImage(image.Rect(32, 64, 32+32, 64+32)).(*ebiten.Image)
	barrelLogStorageImg = assets.Barrel.SubImage(image.Rect(0, 0, 0+64, 0+64)).(*ebiten.Image)
	houseImg            = assets.House.SubImage(image.Rect(0, 0, 0+96, 0+96)).(*ebiten.Image)
	grassTileImg        = assets.GrassTile.SubImage(image.Rect(0, 0, 0+32, 0+32)).(*ebiten.Image)

	// Road autotile arrays indexed by 4-bit neighbor bitmask (bit0=N, bit1=E, bit2=S, bit3=W).
	// Corner assignment: TL=road if N|W, TR=road if N|E, BL=road if S|W, BR=road if S|E.
	// Tile IDs verified against terrain-v7.tsx (terrain 14=Soil, terrain 18=Gravel_1).
	soilAutotile   [16]*ebiten.Image // trodden path (level 1)
	gravelAutotile [16]*ebiten.Image // road (level 2)

	// lpc-trees: sapling (128×96), young (128×128), mature (160×192), trunk (80x50)
	lpcTreesSaplingImg = assets.TreesGreen.SubImage(image.Rect(64, 226, 64+96, 226+128)).(*ebiten.Image)
	lpcTreesYoungImg   = assets.TreesGreen.SubImage(image.Rect(256, 224, 256+128, 224+128)).(*ebiten.Image)
	lpcTreesMatureImg  = assets.TreesGreen.SubImage(image.Rect(0, 512, 0+160, 512+192)).(*ebiten.Image)
	lpcTreesTrunkImg   = assets.TreesGreen.SubImage(image.Rect(36, 655, 36+80, 655+50)).(*ebiten.Image)

	// Characters
	villagerImg = assets.Villager.SubImage(image.Rect(0, 128, 0+64, 128+64)).(*ebiten.Image)

	// Player animation frames: [direction][frame], direction 0=up 1=left 2=down 3=right.
	playerWalkFrames     [4][8]*ebiten.Image
	playerThrustFrames   [4][8]*ebiten.Image
	playerSlash128Frames [4][lpcSlash128Frames]*ebiten.Image
)

func init() {
	// Soil autotile (terrain 14, trodden paths). Mapping derived from corner rules:
	// bitmask → tile ID (center tile 333 used for all-corners-road cases and isolated fallback).
	soilIDs := [16]int{333, 365, 332, 238, 301, 333, 270, 333, 334, 237, 333, 333, 269, 333, 333, 333}
	for i, id := range soilIDs {
		soilAutotile[i] = terrainTile(id)
	}

	// Gravel autotile (terrain 18, roads). Same corner logic, different terrain.
	gravelIDs := [16]int{345, 377, 344, 250, 313, 345, 282, 345, 346, 249, 345, 345, 281, 345, 345, 345}
	for i, id := range gravelIDs {
		gravelAutotile[i] = terrainTile(id)
	}

	for dir := 0; dir < 4; dir++ {
		walkY := (lpcWalkBaseRow + dir) * lpcFrameSize
		thrustY := (lpcThrustBaseRow + dir) * lpcFrameSize
		for frame := 0; frame < 8; frame++ {
			x := frame * lpcFrameSize
			playerWalkFrames[dir][frame] = assets.PlayerSheet.SubImage(
				image.Rect(x, walkY, x+lpcFrameSize, walkY+lpcFrameSize)).(*ebiten.Image)
			playerThrustFrames[dir][frame] = assets.PlayerSheet.SubImage(
				image.Rect(x, thrustY, x+lpcFrameSize, thrustY+lpcFrameSize)).(*ebiten.Image)
		}
		dirY := lpcSlash128DirY[dir]
		for frame := 0; frame < lpcSlash128Frames; frame++ {
			x := frame * lpcSlash128FrameW
			playerSlash128Frames[dir][frame] = assets.PlayerSheet.SubImage(
				image.Rect(x, dirY, x+lpcSlash128FrameW, dirY+lpcSlash128FrameH)).(*ebiten.Image)
		}
	}
}

// Universal LPC spritesheet constants (64×64 px per frame).
// Row groups each have 4 rows: +0=up, +1=left, +2=down, +3=right.
const (
	lpcFrameSize     = 64
	lpcWalkBaseRow   = 8 // rows 8–11
	lpcThrustBaseRow = 4 // rows 4–7
)

// Slash128 section constants. Each 128px-tall block in the sheet contains two
// 64px sub-rows: the main character (top 64px) and an animal companion (bottom
// 64px). Only the top 64px is used here. Frames are 128px wide (axe arc extends
// beyond the character body). All 4 directions are available.
const (
	lpcSlash128FrameW = 128
	lpcSlash128FrameH = 128
	lpcSlash128Frames = 6
)

// lpcSlash128DirY maps direction index to the y-start of the main-character sub-row.
var lpcSlash128DirY = [4]int{3488, 3616, 3744, 3872} // up, left, down, right

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

// roadNeighborMask returns a 4-bit mask indicating which cardinal neighbors of
// (x, y) have a road level >= level. Bit 0=N, bit 1=E, bit 2=S, bit 3=W.
func roadNeighborMask(world *game.World, x, y, level int) int {
	dirs := [4][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}} // N, E, S, W
	mask := 0
	for i, d := range dirs {
		t := world.TileAt(x+d[0], y+d[1])
		if t != nil && game.RoadLevelFor(t) >= level {
			mask |= 1 << i
		}
	}
	return mask
}

// spriteForTile returns the base terrain sprite and any overlay sprites for a
// world tile. The base is drawn in pass 1 (all tiles); overlays are drawn in
// pass 2 (after all bases) so overflowing sprites are never masked by a
// neighbouring tile's ground layer.
func spriteForTile(tile *game.Tile, world *game.World, x, y int) (base drawArgs, overlays []drawArgs) {
	switch tile.Structure {
	case structures.FoundationLogStorage, structures.FoundationHouse:
		return drawArgs{img: dirtFoundationImg, scale: 1.0}, nil
	case structures.LogStorage:
		return drawArgs{img: barrelLogStorageImg, scale: 0.5}, nil
	case structures.House:
		return drawArgs{img: houseImg, scale: 1.0 / 3.0}, nil
	}

	switch tile.Terrain {
	case game.Forest:
		base = drawArgs{img: grassTileImg, scale: 1.0}
		switch {
		case tile.TreeSize == 0:
			// Stump: trunk sprite over grass.
			return base, []drawArgs{{img: lpcTreesTrunkImg, scale: 1.0 / 3.0}}
		case tile.TreeSize >= 7:
			// Mature: large tree scaled to 64px, offset up so the canopy overhangs
			// the tile above (rendered on top because pass 2 draws after all bases).
			return base, []drawArgs{{img: lpcTreesMatureImg, scale: 1.0 / 3.0, offsetY: -float64(tileSize)}}
		case tile.TreeSize >= 4:
			// Young: medium tree.
			return base, []drawArgs{{img: lpcTreesYoungImg, scale: 0.4}}
		default:
			// Sapling: small bush sprite.
			return base, []drawArgs{{img: lpcTreesSaplingImg, scale: 0.25}}
		}
	default:
		switch game.RoadLevelFor(tile) {
		case 2:
			return drawArgs{img: gravelAutotile[roadNeighborMask(world, x, y, 2)], scale: 1.0}, nil
		case 1:
			return drawArgs{img: soilAutotile[roadNeighborMask(world, x, y, 1)], scale: 1.0}, nil
		default:
			return drawArgs{img: grassTileImg, scale: 1.0}, nil
		}
	}
}

// Animation durations. Each animation runs to completion before a new cycle can start.
const (
	slashAnimDuration  = 750 * time.Millisecond  // 6 frames × ~125ms each
	thrustAnimDuration = 1000 * time.Millisecond // 8 frames × 125ms each
)

// playerAnimFrame selects the animation group and frame index for the player.
// slashCycleStart and thrustCycleStart are the wall-clock times when the current
// slash/thrust cycle began (zero = not active). Priority: slash > thrust > walk > idle.
// slash128 = true means the slash Slash128 section (128×128 frames) should be used;
// baseRow is unused in that case.
func playerAnimFrame(slashCycleStart, thrustCycleStart, now time.Time, moving bool, animTick int) (baseRow, frame int, slash128 bool) {
	if !slashCycleStart.IsZero() {
		if elapsed := now.Sub(slashCycleStart); elapsed >= 0 && elapsed < slashAnimDuration {
			// Slash128: 6 frames across slashAnimDuration.
			return 0, int(elapsed.Milliseconds() * lpcSlash128Frames / slashAnimDuration.Milliseconds()), true
		}
	}
	if !thrustCycleStart.IsZero() {
		if elapsed := now.Sub(thrustCycleStart); elapsed >= 0 && elapsed < thrustAnimDuration {
			// Thrust: 8 frames across thrustAnimDuration.
			return lpcThrustBaseRow, int(elapsed.Milliseconds() * 8 / thrustAnimDuration.Milliseconds()), false
		}
	}
	if moving {
		// Walk: 8 frames cycling at ~8fps (advance every 7 Update ticks at 60fps TPS).
		return lpcWalkBaseRow, (animTick / 7) % 8, false
	}
	return lpcWalkBaseRow, 0, false
}

// spriteForPlayer returns drawArgs for the player, selecting the correct frame
// from the Universal LPC spritesheet.
// When slash128=true, uses the Slash128 section (128×128 px frames) with an offset
// to center the larger sprite over the 32×32 tile. Otherwise, baseRow selects the
// animation group (lpcWalkBaseRow, lpcThrustBaseRow), dir the row within that group
// (0=up,1=left,2=down,3=right), and frame the column (0-based).
func spriteForPlayer(baseRow, dir, frame int, slash128 bool) drawArgs {
	if slash128 {
		return drawArgs{img: playerSlash128Frames[dir][frame], scale: 0.5, offsetX: -16, offsetY: 0}
	}
	if baseRow == lpcThrustBaseRow {
		return drawArgs{img: playerThrustFrames[dir][frame], scale: 0.5}
	}
	return drawArgs{img: playerWalkFrames[dir][frame], scale: 0.5}
}

// spriteForVillager returns drawArgs for a villager character.
func spriteForVillager() drawArgs {
	return drawArgs{img: villagerImg, scale: 0.5}
}
