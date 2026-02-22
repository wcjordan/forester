# Plan: Typed Cooldown System

## Goal
Replace the single `DepositCooldown time.Time` field on `Player` with a typed
cooldown map (`map[CooldownType]time.Time`). Each `OnPlayerInteraction`
implementation is responsible for checking and setting its own cooldown type.
`TryDeposit` is removed; `game.Tick()` calls `TickAdjacentStructures(now)` directly.

## Non-goals
- No new cooldown types beyond `Deposit`
- No changes to game timing or deposit interval values
- No changes to render or e2e test logic beyond field/method name updates

---

## Stage 1 — Add `CooldownType` + new Player API; remove `TryDeposit`

**Files:** `game/player.go`

Steps:
1. Add `CooldownType` (exported typed int) and `Deposit CooldownType = iota` constant.
2. Replace `DepositCooldown time.Time` with `Cooldowns map[CooldownType]time.Time`.
3. Initialize `Cooldowns` map in `NewPlayer`.
4. Add `CooldownExpired(ct CooldownType, now time.Time) bool` — returns `now.After(p.Cooldowns[ct])`.
5. Add `SetCooldown(ct CooldownType, until time.Time)` — sets `p.Cooldowns[ct] = until`.
6. Delete `TryDeposit`.

Exit criteria: package compiles (other packages will temporarily fail).

---

## Stage 2 — Update `StructureDef` interface + `TickAdjacentStructures`

**Files:** `game/structure.go`, `game/state.go`

Steps:
1. Add `now time.Time` parameter to `OnPlayerInteraction` in `StructureDef` interface.
2. Update `TickAdjacentStructures(now time.Time)` to accept and thread `now` through to `entry.Def.OnPlayerInteraction(s, entry.Origin, now)`.

Exit criteria: package compiles (log_storage.go will fail at implementation site).

---

## Stage 3 — Update `logStorageDef.OnPlayerInteraction`

**File:** `game/log_storage.go`

Steps:
1. Update method signature to `OnPlayerInteraction(s *State, _ Point, now time.Time)`.
2. At top of method: if `!s.Player.CooldownExpired(Deposit, now)` → return early.
3. After `Deposit(1)`, if `deposited > 0` call `s.Player.SetCooldown(Deposit, now.Add(DepositTickInterval))`.

Exit criteria: package compiles.

---

## Stage 4 — Update `game.Tick()` call site

**File:** `game/game.go`

Steps:
1. Replace `g.State.Player.TryDeposit(g.State, now)` with `g.State.TickAdjacentStructures(now)`.

Exit criteria: `make check` passes.

---

## Stage 5 — Update tests

**Files:** `game/player_test.go`

Steps:
1. Remove/rewrite `TestTryDeposit` to test `OnPlayerInteraction` behavior directly (or via the state tick path).
2. Replace `s.Player.DepositCooldown` references with `s.Player.Cooldowns[Deposit]` / `s.Player.SetCooldown(...)`.
3. Verify the four deposit cases still have coverage:
   - cooldown not expired → no deposit
   - cooldown expired → deposit fires
   - deposit fires → cooldown is set to future
   - no wood → cooldown not set

Exit criteria: `make check` passes.

---

## Stage 6 — Commit + PR

1. Commit with clear message.
2. Push branch and open GitHub PR with `copilot` as reviewer.
3. Delete `docs/in_progress/` files.
