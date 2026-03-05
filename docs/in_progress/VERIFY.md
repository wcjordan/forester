# Verify: Issue #56 Path Backoff

## Commands
```bash
make check   # lint + test (primary gate)
make test    # race-detector tests only
```

## Success criteria
- `make check` exits 0
- `TestVillagerPathFailureResetsToIdle` passes: villager with fully-walled-off target goes idle after pathMaxFailures attempts
- `TestVillagerPathFailureBackoff` passes: moveCooldown doubles each failure, capped at villagerPathBackoffMax
- `TestVillagerRoutesAroundObstacle` still passes (updated call signature)
- All pre-existing tests still pass
