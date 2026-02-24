# Verify

## Primary gate
```
make check   # lint + test (must pass after every stage)
```

## Stage 1 success
- `TestGenerateWorld_SpawnClear` passes with circular radius-5 check.
- No tile at Euclidean distance > 5 from center is forced to Grassland by Step 3.
- Corner tiles at (±5, 0) and (0, ±5) are clear; diagonal corner (±4, ±4) is also within radius (dist ≈ 5.66 > 5, so NOT cleared).

## Stage 2 success
- New subtest `forest within spawn no-grow zone does not grow` passes.
- New subtest `forest within building no-grow zone does not grow` passes.
- Existing growth subtests still pass (tile moved outside no-grow zone).
- `make check` passes clean.
