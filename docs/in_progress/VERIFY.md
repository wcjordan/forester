# Verify

## Primary gate (run after each stage)
```bash
make check   # lint + test with race detector
```

## Per-stage assertions

### Stage 1
- `make check` passes.
- `game/env.go` exists with `Env` struct.
- `game/storage_manager.go` exists with `StorageManager`, `StorageState`.
- `logStorageDef` implements `StorageDef` (compiler enforces via interface).
- No existing files modified (diff only shows new files + `storage.go` + `log_storage.go`).

### Stage 2
- `make check` passes.
- `State` has no `Storage`, `StorageByOrigin`, `getStorage`, `TotalStored` fields/methods.
- `Game` has `Stores *StorageManager` field.
- `StructureDef` interface methods all take `*Env`.
- `e2e_tests/log_storage_test.go` uses `g.Stores.Total(game.Wood)`.

### Stage 3
- `make check` passes.
- `TestStorageManagerRoundTrip` exists and passes.
- Round-trip: `manager2.Total(Wood) == manager.Total(Wood)` after `LoadFrom(SaveData())`.
- Round-trip: `FindByOrigin` on loaded manager returns correct `Stored` values.

## Success looks like
```
ok  	forester/game	...
ok  	forester/e2e_tests	...
```
All lint checks pass.

## Failure signals
- Any test failure or lint error — treat as a blocker, do not proceed to next stage.
- `State` still having `Storage` or `StorageByOrigin` after stage 2 — architectural goal not met.
