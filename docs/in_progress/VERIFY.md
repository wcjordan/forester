# Verification

## Primary gate
```bash
make check   # lint + test (must pass at end of every stage)
```

## Per-stage checks

### Stage 1 (geom package extracted)
```bash
make check
# Also verify package is standalone:
go build forester/game/geom
```
- `game/geom` must have zero imports from `game`
- `game.Point` and `geom.Point` are the same type (alias confirmed by compiler)

### Stage 2 (pathfinding moved)
```bash
make check
go build forester/game/geom
```
- `game/geom` must have zero imports from `game`
- All pathfinding tests pass (still in `game` package, call `geom.FindPath`)
- `World` satisfies `geom.Grid` implicitly (compiler verifies at `geom.FindPath(world, ...)` call site)

## Success criteria
- `make check` exits 0
- `game/geom` imports only stdlib
- No behavior changes (all existing tests pass, no tests deleted or skipped)
