# S3 — Road Autotiling

## Goal
Replace flat-color road/trodden-path tiles with textured autotile sprites from lpc-terrains.
Each road tile samples the correct corner-blend tile based on which of its 4 cardinal neighbors
are also road, producing smooth road-to-grass transitions at path edges.

## Asset approach
lpc-terrains uses a corner-based terrain autotile (not road-shape tiles).
- Trodden path → Soil (terrain 14, center tile 333)
- Road → Gravel_1 (terrain 18, center tile 345)
The 4-bit N/E/S/W bitmask maps to 4 tile corners (TL/TR/BL/BR):
  TL = road if N or W; TR = road if N or E; BL = road if S or W; BR = road if S or E.
The result: full-tile textured road with grass blending at unconnected edges.

## Non-goals
- Narrow path shapes (lpc-terrains doesn't have them; full-tile roads are correct for this game)
- Diagonal neighbor checks
- Villager animation (separate future task)

## Steps

### Stage 1 — Assets
- Add `lpcTerrainsFS` embed + `TerrainSheet *ebiten.Image` to `assets/assets.go`
- Remove now-unused `TroddenPath` and `Road` solid-color vars from `assets/assets.go`
- Fix README.md attribution filename: CREDITS.txt → CREDITS-terrain.txt
- Exit: compiles, `make check` passes

### Stage 2 — Autotile data + helper
- Add `terrainTile()` helper in `render/sprites.go` (tile ID → SubImage)
- Add `soilAutotile [16]*ebiten.Image` and `gravelAutotile [16]*ebiten.Image` vars, init in `init()`
- Add `roadNeighborMask(world *game.World, x, y, level int) int` helper
- Remove old `troddenPathImg` / `roadImg` package vars
- Exit: compiles (will break until callers updated)

### Stage 3 — Wire up in renderer
- Update `spriteForTile(tile, world, x, y)` signature; update road case to use autotile
- Update both `spriteForTile` call sites in `render/ebiten_model.go`
- Exit: `make check` passes, roads/paths show lpc-terrains textures

### Stage 4 — Commit + PR
- Commit with `make check` passing
- Open PR
