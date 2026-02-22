# Verify

## Primary gate
```bash
make check   # lint + test (run after every stage)
```

## Per-stage checks

### Stage 1
- `make check` passes
- `TestIndexStructure` in world_test.go: single-tile and 4×4 multi-tile cases pass

### Stage 2
- `make build` — compiler confirms no missed interface implementations
- `make check` passes

### Stage 3
- `make check` passes
- `TestTickAdjacentStructures` all sub-tests pass
- New test: two adjacent LogStorage instances → `OnPlayerInteraction` called twice (two deposits)
- `IsAdjacentToStructure` and `TestIsAdjacentToStructure` are deleted

## Success criteria
- All tests pass with race detector (`make test`)
- No lint errors (`make lint`)
- `TickAdjacentStructures` no longer iterates the structures registry
- `IsAdjacentToStructure` is removed
