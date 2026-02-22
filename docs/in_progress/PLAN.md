# Plan: Structure Spatial Index

## Goal
Replace the per-type scan in `TickAdjacentStructures` with a spatial index
(`map[Point]StructureEntry` on `World`) enabling O(1) adjacency lookup,
per-instance interaction, and extensibility to future entities (villagers).

## Non-goals
- Villager interactions (future; index is designed to support them)
- Structure removal (no removal mechanic exists yet)
- Diagonal adjacency

---

## Stage 1: Add `StructureIndex` to `World`

**Goal**: establish the data structure and population method.

Changes:
- Add `type Point struct { X, Y int }` in `world.go`
- Add `type StructureEntry struct { Def StructureDef; Origin Point }` in `structure.go`
- Add `StructureIndex map[Point]StructureEntry` field to `World`; initialise it in `NewWorld`
- Add `World.IndexStructure(x, y, w, h int, def StructureDef)` — stamps every tile in
  the footprint into the index with `Origin = {x, y}`

Tests (world_test.go):
- Single-tile: entry exists at correct coordinate with correct Origin
- Multi-tile 4×4: all 16 entries exist, all have the same Origin (top-left)
- Second call with same origin overwrites (idempotent)

Commit after green.

---

## Stage 2: Rename `OnAdjacentTick` → `OnPlayerInteraction(s *State, origin Point)`

**Goal**: update the interface and implementation; keep compiler as the gate.

Changes:
- Rename method in `StructureDef` interface (`structure.go`)
- Update `logStorageDef` in `log_storage.go` (origin param unused for now — `_`)
- Fix any remaining call sites (currently just `TickAdjacentStructures` in `state.go`)

No behaviour changes in this stage — `TickAdjacentStructures` is updated in Stage 3.

Commit after green.

---

## Stage 3: Wire index + replace `TickAdjacentStructures`

**Goal**: make the index the authoritative adjacency source; delete the old scan.

Changes:
- `State.AdvanceBuild`: after `World.SetStructure`, call
  `s.World.IndexStructure(x, y, w, h, def)` using the def returned by `findDefForBuilt`
- Replace `TickAdjacentStructures`:
  1. Check the 4 cardinal neighbours of player position in `World.StructureIndex`
  2. Collect unique `Origin` values (deduplicates multi-tile adjacency for one instance)
  3. Call `def.OnPlayerInteraction(s, origin)` once per unique origin
- Update `TestTickAdjacentStructures` helper: add `w.IndexStructure(5, 4, 4, 4, logStorageDef{})` after `SetStructure`
- Add test: player adjacent to two separate LogStorage instances calls interaction twice
- Remove `World.IsAdjacentToStructure` and `TestIsAdjacentToStructure` (replaced)

Commit after green.
