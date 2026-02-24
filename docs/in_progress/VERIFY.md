# Verify: Foundation Placement Constraints

## Commands
```
make check   # lint + test (primary gate)
make test    # race-detector tests only
```

## Success criteria
- `make check` exits 0, zero lint errors, zero test failures.
- Log storage e2e test: no change in foundation position (48,46).
- House e2e test: house spawns near spawn (50,50), NOT at player's current position.
- No new TODOs left unaddressed.

## Failure indicators
- Foundation spawns on player tile → `isValidArea` player check missing or wrong.
- Foundations spawn adjacent (0 gap) → border check not expanding by 1 in all 8 dirs.
- House spawns at player location → `SpawnAnchoredPlacer` not wired up.
- Log storage position changes → accidentally changed log storage algorithm.
