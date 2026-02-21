# PLAN: Single Tick API — Move Game Loop Orchestration to Game

## Goal
Give the render layer a single method to advance the game each tick.
`game.Game` becomes the orchestrator; `render.Model` becomes a pure
input/display adapter with no game logic of its own.

## Non-Goals
- No subsystem/component architecture yet (that's the next step when warranted)
- No changes to `State` methods — they stay as-is
- No changes to how input (`State.Move`) is wired
- No changes to the two-timer approach (`tickMsg` / `regrowTickMsg`) in render

---

## Stage 1 — Add `Game.Tick()` and `Game.RegrowTick()` (additive)

**File:** `game/game.go`

### Changes
- Add `depositCooldown time.Time` field to `Game`
- Move `DepositTickInterval` constant from `render/model.go` to `game/game.go`
  (it's a game-logic rate, not a display concern)
- Add `Game.Tick()` — sequences the 100ms harvest tick:
  ```go
  func (g *Game) Tick() {
      g.State.Harvest()
      g.State.AdvanceBuild()
      if time.Now().After(g.depositCooldown) {
          before := g.State.LogStorageDeposited
          g.State.TickAdjacentStructures()
          if g.State.LogStorageDeposited > before {
              g.depositCooldown = time.Now().Add(DepositTickInterval)
          }
      }
  }
  ```
- Add `Game.RegrowTick()` — wraps the 20s regrowth tick:
  ```go
  func (g *Game) RegrowTick() {
      g.State.Regrow()
  }
  ```

### Exit criteria
- `make check` passes; no behavior change (render still calls old methods; new
  methods exist but are not yet wired up)

Commit: `Add Game.Tick() and Game.RegrowTick(); move DepositTickInterval to game`

---

## Stage 2 — Wire render to the new API, remove old wiring (cleanup)

**File:** `render/model.go`

### Changes
- Replace the `tickMsg` handler body with `m.game.Tick()`
- Replace the `regrowTickMsg` handler body with `m.game.RegrowTick()`
- Remove `depositCooldown time.Time` field from `Model`
- Remove `DepositTickInterval` constant (now in game package)
- Remove `"time"` import if no longer needed (check; `lastMoveTime` still uses it)

### Before / After

**Before:**
```go
case tickMsg:
    m.game.State.Harvest()
    m.game.State.AdvanceBuild()
    if time.Now().After(m.depositCooldown) {
        before := m.game.State.LogStorageDeposited
        m.game.State.TickAdjacentStructures()
        if m.game.State.LogStorageDeposited > before {
            m.depositCooldown = time.Now().Add(DepositTickInterval)
        }
    }
    return m, doTick()

case regrowTickMsg:
    m.game.State.Regrow()
    return m, doRegrowTick()
```

**After:**
```go
case tickMsg:
    m.game.Tick()
    return m, doTick()

case regrowTickMsg:
    m.game.RegrowTick()
    return m, doRegrowTick()
```

### Exit criteria
- `make check` passes
- Render model no longer calls any `State` methods for ticking (only `Move` for input)
- Render model has no game-logic fields (`depositCooldown` gone)

Commit: `Wire render to Game.Tick/RegrowTick; remove game logic from Model`

---

## What render is allowed to do after this refactor

| Allowed | Not allowed |
|---|---|
| `m.game.Tick()` | `m.game.State.Harvest()` |
| `m.game.RegrowTick()` | `m.game.State.AdvanceBuild()` |
| `m.game.State.Move(dx, dy)` — input | `m.game.State.TickAdjacentStructures()` |
| Read `State` fields for display | Mutate `State` fields for timing |

`State.Move` stays in render because it's input handling — the render layer
translates a keypress into a game action. If we ever abstract input, that
moves too, but that's out of scope.
