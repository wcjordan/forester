# S4 — House Building Footprint Rendering

## Goal

Replace the current per-tile house sprite with a single coherent 2×2 building anchored
at its NW tile:

- **Thatched roof**: 64×64 crop from `thatched-roof.png`, drawn with a –32px Y overflow
  so the roof peak extends above the north footprint row (same pattern as mature trees).
- **Half-timber wall face**: 64×32 crop from `cottage.png`, visible on the south row.
- **Door + windows**: centered wooden door and flanking flower-box windows overlaid on the
  wall face, sourced from `windows-doors.png`.

Target appearance: reference image supplied by user (thatched cottage front facade).

## Non-goals

- Log storage rendering (separate task).
- Foundation tiles (`FoundationHouse`) — keep existing per-tile dirt rendering.
- Villager animation (separate task).

---

## Stages

### Stage 1 — Embed spritesheets & pre-slice frames

**Goal**: Load the three new sheets and expose pre-sliced frame vars. No rendering change.

**Steps**:
1. Add `//go:embed` directives and image vars in `assets/assets.go`:
   - `//go:embed sprites/lpc-thatched-roof-cottage` → `cottageFS`
   - `//go:embed sprites/lpc-windows-doors-v2` → `windowsDoorsFS`
   - Load `ThatchedRoofSheet`, `CottageSheet`, `WindowsDoorsSheet *ebiten.Image` in `init()`
2. In `render/sprites.go`, manually inspect each sheet and add pre-sliced vars:
   - `houseThatcedRoofImg` — roof section from `thatched-roof.png` (yellow/wheat variant)
   - `houseWallImg` — half-timber wall panel from `cottage.png`
   - `houseDoorImg` — small wooden door from `windows-doors.png`
   - `houseWindowImg` — wood-framed window with flower box from `windows-doors.png`
3. Commit with `make check` passing.

**Exit criteria**: `make check` passes; no panic on startup; sprites load correctly.

---

### Stage 2 — NW-anchor detection + roof overlay

**Goal**: House tiles render as grass; NW anchor tile renders a roof overlay (placeholder
to validate anchor logic and overflow rendering).

**Steps**:
1. Add `isStructureNWAnchor(world *game.World, x, y int, st game.StructureType) bool`
   in `render/sprites.go`:
   - Returns true iff neither `(x-1, y)` nor `(x, y-1)` holds the same StructureType.
2. Update `spriteForTile()` house case:
   - All house tiles: `return drawArgs{img: grassTileImg, scale: 1.0}, nil`
   - NW anchor only: return grass base + `[]drawArgs{houseThatcedRoofImg overlay}`
     with `offsetY: -tileSize` to verify overflow above north row.
3. Commit with `make check` passing.

**Exit criteria**: 3 non-anchor house tiles show plain grass; NW tile shows roof
sprite overflowing above the north row; foundation tiles unchanged.

---

### Stage 3 — Full composition (wall + door + windows)

**Goal**: Pre-compose a single 64×96 building image at init time and use it as the
house overlay, matching the reference image.

**Steps**:
1. In `render/sprites.go` `init()`, create `houseBuildingImg *ebiten.Image` (64×96):
   - Draw `houseThatcedRoofImg` scaled to 64×64 at (0, 0) — roof fills top 64px.
   - Draw `houseWallImg` scaled to 64×32 at (0, 64) — wall fills bottom 32px.
   - Draw `houseDoorImg` centered horizontally at (0, 64) — door on wall face.
   - Draw `houseWindowImg` at left and right of door on wall face.
2. Replace Stage 2 placeholder overlay with:
   `drawArgs{img: houseBuildingImg, scale: 1.0, offsetY: -float64(tileSize)}`
3. Tune crop coordinates and draw offsets to match reference.
4. Commit with `make check` passing.

**Exit criteria**:
- House renders as a coherent thatched cottage (roof + wall + door + windows).
- Visual matches the reference image.
- No repeated-tile artifacts on the other 3 footprint tiles.
- Foundation tiles (`FoundationHouse`) still render as per-tile dirt.
- `make check` passes.

---

## Key files

| File | Change |
|---|---|
| `assets/assets.go` | Add embeds + image vars for 2 new sheets |
| `render/sprites.go` | Pre-slice frames; add `isStructureNWAnchor`; update `spriteForTile` house case; compose `houseBuildingImg` |
| `render/ebiten_model.go` | No change expected |
| `game/` | No change |
