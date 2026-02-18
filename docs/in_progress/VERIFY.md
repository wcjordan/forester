# VERIFY.md — Phase 1: TUI Rendering + Player Movement

## Primary gate
```bash
make check   # lint + test — must pass after every stage
```

## Stage 1 — Static render
```bash
make run
```
Expected:
- Terminal clears and shows a grid of `.` characters
- `@` visible at center of grid
- No crashes, no stray output
- Press `q` → exits cleanly, terminal restored

Failure signals:
- Blank screen / immediate exit → bubbletea init broken
- `@` at wrong position → viewport offset math wrong
- Terminal not restored on exit → missing `tea.Quit` handling

## Stage 2 — Player movement
```bash
make run
```
Then interactively:
- Press right arrow / `d` → `@` moves right, coordinate in status line updates
- Press left arrow / `a` → `@` moves left
- Press up arrow / `w` → `@` moves up
- Press down arrow / `s` → `@` moves down
- Hold arrow key against world boundary → player stops, no crash
- Move to corner of world → viewport clamps correctly (no blank tiles outside world)

Unit tests (to be added):
- `TestMovePlayer` — moves within bounds, rejected at edges
- `TestMovePlayerBounds` — x/y never exceed world dimensions or go below 0

## Environment assumptions
- Terminal must support at least 42 cols × 22 rows for default viewport
- `golangci-lint` installed (`brew install golangci-lint`)
- Go 1.23+
