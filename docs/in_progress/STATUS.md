# Status

## Current
- Planning complete, awaiting approval

## Stages
- [ ] Stage 1: Create `game/geom` with geometry + `Point`
- [ ] Stage 2: Move pathfinding to `game/geom` with `Grid` interface
- [ ] Stage 3: Cleanup

## Key decisions
- `type Point = geom.Point` alias in `game` — external callers unchanged
- `geom` and pathfinding share one package
- Unexported in `game/geom`: `abs`, `chebyshevRingDo`, `manhattan`
- Exported (renamed): `SpiralSearchDo`, `ForFootprintCardinalNeighbors`, `FindPath`, `FootprintBorderDo` (unchanged)
- Pathfinding tests stay in `game` package as integration tests
