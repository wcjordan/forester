package gui

import (
	"image"
	"image/color"
	"time"

	ebiten "github.com/hajimehoshi/ebiten/v2"

	"forester/assets"
	"forester/game"
	"forester/game/structures"
	"forester/render/roads"
	"forester/render/spritedata"
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
	dirtFoundationImg = assets.Dirt.SubImage(image.Rect(32, 64, 32+32, 64+32)).(*ebiten.Image)
	grassTileImg      = assets.GrassTile.SubImage(image.Rect(0, 0, 0+32, 0+32)).(*ebiten.Image)

	// logStorageBuildingImg is the pre-composed 128×128 log-storage yard image assembled in init.
	logStorageBuildingImg *ebiten.Image

	// Road autotile arrays indexed by 4-bit neighbor bitmask (bit0=N, bit1=E, bit2=S, bit3=W).
	// Tiles are composed from four 16×16 quadrants; mappings are in roads.SoilComposed /
	// roads.GravelComposed.
	soilAutotile   [16]*ebiten.Image // trodden path (level 1)
	gravelAutotile [16]*ebiten.Image // road (level 2)

	// lpc-trees: sapling (128×96), young (128×128), mature (160×192), trunk (80x50)
	lpcTreesSaplingImg = assets.TreesGreen.SubImage(spritedata.SaplingRect).(*ebiten.Image)
	lpcTreesYoungImg   = assets.TreesGreen.SubImage(spritedata.YoungRect).(*ebiten.Image)
	lpcTreesMatureImg  = assets.TreesGreen.SubImage(spritedata.MatureRect).(*ebiten.Image)
	lpcTreesTrunkImg   = assets.TreesGreen.SubImage(spritedata.TrunkRect).(*ebiten.Image)

	// houseBuildingImg is the pre-composed 64×96 house sprite assembled in init.
	// Layout: y=0..64 thatched roof (overflows one row above footprint), y=64..96 wall face.
	houseBuildingImg *ebiten.Image

	// resourceDepotPlaceholderImg is a solid amber rectangle spanning the 5×4 footprint (160×128).
	resourceDepotPlaceholderImg *ebiten.Image

	// Characters
	villagerImg = assets.Villager.SubImage(image.Rect(0, 128, 0+64, 128+64)).(*ebiten.Image)

	// Player animation frames: [direction][frame], direction 0=up 1=left 2=down 3=right.
	playerWalkFrames     [4][8]*ebiten.Image
	playerThrustFrames   [4][8]*ebiten.Image
	playerSlash128Frames [4][lpcSlash128Frames]*ebiten.Image
)

func init() {
	for i, c := range roads.SoilComposed {
		soilAutotile[i] = roads.ComposeFromSheet(assets.TerrainSheet, c)
	}
	for i, c := range roads.GravelComposed {
		gravelAutotile[i] = roads.ComposeFromSheet(assets.TerrainSheet, c)
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

	initHouseBuilding()
	initLogStorageBuilding()
	initResourceDepotPlaceholder()
}

// initLogStorageBuilding composes the 128×128 log-storage yard image via spritedata.BuildLogStorageImg.
func initLogStorageBuilding() {
	logStorageBuildingImg = spritedata.BuildLogStorageImg(assets.ContainerSheet)
}

// initHouseBuilding composes the 64×96 house building image via spritedata.BuildHouseImg.
func initHouseBuilding() {
	houseBuildingImg = spritedata.BuildHouseImg(assets.ThatchedRoofSheet, assets.CottageSheet, assets.WindowsDoorsSheet)
}

// initResourceDepotPlaceholder creates a solid amber 160×128 rectangle (5×4 tiles at 32px each).
func initResourceDepotPlaceholder() {
	w, h := 5*tileSize, 4*tileSize
	img := ebiten.NewImage(w, h)
	img.Fill(color.RGBA{R: 0xD4, G: 0x7F, B: 0x00, A: 0xFF}) // amber/gold
	resourceDepotPlaceholderImg = img
}

// tileSize is the side length in pixels of one world tile at zoom=1.
const tileSize = 32

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

// logStorageOverlays returns the overlay drawArgs for the NW-anchor log-storage tile.
// The 128×128 yard image fits exactly within the 4×4 footprint — no offset needed.
func logStorageOverlays() []drawArgs {
	return []drawArgs{{img: logStorageBuildingImg, scale: 1.0}}
}

// houseOverlays returns the overlay drawArgs for the NW-anchor house tile.
// The 64×96 building sprite is drawn with offsetY=-tileSize so it overflows
// one row above the 2×2 footprint.
func houseOverlays() []drawArgs {
	return []drawArgs{{img: houseBuildingImg, scale: 1.0, offsetY: -tileSize}}
}

// roadNeighborMask returns a 4-bit mask indicating which cardinal neighbors of
// (x, y) have any road (any level) or a building. Bit 0=N, bit 1=E, bit 2=S, bit 3=W.
// Roads of all levels are treated as connected to each other.
func roadNeighborMask(world *game.World, x, y int) int {
	dirs := [4][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}} // N, E, S, W
	mask := 0
	for i, d := range dirs {
		t := world.TileAt(x+d[0], y+d[1])
		if t != nil && (game.RoadLevelFor(t) > 0 || t.Structure != game.NoStructure) {
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
	case structures.FoundationLogStorage, structures.FoundationHouse, structures.FoundationResourceDepot:
		return drawArgs{img: dirtFoundationImg, scale: 1.0}, nil
	case structures.LogStorage:
		// Only draw the logStorage sprite from the origin tile
		if !world.IsStructureOrigin(x, y) {
			return drawArgs{img: grassTileImg, scale: 1.0}, nil
		}
		return drawArgs{img: grassTileImg, scale: 1.0}, logStorageOverlays()
	case structures.House:
		// Only draw the house sprite from the origin tile
		if !world.IsStructureOrigin(x, y) {
			return drawArgs{img: grassTileImg, scale: 1.0}, nil
		}
		return drawArgs{img: grassTileImg, scale: 1.0}, houseOverlays()
	case structures.ResourceDepot:
		// Only draw the depot placeholder from the origin tile
		if !world.IsStructureOrigin(x, y) {
			return drawArgs{img: grassTileImg, scale: 1.0}, nil
		}
		return drawArgs{img: grassTileImg, scale: 1.0}, []drawArgs{{img: resourceDepotPlaceholderImg, scale: 1.0}}
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
			return drawArgs{img: grassTileImg, scale: 1.0}, []drawArgs{{img: gravelAutotile[roadNeighborMask(world, x, y)], scale: 1.0}}
		case 1:
			return drawArgs{img: grassTileImg, scale: 1.0}, []drawArgs{{img: soilAutotile[roadNeighborMask(world, x, y)], scale: 1.0}}
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
		return drawArgs{img: playerSlash128Frames[dir][frame], scale: 0.5, offsetX: -32, offsetY: -32}
	}
	if baseRow == lpcThrustBaseRow {
		return drawArgs{img: playerThrustFrames[dir][frame], scale: 0.5, offsetX: -16, offsetY: -16}
	}
	return drawArgs{img: playerWalkFrames[dir][frame], scale: 0.5, offsetX: -16, offsetY: -16}
}

// spriteForVillager returns drawArgs for a villager character.
func spriteForVillager() drawArgs {
	return drawArgs{img: villagerImg, scale: 0.5}
}
