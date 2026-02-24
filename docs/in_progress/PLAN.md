# Plan: Circular Spawn Clearing + Tree No-Grow Zones

## Goal
1. Replace 5×5 square spawn clearing with a circle of Euclidean radius 5.
2. Suppress tree regrowth within Euclidean distance 8 of the spawn point.
3. Suppress tree regrowth within Euclidean distance 8 of any structure (all entries in StructureIndex).

## Non-goals
- No change to initial tree placement (Steps 1–2.5 in GenerateWorld).
- No change to `Regrow` signature — spawn is always `width/2, height/2`.
- No performance optimization beyond what's needed for the game's scale.

## Distance math
Use squared-distance comparison to avoid floating point:
`dx*dx + dy*dy <= r*r` (inline at each use site — no helper needed).

---

## Stage 1: Circular spawn clearing (`worldgen.go` + `worldgen_test.go`)

**Goal:** Replace the `dy in [-2,2], dx in [-2,2]` loop with a loop that covers
a bounding box of radius 5 and clears only tiles where `dx*dx+dy*dy <= 5*5`.

**Steps:**
1. In `worldgen.go` Step 3, replace the 5×5 loop with a radius-5 circular loop.
2. In `worldgen_test.go` `TestGenerateWorld_SpawnClear`, update the loop to check
   the same circle (radius 5) and verify tiles inside are Grassland.
   Also verify at least one corner tile at dx=5,dy=0 is clear, and that the
   test makes sense geometrically.
3. Run `make check`.
4. Commit: `Worldgen: circular spawn clearing with Euclidean radius 5`

**Exit criteria:**
- All tiles within Euclidean radius 5 of center are Grassland after GenerateWorld.
- `make check` passes.

---

## Stage 2: No-grow zones in `Regrow` (`world.go` + `world_test.go`)

**Goal:** Modify `Regrow` to skip eligible Forest tiles that are within distance 8
of spawn or within distance 8 of any structure in StructureIndex.

**Steps:**
1. In `world.go` `Regrow`, compute `cx, cy = w.Width/2, w.Height/2` at the top.
2. For each eligible Forest tile `(x, y)`, before rolling:
   - If `(x-cx)*(x-cx)+(y-cy)*(y-cy) <= 8*8` → skip.
   - If any `StructureIndex` point is within `dx*dx+dy*dy <= 8*8` → skip.
3. Update existing `TestRegrow` subtests that use a 3×3 world (spawn = center),
   which would be inside the no-grow zone — move them to a 20×20 world with the
   Forest tile at `(0, 0)` (≈14 tiles from center, safely outside the zone).
4. Add new subtests:
   - `forest within spawn no-grow zone does not grow` — Forest tile at spawn center,
     run 1000 Regrow iterations, assert TreeSize stays 0.
   - `forest within building no-grow zone does not grow` — Forest tile adjacent to
     a structure, run 1000 iterations, assert TreeSize stays 0.
5. Run `make check`.
6. Commit: `World: suppress tree regrowth near spawn and structures`

**Exit criteria:**
- Forest tiles ≤8 tiles (Euclidean) from spawn never grow in Regrow.
- Forest tiles ≤8 tiles from any structure tile never grow in Regrow.
- Tiles outside both zones still grow.
- `make check` passes.
