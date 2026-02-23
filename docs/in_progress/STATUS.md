# Status

## Current state
Planning complete. Ready to implement.

## Stages
- [ ] Stage 1: Define new types (StorageState, StorageManager, Env, StorageDef)
- [ ] Stage 2: Wire everything — interface changes + state migration
- [ ] Stage 3: LoadFrom round-trip test

## Key decisions
- `StorageState.Amounts map[Point]int` is the only truth — resource type/capacity derived from `StorageDef` on load.
- `StorageManager` keeps a live `amounts map[Point]int` in sync via `DepositAt`.
- `Env{State, Stores}` replaces `*State` in all `StructureDef` method signatures.
- `Player`, `World`, `FoundationDeposited` stay on `State` — deferred to future session.
- No disk save/load — infrastructure (SaveData/LoadFrom) only.

## Next action
Begin Stage 1.
