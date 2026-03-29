# Verification — Resource Depot Phase 1

## Primary gate

```bash
make check   # lint + test (must pass at each stage commit)
```

## Stage-by-stage checks

### Stage 1 (game logic)

```bash
make test
```

Expected:
- All existing tests still pass.
- New unit tests in `game/structures/` (if added) pass.
- `FoundationResourceDepot` spawns when 4 houses exist and no depot is pending.
- Depot built → `large_carry_capacity` offer queued.
- `largeCarryCapacityUpgrade.Apply` increases `MaxCarry` by 100.

### Stage 2 (rendering)

```bash
make check
make build   # binary compiles without errors
```

Expected:
- TUI: `D` visible in bold cyan for built depot; `?` for foundation (same as other foundations).
- Ebitengine: no panic; depot NW tile renders a colored rectangle.

### Stage 3 (E2E test)

```bash
make test
```

Expected:
- `TestResourceDepotWorkflow` passes.
- Foundation spawns after 4th house.
- Depot builds on 800 wood deposits.
- Card offer fires; MaxCarry increases.

## Failure signals

- `make check` failures → fix before committing; do not `--no-verify`.
- E2E test panics / infinite loops → check story beat ordering and placement validity.
- Ebitengine sprite crash → ensure non-anchor tiles skip building draw.
