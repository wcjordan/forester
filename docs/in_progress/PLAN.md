# PLAN: Road Formation

## Goal
Implement organic road formation: Grassland tiles accumulate walk traffic from the player
and villagers. After enough steps, tiles upgrade to Trodden Path then Road. Roads give
faster movement and lower A* path cost, so villagers naturally reinforce existing roads.

## Constraints / Non-goals
- No decay (roads persist permanently once formed)
- 2 road levels only: Trodden (WalkCount >= 20) and Road (WalkCount >= 100)
- Forest tiles are not road-eligible (extensible via `isRoadEligible`)
- Villagers contribute WalkCount but don't get movement speed benefits (deferred)
- No new sprite assets — use programmatic solid-color images for Ebiten road tiles

## Parameters
| Constant             | Value  | Notes                         |
|----------------------|--------|-------------------------------|
| WalkCountTrodden     | 20     | Trodden path threshold        |
| WalkCountRoad        | 100    | Road threshold                |
| troddenMoveCooldown  | 120 ms | Player move cooldown          |
| roadMoveCooldown     | 90 ms  | Player move cooldown          |

## A* admissibility fix
`World.MoveCost` normalizes by `defaultMoveCooldown` (150 ms). Road tiles at 90 ms would
produce cost 0.6 < 1.0, breaking A* admissibility. Fix: normalize by `roadMoveCooldown`
(90 ms) so all terrain costs are >= 1.0:
- Road: 90/90 = 1.0
- Trodden: 120/90 = 1.33
- Grassland: 150/90 = 1.67
- Forest: 300/90 = 3.33

---

## Stage 1 — Core road logic + traffic counting
**Goal**: Data helpers and WalkCount increment on player/villager movement.

Steps:
1. Add constants `WalkCountTrodden`, `WalkCountRoad` to `game/player.go`
2. Add `isRoadEligible(tile *Tile) bool` (Grassland only) to `game/tile.go`
3. Add `RoadLevelFor(tile *Tile) int` (0/1/2) to `game/tile.go`
4. In `Player.Move()`: after `p.X, p.Y = nx, ny`, increment `destTile.WalkCount` if `isRoadEligible(destTile)`
5. In `Villager.move()`: after `v.X, v.Y = next.X, next.Y`, increment tile WalkCount if eligible
6. Add unit tests: `TestRoadLevelFor`, `TestPlayerMove_IncrementsWalkCount`, `TestVillagerMove_IncrementsWalkCount`
7. `make check` passes → commit

Exit criteria: WalkCount increments on grassland tiles when player/villager steps on them.
RoadLevelFor returns correct level for each threshold.

---

## Stage 2 — Movement speed + pathfinding cost
**Goal**: Road tiles grant faster player movement; A* prefers roads.

Steps:
1. Add `troddenMoveCooldown = 120ms`, `roadMoveCooldown = 90ms` constants in `game/player.go`
2. Update `MoveCooldownFor(tile *Tile)`: check `RoadLevelFor` before terrain lookup
3. Update `World.MoveCost`: normalize by `roadMoveCooldown` instead of `defaultMoveCooldown`
4. Update comment on `moveCooldowns` map to reflect new baseline
5. Add unit tests: `TestMoveCooldownFor_RoadLevels`, `TestMoveCost_RoadLevels`
6. `make check` passes → commit

Exit criteria: Player moves faster on trodden/road tiles. Pathfinding prefers roads (lower cost).

---

## Stage 3 — Rendering
**Goal**: TUI and Ebiten renderers show road level visually.

Steps:
1. TUI (`render/tui_model.go`):
   - Add `troddenStyle` (lipgloss color "130"/brown) and `roadStyle` (color "8"/gray)
   - In `View()`: after structure check, before terrain check, render `:` or `=` based on RoadLevelFor
2. Ebiten (`assets/assets.go`):
   - Add `TroddenPath *ebiten.Image` (solid tan: `0xC8,0xA0,0x60`) and `Road *ebiten.Image` (solid gray-brown: `0x90,0x78,0x60`)
   - Initialize both as 32x32 solid-color images in `init()`
3. Ebiten (`render/sprites.go`):
   - Add `troddenPathImg` and `roadImg` package-level vars
   - In `spriteForTile()`: before terrain switch, check `RoadLevelFor` and return road sprite
4. `make check` passes → commit

Exit criteria: Trodden tiles render as `:` (TUI) / tan (Ebiten). Road tiles render as `=` (TUI) / gray-brown (Ebiten).

---

## Stage 4 — Documentation + PR
**Goal**: Clean up, update docs, open PR.

Steps:
1. Update `docs/PROJECT_PLAN.md`: mark "Road formation" as implemented in Phase 2 notes
2. Update `docs/future_ideas/road_formation.md`: note 2-level impl complete
3. `make check` passes
4. Self-review all diffs
5. Commit docs
6. Push branch, open PR
