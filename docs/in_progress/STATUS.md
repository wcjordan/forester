# Status

## Stages
- [ ] Stage 1: Add `StructureIndex` to `World`
- [ ] Stage 2: Rename `OnAdjacentTick` → `OnPlayerInteraction(s *State, origin Point)`
- [ ] Stage 3: Wire index + replace `TickAdjacentStructures`

## Current
Planning complete. Starting Stage 1.

## Decisions
- `Point` lives in `world.go` (world coordinates)
- `StructureEntry` lives in `structure.go` (alongside `StructureDef`)
- Index on `World` (spatial data), populated by `State.AdvanceBuild` (game logic)
- Ghost structures are NOT indexed — only built structures
- Multi-tile dedup: collect unique `Origin` values from adjacent tiles; call once per instance
