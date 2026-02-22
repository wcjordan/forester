# Verify

## Primary gate
```bash
make check   # lint + tests (with race detector)
```

## Success
- All tests pass (no new failures)
- No lint errors

## Failure signals
- Compile error: missing function / wrong package
- Test failure: behavior changed (should not happen — pure move)
- Lint: unused import or duplicate symbol
