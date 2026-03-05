# Plan: Villager Path Failure Backoff (Issue #56)

## Goal
Prevent unreachable targets from becoming a CPU hotspot. When `FindPath` returns nil in `villager.go`'s `move()`, the villager currently retries A* every 300ms indefinitely. Add exponential backoff and a give-up-to-idle fallback.

## Non-goals
- No changes to `FindPath` itself
- No serialization of `pathFailures` (VillagerManager is ephemeral runtime state)
- No task-specific fallback (always go idle on give-up)

## Constants (proposed)
- `villagerPathMaxFailures = 5`
- `villagerPathBackoffBase = villagerMoveCooldown` (300ms)
- `villagerPathBackoffMax = 5 * time.Second`
- Backoff sequence: 300ms → 600ms → 1.2s → 2.4s → idle at 5th failure

## Stage 1: Implement backoff + tests

### Steps
1. Add `pathFailures int` to `Villager` struct
2. Change `move(world *World)` → `move(world *World, now time.Time)` in signature and all 4 call sites in `Tick()`
3. Add constants (`villagerPathMaxFailures`, `villagerPathBackoffBase`, `villagerPathBackoffMax`)
4. In `move()`: on `FindPath` nil → increment `pathFailures`, extend `moveCooldown` with capped exponential backoff, reset to idle if `pathFailures >= villagerPathMaxFailures`
5. In `move()`: on `FindPath` success → reset `pathFailures = 0`
6. Reset `pathFailures = 0` alongside `v.path = nil` in all target-assigning helpers: `tryAssignChopTask`, `tryAssignDeliverTask`, `headToStorage`, `headToHouse`
7. Update `TestVillagerRoutesAroundObstacle` (calls `v.move(w)` directly — needs `now` arg)
8. Add `TestVillagerPathFailureResetsToIdle` — unreachable target → idle after max failures
9. Add `TestVillagerPathFailureBackoff` — moveCooldown extended on each failure

### Exit criteria
- `make check` passes
- New tests cover backoff and idle-reset behavior
- Commit: `Fix(#56): add exponential backoff for unreachable path targets in villager`

## Stage 2: Verify and PR

### Steps
1. Run `make check` (lint + tests)
2. Open PR referencing issue #56
