# Verification

## Primary gate

```bash
make check   # lint + tests with race detector — must pass clean
```

## Key behaviors to verify

1. Player can move in all 4 directions (existing movement tests)
2. Movement is blocked at world bounds (existing bounds tests)
3. Movement is blocked by any structure tile (generalized, not just LogStorage/Foundation)
4. Cooldown: second move with same timestamp is blocked; move after cooldown elapses succeeds
5. Facing direction updates correctly on each move attempt that passes cooldown
6. `model.go` has no `lastMoveTime` field and no `canMove()` method
7. `State.Move` does not exist
8. Full e2e workflow (TestLogStorageWorkflow) passes: movement, blocking, building, depositing

## Success looks like

```
ok  	forester/game	  (no failures)
ok  	forester/e2e_tests  (no failures)
--- PASS
```

## Failure signals

- Any test failure = stop, diagnose before proceeding
- Lint warnings about unused fields/methods = clean up
- Race detector errors = fix before committing
