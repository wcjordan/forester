# VERIFY — E2E Test: Log Storage Workflow

## Primary gate

```bash
make check        # lint + tests (must be green before any commit)
```

## Stage-by-stage commands

```bash
make test         # run all tests with race detector
go test ./game/... -v                  # stage 1 & 2: clock + game changes
go test ./render/... -v               # stage 3: render changes
go test ./e2e_tests/... -v            # stage 4: new E2E test
```

## What success looks like

- `make check` exits 0 with no lint warnings and no failing tests.
- `go test ./e2e_tests/... -v` prints `PASS` and shows the `TestLogStorageWorkflow` scenario passing all assertions.
- E2E test runs in < 1 second (clock is synthetic, no real sleeps).

## What failure looks like

- Any `FAIL` or `panic` in test output.
- Lint errors (unused imports, shadowed vars, etc.).
- E2E test timeout (would indicate a loop waiting on a condition that never becomes true — add a max-iteration guard to each polling loop).

## Env assumptions

- Go toolchain available (`go version`).
- `golangci-lint` available (used by `make lint`).
- No TTY or display required — bubbletea model driven directly via `Update()`, not `tea.NewProgram`.
