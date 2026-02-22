# PLAN: Deposit into specific adjacent storage

## Problem
`logStorageDef.OnPlayerInteraction` calls `s.getStorage(Wood).Deposit(1)`, which deposits
into the first non-full `StorageInstance` in the global `ResourceStorage` — regardless of
which structure the player is actually standing next to. There is also no per-instance
capacity enforcement at the interaction layer.

## Goal
- Deposit routes to the **specific `StorageInstance`** that belongs to the adjacent
  `LogStorage` structure instance.
- Deposit respects that instance's capacity; a full instance silently skips (no cooldown).
- No behaviour change when exactly one storage exists.

## Non-goals
- Multi-resource-type support (future work).
- UI changes.

---

## Stage 1 — Wire `StorageByOrigin` into `State`

**Steps:**
1. Add `Deposit(amount int) int` method to `StorageInstance` in `storage.go`.
2. Remove now-dead `ResourceStorage.Deposit` method from `storage.go`.
3. Add `StorageByOrigin map[Point]*StorageInstance` field to `State` struct in `state.go`.
4. Initialise the map in `newState()`.
5. Change `StructureDef.OnBuilt(s *State)` → `OnBuilt(s *State, origin Point)` in `structure.go`.
6. Update `AdvanceBuild` in `state.go` to pass `Point{s.Building.X, s.Building.Y}`.
7. Update `logStorageDef.OnBuilt` to store the instance: `s.StorageByOrigin[origin] = inst`.
8. Update `logStorageDef.OnPlayerInteraction` to look up the specific instance and call
   `inst.Deposit(1)` rather than the global `Deposit`.

**Exit criteria:**
- `make check` passes.
- A new unit test in `state_test.go` verifies that two adjacent storage instances each
  receive deposits only in their own instance.
- Existing e2e `TestLogStorageWorkflow` passes unchanged.

**Commit:** `Fix: deposit into specific adjacent storage instance`

---

## Stage 2 — Update tests that build `StorageByOrigin` manually

**Steps:**
1. Update `makeDepositState` helper and the "two adjacent instances" test in `state_test.go`
   to populate `StorageByOrigin` alongside the existing `Storage` map.
2. Verify `make check` passes.

**Exit criteria:**
- All unit and e2e tests pass.
- No linter errors.

**Commit:** `Test: update storage unit tests for StorageByOrigin`
