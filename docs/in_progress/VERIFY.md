# Verification — House Building

## Primary gate
```bash
make check   # lint + test with race detector; must pass clean
```

## Per-stage verification

### Stage 1
```bash
make build   # must compile with new tile types and house.go
make check   # no lint errors; existing tests pass
```
Expected: `FoundationHouse`/`House` constants compile; `houseDef` satisfies `StructureDef`;
render glyphs update does not break `TestLogStorageWorkflow`.

### Stage 2
```bash
make check
```
Expected: `Build` cooldown type compiles; `BuildInterval`/`DepositInterval` on Player;
`log_storage.go` uses correct cooldown per path; existing E2E test still passes.

### Stage 3
```bash
make check
```
Expected: both upgrade IDs registered; 2-card screen renders without panic when offer
has 2 entries; "2" key selects second card.

### Stage 4
```bash
make check         # full gate
make e2e_viz       # visual sanity check (optional)
```

## Key test assertions (Stage 4)

- After ≥50 wood deposited into Log Storage → `HasStructureOfType(FoundationHouse)` is true.
- House foundation blocks movement into its 2×2 footprint.
- After 50 wood deposited into foundation → `HasStructureOfType(House)` is true;
  `HasStructureOfType(FoundationHouse)` is false.
- A tile inside the 2×2 house footprint has `Structure == House`.
- `HasPendingOffer()` is true after build; `CurrentOffer()` has 2 entries.
- Selecting card 0 reduces `Player.BuildInterval` by 10%.
- Selecting card 1 reduces `Player.DepositInterval` by 10%.

## Failure interpretation
- Compile error → wrong interface implementation or missing type constant.
- E2E test fails at "foundation not found" → `ShouldSpawn` condition wrong or
  storage not accumulating correctly.
- E2E test fails at "house not built" → `OnPlayerInteraction` cooldown logic wrong
  or `Build` cooldown not committed properly.
- Card count mismatch → `AddOffer` call in `OnBuilt` missing second ID.
