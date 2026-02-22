# VERIFY

## Primary gate
```bash
make check   # lint + test (must pass after every stage)
```

## Success looks like
- Zero lint errors from golangci-lint.
- All unit tests pass with race detector (`make test`).
- `TestLogStorageWorkflow` e2e test passes.

## Specific test cases to validate

### New unit test: deposit targets specific instance
- World with two LogStorage instances at distinct origins (A and B), each in `StorageByOrigin`.
- Player adjacent only to A → deposit goes to A.Stored; B.Stored stays 0.
- Player adjacent only to B → deposit goes to B.Stored; A.Stored stays 0.

### Existing: capacity respected
- StorageInstance at capacity → `Deposit` returns 0, no cooldown queued.

### Existing: "two adjacent instances each trigger an interaction"
- After fix: each instance deposits into its own `StorageByOrigin` entry.
- Total stored across both = 2.

### E2E: `TestLogStorageWorkflow`
- Must pass unmodified (single storage scenario — no behaviour change).
