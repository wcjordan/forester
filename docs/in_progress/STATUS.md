# STATUS

- [ ] Stage 1: Wire StorageByOrigin into State + fix deposit routing
- [ ] Stage 2: Update unit tests for StorageByOrigin

## Current state
Planning complete. Ready to implement Stage 1.

## Key decisions
- `StorageInstance.Deposit` added; `ResourceStorage.Deposit` removed (dead code).
- `OnBuilt(s *State)` → `OnBuilt(s *State, origin Point)` to pass origin at build time.
- `State.StorageByOrigin map[Point]*StorageInstance` maps structure origin → instance.
