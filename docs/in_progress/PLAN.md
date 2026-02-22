# Plan: E2E Test — Log Storage Workflow

## Context

The game has solid unit tests for individual components but no end-to-end test that exercises
the full player journey: move → harvest → build → deposit. This plan adds an `e2e_tests` package
driven through the real bubbletea `model.Update()` path with `tea.KeyMsg` and exported `TickMsg`,
controlled by an injectable fake clock so the test is fast, deterministic, and covers all cooldowns.

## Decisions

- **Full-stack via bubbletea**: Tests drive `render.Model.Update()` with `tea.KeyMsg` and `TickMsg` — same code path as a real user.
- **Exported `TickMsg`**: Rename `tickMsg` → `TickMsg` in `render/model.go` so external packages can fire ticks.
- **Injected Clock**: A `Clock` interface (defined in `game/`) is injected into `game.Game` and `render.Model`. A `FakeClock` starting at `2024-01-01T00:00:00Z` gives tests explicit, deterministic time control.
- **Explicit clock advancement**: The test calls `clock.Advance(d)` before each action. A helper `tick(n)` bundles advance+TickMsg for multi-tick sequences.
- **Seed 42 world**: Uses `DefaultSeed` for deterministic layout. Player spawns at (50, 50).
- **ANSI stripping**: `View()` output is stripped of escape codes before assertions.

---

## Stage 1 — Define Clock interface + FakeClock

**Goal:** Add testable time abstraction to the `game` package with no behavior change.

**Non-goals:** Do not inject into anything yet.

**New file:** `game/clock.go`

```
type Clock interface { Now() time.Time }

type RealClock struct{}
func (RealClock) Now() time.Time { return time.Now() }

type FakeClock struct{ t time.Time }
func NewFakeClock() *FakeClock       // starts at 2024-01-01T00:00:00Z
func (f *FakeClock) Now() time.Time
func (f *FakeClock) Advance(d time.Duration)
```

**Exit criteria:** `make check` passes, no behavior change.
**Commit:** `Add Clock interface and FakeClock to game package`

---

## Stage 2 — Wire clock into game.Game + game.Player

**Goal:** Replace all `time.Now()` calls in game logic with clock-based equivalents.

**Files:** `game/game.go`, `game/player.go`

- `Game`: add `clock Clock` field; `New()` uses `RealClock{}` (unchanged externally); new `NewWithClock(clock Clock) *Game` constructor; `Tick()` uses `g.clock.Now()` for regrowth cooldown; passes `g.clock.Now()` to `TryDeposit`.
- `Player.TryDeposit(s *State, now time.Time)`: replaces both `time.Now()` calls with `now` parameter.

**Exit criteria:** `make check` passes, all existing tests pass.
**Commit:** `Wire Clock into Game and Player for testable time control`

---

## Stage 3 — Wire clock into render.Model + export TickMsg

**Goal:** Complete clock injection at the render layer; expose TickMsg for external tests.

**File:** `render/model.go`

- `type tickMsg time.Time` → `type TickMsg time.Time`; update all internal references.
- Add `clock game.Clock` field to `Model`.
- `NewModel(g)` unchanged — uses `RealClock{}` internally.
- New constructor: `NewModelWithClock(g *game.Game, clock game.Clock) Model`.
- `Update()`: `tickMsg` case → `TickMsg`; movement handlers use `m.clock.Now()` for `lastMoveTime`.
- `canMove()`: `m.clock.Now().Sub(m.lastMoveTime) >= cooldown`.

`main.go` — no changes needed.

**Exit criteria:** `make check` passes.
**Commit:** `Export TickMsg and wire Clock into render.Model`

---

## Stage 4 — E2E test: log_storage_test.go

**Goal:** Full end-to-end test covering movement, harvesting, building, depositing, and UI assertions.

**New:** `e2e_tests/log_storage_test.go`

### Helpers

- `stripANSI(s string) string` — strips ANSI escape codes from View() output.
- `newTestModel()` — returns model + FakeClock with 80×24 terminal, seed 42 world.
- `sendKey(m, key)` — fires one `tea.KeyMsg` through `model.Update()`.
- `tick(m, clock, n)` — advances clock by `HarvestTickInterval` then sends `TickMsg`, repeated n times.
- `move(m, clock, dir)` — advances clock by current tile's move cooldown then sends direction key.

### TestLogStorageWorkflow scenario

```
1. Setup: player at (50, 50), facing north, 80×24 terminal, FakeClock at 2024-01-01.

2. Move north ×3 → player at (50, 47), adjacent to forest.

3. tick() in loop until State.HasStructureOfType(GhostLogStorage)
   (TotalWoodCut >= 10 triggers ghost spawn; ~10–20 ticks expected)

4. Find ghost origin via State.ghostOriginFor(GhostLogStorage).

5. Move player onto a GhostLogStorage tile → checkGhostContact fires →
   BuildOperation starts, player nudged outside.

6. tick(30) until State.Building == nil (build completes).

7. Move player adjacent (cardinal) to LogStorage footprint.

8. Advance clock > DepositTickInterval; tick() until Player.Wood decreases.

9. Assertions:
   a. player.X, player.Y match expected coordinates.
   b. stripANSI(View()) status bar contains "Player: (X, Y)" and correct wood count.
   c. LogStorage tile exists in World at expected (gx, gy).
   d. View() shows "L" at correct screen position for LogStorage.
   e. View() shows "%" (stump) at harvested tree positions.
```

**Exit criteria:** `make check` passes including new E2E test.
**Commit:** `Add E2E test for log storage build and deposit workflow`

---

## Files Modified

| File | Change |
|---|---|
| `game/clock.go` | NEW — Clock interface + RealClock + FakeClock |
| `game/game.go` | clock field, NewWithClock(), update Tick() |
| `game/player.go` | TryDeposit(now time.Time) param |
| `render/model.go` | Export TickMsg, clock field, NewModelWithClock() |
| `e2e_tests/log_storage_test.go` | NEW — full E2E test |
