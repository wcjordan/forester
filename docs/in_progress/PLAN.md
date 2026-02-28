# Plan: Extract `game/geom` subpackage

## Goal
Move pure geometry functions and pathfinding out of `game` into a `game/geom` subpackage.
`game` keeps a `type Point = geom.Point` alias so external callers (`render/`, `game/structures/`, etc.) need no changes.

## Non-goals
- No changes to `render/`, `game/resources/`, `game/structures/` beyond the one call-site fix in `house.go`
- No behavior changes — pure mechanical extraction

---

## Stage 1: Create `game/geom` with pure geometry + `Point`

**Goal:** New `game/geom` package compiles and passes tests; `game` imports it.

**Steps:**
1. Create `game/geom/geom.go`:
   - `package geom`
   - `type Point struct{ X, Y int }` (moved from `game/world.go`)
   - Copy all functions from `game/geom.go`
   - Export the ones used outside `game`: `spiralSearchDo` → `SpiralSearchDo`, `forFootprintCardinalNeighbors` → `ForFootprintCardinalNeighbors`
   - `abs`, `chebyshevRingDo`, `manhattan` stay unexported (internal use only)
2. Create `game/geom/geom_test.go`:
   - Move from `game/geom_test.go`, set `package geom`
   - Update function calls for renamed exports
3. Update `game/world.go`:
   - Add `import "forester/game/geom"`
   - Add `type Point = geom.Point`
   - Remove `type Point struct{ X, Y int }` definition
4. Update `game/villager.go`: import geom; `spiralSearchDo` → `geom.SpiralSearchDo`; `forFootprintCardinalNeighbors` → `geom.ForFootprintCardinalNeighbors`
5. Update `game/spawn.go`: import geom; `spiralSearchDo` → `geom.SpiralSearchDo`
6. Update `game/structures/house.go`: import geom; `game.FootprintBorderDo` → `geom.FootprintBorderDo`
7. Delete `game/geom.go` and `game/geom_test.go`
8. `make check` → commit

**Exit criteria:** `make check` passes; `game/geom` is a standalone compilable package.

---

## Stage 2: Move pathfinding to `game/geom` with `Grid` interface

**Goal:** A* lives in `game/geom`; `World` satisfies a `Grid` interface; no `game` types leak into `game/geom`.

**Steps:**
1. Create `game/geom/pathfinding.go`:
   - Define `Grid` interface: `InBounds(x, y int) bool`, `IsBlocked(x, y int) bool`, `MoveCost(x, y int) int`
   - `FindPath(g Grid, fromX, fromY, toX, toY int) []Point`
   - Move `reconstructPath` and priority queue from `game/pathfinding.go`
   - Replace `w.TileAt` / `tile.Structure` / `tileCost(tile)` with `g.IsBlocked` / `g.MoveCost`
2. Add to `game/world.go` (or thin `game/grid.go`):
   - `func (w *World) IsBlocked(x, y int) bool` — `tile == nil || tile.Structure != NoStructure`
   - `func (w *World) MoveCost(x, y int) int` — absorbs `tileCost` logic: Forest+TreeSize>0 → 2, else 1
3. Update `game/villager.go`: `findPath(world, ...)` → `geom.FindPath(world, ...)`
4. Update `game/pathfinding_test.go`: add `import "forester/game/geom"`; `findPath(w, ...)` → `geom.FindPath(w, ...)`
5. Delete `game/pathfinding.go`
6. `make check` → commit

**Exit criteria:** `make check` passes; `game/geom` has zero imports from `game`; `World` satisfies `geom.Grid` implicitly.

---

## Stage 3: Cleanup

1. Delete `docs/in_progress/` files
2. Final `make check`
3. Commit
