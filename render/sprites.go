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

	// Player animation frames: [direction][frame], direction 0=up 1=left 2=down 3=right.
	playerWalkFrames     [4][8]*ebiten.Image
	playerThrustFrames   [4][8]*ebiten.Image
	playerSlash128Frames [4][lpcSlash128Frames]*ebiten.Image
)

func init() {
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

// spriteForTile returns the drawArgs for a world tile (terrain + structure).
func spriteForTile(tile *game.Tile) drawArgs {
	switch tile.Structure {
	case structures.FoundationLogStorage, structures.FoundationHouse:
		return drawArgs{img: dirtFoundationImg, scale: 1.0}
	case structures.LogStorage:
		return drawArgs{img: barrelLogStorageImg, scale: 0.5}
	case structures.House:
		return drawArgs{img: houseImg, scale: 1.0 / 3.0}
	}

	switch tile.Terrain {
	case game.Forest:
		switch {
		case tile.TreeSize == 0:
			return drawArgs{img: trunkSmallImg, scale: 1.0 / 3.0}
		case tile.TreeSize >= 7:
			return drawArgs{img: treetopMatureImg, scale: 1.0 / 3.0}
		case tile.TreeSize >= 4:
			return drawArgs{img: treetopYoungImg, scale: 1.0 / 3.0}
		default:
			return drawArgs{img: grassForestImg, scale: 1.0}
		}
	default:
		switch game.RoadLevelFor(tile) {
		case 2:
			return drawArgs{img: roadImg, scale: 1.0}
		case 1:
			return drawArgs{img: troddenPathImg, scale: 1.0}
		default:
			return drawArgs{img: grassTileImg, scale: 1.0}
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
