# Better Sprites Plan

Four targeted visual improvements to the Ebitengine renderer. All changes are render-layer only — no game logic (`game/`) changes required.

---

## Problems to Solve

| # | Problem | Impact |
|---|---|---|
| S1 | Player and villager show a static sprite regardless of movement direction | Avatars feel frozen; no sense of life |
| S2 | Trees are visually inconsistent — small trees show a plain grass tile; different sizes pull from unrelated sprite sources | Forest reads as noise rather than density |
| S3 | Roads and trodden paths are solid colored rectangles; no visual connection between adjacent tiles | Roads look painted on; no sense of paths forming |
| S4 | House (2×2) and log storage (4×4) repeat one small sprite across every tile in their footprint | Buildings look like a grid of stamps, not structures |

---

## Asset Sources

Download/generate these before implementing each stage:

| Asset | Source | Stage |
|---|---|---|
| Player walkcycle spritesheet (composed) | [LPC Universal Character Generator](https://sanderfrenken.github.io/Universal-LPC-Spritesheet-Character-Generator/) — pick soldier armor layers, export single PNG | S1 |
| Villager walkcycle spritesheet (composed) | Same tool, different color scheme | S1 |
| Tree sprites (sapling/mid/mature progression) | [lpc-trees](https://opengameart.org/content/lpc-trees) | S2 |
| Path/road autotile set | [lpc-terrains](https://opengameart.org/content/lpc-terrains) | S3 |
| Thatched cottage tiles | [lpc-thatched-roof-cottage](https://opengameart.org/content/lpc-thatched-roof-cottage) | S4 |
| Adobe building tiles (fallback if cottage too small) | [lpc-adobe-building-set](https://opengameart.org/content/lpc-adobe-building-set) | S4 |
| Exterior containers (barrels, crates) | [lpc-containers](https://opengameart.org/content/lpc-containers) | S4 |

All sources are CC-BY or CC0 licensed. Add attribution to `assets/sprites/CREDITS.txt` when adding each pack.

---

## Stage S1 — Walk Animation (Player & Villager)

**Goal:** Player and villager animate while moving; idle pose when standing still.

### LPC walkcycle layout

The LPC Character Generator exports a standard walkcycle sheet: 4 rows × 9 columns, 64×64 px per frame.

```
Row 0 (y=0)   — walk up
Row 1 (y=64)  — walk left
Row 2 (y=128) — walk down   ← currently used as the static player sprite
Row 3 (y=192) — walk right
```

Frames 0 and 4 are neutral (feet together); frames 1–3 and 5–7 are walk steps. Cycle frames 0–7 while moving at ~8 FPS; hold frame 0 of the facing row when idle.

### Code changes (`render/` only)

- Add `playerDir`, `villagerDirs []int`, `animFrame int`, `animTimer float64` to `EbitenGame`.
- Derive `playerDir` from the last pressed movement key (W=up, S=down, A=left, D=right) in `Update()`.
- Advance `animFrame` by elapsed time while a movement key is held; reset to 0 when no key is held.
- Update `spriteForPlayer()` / `spriteForVillager()` to accept direction + frame and return the correct 64×64 `SubImage` crop.
- For villagers: derive direction from position delta between ticks (store previous position in `EbitenGame`).

### Exit criteria

- Player animates through walk frames in the correct direction while WASD is held.
- Player shows idle (frame 0, facing direction) when no key is held.
- Villagers animate while moving; idle when standing.
- `make check` passes.

---

## Stage S2 — Tree Visual Cohesion

**Goal:** All forest tiles share a consistent grass base; tree canopy and trunk layer on top; density is legible from sapling to mature.

### Current problems

- `TreeSize 1–3` (sapling): renders the forest grass tile — looks like plain grass, no tree visible.
- `TreeSize 4–6` (young) and `≥7` (mature): use unrelated sprite crops; no visual continuity.
- Stump (`TreeSize 0`): trunk crop is reasonable but sits on no grass base.

### Rendering approach

Draw in layers per forest tile:

1. **Always:** grass base tile (same tile used for grassland, so forest blends naturally).
2. **Sapling (1–3):** small canopy sprite from lpc-trees, centered on tile.
3. **Young (4–6):** medium canopy sprite.
4. **Mature (7–10):** large canopy sprite + trunk sprite drawn beneath (trunk at tile bottom, canopy overlapping tile above).
5. **Stump (0):** stump/trunk sprite only.

Multi-tile trees (mature): the trunk occupies the tile, the canopy overhangs into the tile above. Draw canopy with a negative Y offset to simulate height. This requires drawing canopy after all terrain in the row above is drawn — use a second pass or draw order adjustment.

### Code changes

- Update `spriteForTile()` to return a slice of `drawArgs` (layers) instead of a single `drawArgs`.
- Add lpc-trees frames to `assets/assets.go` and pre-slice in `render/sprites.go`.
- Update the draw loop in `Draw()` to iterate layers.

### Exit criteria

- All forest tiles have a visible grass base.
- Sapling/young/mature form a clear visual size progression using sprites from the same source.
- Stump reads as a cut tree.
- No visual seam between forest and grassland terrain.
- `make check` passes.

---

## Stage S3 — Road Autotiling

**Goal:** Road and trodden-path tiles connect visually to neighbors; tile drawn reflects which directions are connected.

### Autotile logic

For each road-level tile, check its 4 cardinal neighbors (N, S, E, W) to see if they are at the same road level or higher. Build a 4-bit bitmask:

```
bit 0 = N neighbor is road
bit 1 = E neighbor is road
bit 2 = S neighbor is road
bit 3 = W neighbor is road
```

Map the 16 bitmask values to specific tiles from lpc-terrains (straight H, straight V, corner ×4, T-junction ×4, cross, and dead-end ×4). Do this separately for trodden (level 1) and road (level 2).

For transitions between road levels (e.g., road entering a trodden stretch), treat the lower level as "no connection" from the higher level's perspective.

### Code changes

- Add `roadNeighborMask(world, x, y, level int) int` helper in `render/sprites.go`.
- Replace the current `roadImg` / `troddenPathImg` single-sprite lookup with a 16-entry tile map per road level.
- Load lpc-terrains tiles into `assets/assets.go`; pre-slice all variants.

### Exit criteria

- Straight roads (H and V) connect end-to-end.
- Corners, T-junctions, and crossings display the correct tile.
- Trodden paths autotile independently of roads.
- Edge tiles (only one neighbor) show a dead-end or tapered tile, not a through-road.
- `make check` passes.

---

## Stage S4 — Building Footprint Rendering

**Goal:** House and log storage each render as a single coherent multi-tile image rather than a repeated per-tile sprite.

### House (2×2 footprint)

Assemble from lpc-thatched-roof-cottage tile pieces (or the existing `house.png` tileset as fallback):
- Draw terrain (grass) for all 4 tiles first.
- Draw the composed roof/wall assembly once, anchored at the NW tile of the footprint.
- If the cottage tiles are smaller than 64×64 (2 tiles), scale up or use the adobe building set as an alternative.

### Log Storage (4×4 footprint)

Render as a storage shed with containers arranged around it:
- Draw terrain for all 16 tiles.
- Draw a small shed building (from the house/cottage tileset, different from the player house) anchored at the NW corner, occupying roughly the center 2×2 of the footprint.
- Fill the surrounding tiles with stacked containers (barrels, crates) from lpc-containers.

### Anchor detection

In the draw loop, a structure tile should only draw the building sprite if it is the **NW anchor** of its footprint. Detect this by checking that neither `(x-1, y)` nor `(x, y-1)` holds the same `StructureType`. All other footprint tiles draw only terrain.

### Foundation tiles

Foundation tiles (`?`) are single-tile and can keep the current dirt sprite — no change needed.

### Code changes

- Update `spriteForTile()` to return `nil` (skip building draw) for non-anchor structure tiles.
- Add a separate `spriteForBuilding(structType, anchorX, anchorY)` that returns the full multi-tile composed draw call.
- Add new building + container sprites to `assets/assets.go`.

### Exit criteria

- House renders as a single coherent 2×2 building with roof visible.
- Log storage renders as a shed surrounded by stacked containers in the 4×4 area.
- No repeated tile artifacts.
- Foundation tiles unchanged.
- `make check` passes.

---

## Implementation Order

The stages are independent. Recommended order by player-facing impact:

1. **S1** — animation (pure code + LPC generator, no new downloads)
2. **S2** — trees (mostly code, lpc-trees download)
3. **S3** — roads (lpc-terrains download + autotiling code)
4. **S4** — buildings (multiple asset downloads + draw loop refactor)

Each stage should be committed separately with `make check` passing before moving to the next.
