# STATUS

- [ ] Stage 1: Define `StructureDef` interface and registry in `structure.go`
- [ ] Stage 2: Implement `logStorageDef` in new `log_storage.go`
- [ ] Stage 3: Generalize `state.go` dispatch + update tests

## Current state
Planning complete. Ready to begin implementation.

## Key decisions
- Stay in `game` package; file naming convention for grouping (`log_storage.go`, `house.go`, etc.)
- `StructureType` int enum stays on `Tile` for lightweight tile storage
- `StructureDef` interface covers behavioral interactions (spawn condition + adjacent tick)
- Generic "already placed" guard lives in the registry loop in `state.go`, not in `ShouldSpawn`
- `TryDeposit()` replaced by `TickAdjacentStructures()` (no return value; callers check state)
