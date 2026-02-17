# Verification

## Commands
```bash
make check        # lint + test (primary gate)
make build        # binary compiles
make run          # runs and exits cleanly
go test -v ./...  # all tests pass verbose
```

## Success
- All commands exit 0
- No lint warnings
- All tests pass
- Binary runs without errors

## Failure
- Any non-zero exit code
- Lint warnings or errors
- Test failures
- Compilation errors
