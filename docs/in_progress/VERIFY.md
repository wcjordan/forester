# VERIFY

## Primary gate

```bash
make check   # lint + tests with race detector — must be green at end of every stage
```

## Per-stage checks

### After Stage 1 (interface added)
```bash
make check
```
Expected: all tests pass, no lint errors. No behavior change.

### After Stage 2 (logStorageDef added)
```bash
make check
```
Expected: all tests pass. The `init()` registration runs but nothing in state.go
calls the registry yet, so no behavior change.

### After Stage 3 (state.go generalized + tests updated)
```bash
make check
```
Expected: all tests pass. Behavior identical to pre-refactor.

## What success looks like
- `make check` exits 0 at every stage
- No test cases removed — only call-site renames
- All LogStorage scenarios still covered:
  - Ghost spawns after 10 wood cut
  - Ghost does not spawn twice
  - Ghost placed on grassland between player and spawn
  - Walking into ghost starts build + nudges player
  - Build advances and completes to LogStorage tiles
  - Depositing wood decrements player wood, increments LogStorageDeposited

## What failure looks like
- Any `FAIL` in `go test` output
- Any lint error from `golangci-lint`
- A test case that was deleted rather than updated
