# VERIFY

## Primary gate
```bash
make check   # lint + tests — must be green after every stage
```

## Stage 1 (additive)
```bash
make check
```
Expected: passes. `Game.Tick()` and `Game.RegrowTick()` exist but are not yet
called. Render still calls old methods — no behavior change, no double-ticking.

## Stage 2 (wiring)
```bash
make check
```
Expected: passes.

Manually verify (via `make run`):
- Player can still cut trees and wood count increases
- Log Storage ghost appears after 10 wood cut
- Walking into ghost starts and completes the build
- Wood auto-deposits when adjacent to Log Storage (at the same rate as before)
- Trees regrow over time

## What success looks like
- `make check` exits 0 at both stages
- No test cases changed (all existing tests still pass as-is)
- `render/model.go` has no game-timing logic after Stage 2

## What failure looks like
- Deposits happen at wrong rate (double-deposit if old render wiring left in)
- Lint error from unused `depositCooldown` field left on Model
- `time` import warning if removed prematurely (lastMoveTime still needs it)
