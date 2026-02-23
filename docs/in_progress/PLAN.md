# Plan: Player-owned movement with cooldown + generalized structure blocking

## Goal

Consolidate all move logic into `Player.Move` so that:
- `model.go` is simplified to a single call per direction key
- `Player` owns cooldown tracking (`lastMoveTime`), cooldown checking, and position update
- Structure blocking is generalized: any tile with `Structure != NoStructure` blocks movement
- `State.Move` and `Model.canMove` are removed

## Non-goals

- No new `Vec` type (keep `dx, dy int` params)
- No changes to harvest, deposit, or regrowth logic
- No changes to `e2e_tests/` (the moveDir helper already advances time correctly)

---

## Stage 1: Add `Player.Move`, generalize structure blocking, remove `MovePlayer`

**Goal**: Player owns all move logic.

**Steps**:
1. Add `lastMoveTime time.Time` (unexported) to `Player` struct
2. Add `Move(dx, dy int, w *World, now time.Time)` to player.go:
   - Return early if cooldown not expired (use `MoveCooldownFor` on current tile; zero `lastMoveTime` = always expired)
   - Update `FacingDX/FacingDY`
   - Check bounds; return early if out of bounds
   - Check destination tile: block if `tile != nil && tile.Structure != NoStructure`
   - Update `p.X, p.Y` and `p.lastMoveTime = now`
3. Remove `MovePlayer` (fold into `Move`)
4. Update `player_test.go`: replace `MovePlayer(dx, dy, w)` calls with `Move(dx, dy, w, now)`:
   - Sequential moves advance time by cooldown (150 ms grassland) between calls
   - Add a test for cooldown: second move with unchanged `now` should be blocked
5. Update `state_test.go` "foundation blocks player movement" sub-test: replace `s.Move(1, 0)` with `s.Player.Move(1, 0, s.World, time.Now())`
6. `make check` must pass
7. Commit: "Refactor: player owns move cooldown and structure blocking"

**Exit criteria**: `make check` passes; `MovePlayer` deleted; `Player.Move` is the only public move entry point on `Player`.

---

## Stage 2: Remove `State.Move`, simplify `model.go`

**Goal**: `model.go` reduced to one call per direction key; `State.Move` gone.

**Steps**:
1. Delete `State.Move` from `state.go`
2. Remove `lastMoveTime time.Time` field from `Model`
3. Remove `canMove()` method from `model.go`
4. Replace the 4 direction cases in `Update` with:
   ```go
   case "up", "w":
       m.game.State.Player.Move(0, -1, m.game.State.World, m.clock.Now())
   case "down", "s":
       m.game.State.Player.Move(0, 1, m.game.State.World, m.clock.Now())
   case "left", "a":
       m.game.State.Player.Move(-1, 0, m.game.State.World, m.clock.Now())
   case "right", "d":
       m.game.State.Player.Move(1, 0, m.game.State.World, m.clock.Now())
   ```
5. `make check` must pass (including e2e tests)
6. Commit: "Refactor: remove State.Move, simplify model.go key handling"

**Exit criteria**: `make check` passes; no `lastMoveTime` or `canMove` in `model.go`; `State.Move` deleted.

---

## Stage 3: Push + open PR

1. Push branch `player_move` to origin
2. Open PR targeting `main`
