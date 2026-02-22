# Verify

## Primary gate
```bash
make check   # lint + test (must pass before each commit)
```

## Per-stage checks

| Stage | Command | Success |
|-------|---------|---------|
| 1 | `go build ./game/...` | compiles |
| 2 | `go build ./game/...` | compiles |
| 3 | `make check` | lint + all tests green |
| 4 | `make check` | lint + all tests green |

## Key test cases to watch
- `TestTryDeposit/does_not_deposit_when_cooldown_has_not_passed` — cooldown respected
- `TestTryDeposit/deposits_when_cooldown_has_passed` — deposit fires when expired
- `TestTryDeposit/sets_cooldown_after_deposit` — cooldown set by OnPlayerInteraction
- `TestTryDeposit/does_not_set_cooldown_when_nothing_deposited` — no wood → no cooldown
- `TestLogStorageWorkflow` (e2e) — full deposit cycle still works
