# Plan: StorageManager + Env refactor

## Goal
Encapsulate storage truth data and derived runtime structures.
Introduce `Env` as the runtime context passed to `StructureDef` methods,
replacing the raw `*State` parameter.
Scope: storage/structure only. Player and World state migration deferred.

## Constraints / Non-goals
- Do NOT migrate `Player`, `World`, or `FoundationDeposited` off `State` (future session).
- Do NOT implement actual save/load to disk — infrastructure only.
- Keep `ResourceStorage` and `StorageInstance` types in `storage.go`.
- No new game features.

## Architecture

### New types
- `StorageState` — serializable truth: `Amounts map[Point]int` (origin → stored amount).
- `StorageManager` — runtime derived structures: `byOrigin`, `byResource`, live `amounts`.
  - `Register(origin, resource, capacity)` — called from `OnBuilt`.
  - `DepositAt(origin, amount) int` — deposits and keeps `amounts` in sync.
  - `FindByOrigin(origin) *StorageInstance`
  - `Total(r ResourceType) int`
  - `SaveData() StorageState` — snapshot of amounts.
  - `LoadFrom(s StorageState, world *World)` — rebuilds from saved amounts + world StructureIndex.
- `StorageDef` sub-interface (in `storage.go`) — extends `StructureDef` with:
  - `StorageResource() ResourceType`
  - `StorageCapacity() int`
  Used by `LoadFrom` to reconstruct instances without storing resource type in `StorageState`.
- `Env` — runtime context passed to `StructureDef` methods:
  ```go
  type Env struct {
      State  *State
      Stores *StorageManager
  }
  ```

### State changes
- `State` loses `Storage map[ResourceType]*ResourceStorage` and `StorageByOrigin map[Point]*StorageInstance`.
- `State` loses `getStorage()` and `TotalStored()`.
- `Game` gains `Stores *StorageManager`.
- `Game` gains `env() *Env` helper.
- `Game.Tick()` creates env and passes it to `Harvest` and `TickAdjacentStructures`.

### StructureDef interface
All methods switch from `*State` to `*Env`:
```go
ShouldSpawn(env *Env) bool
OnPlayerInteraction(env *Env, origin Point, now time.Time)
OnBuilt(env *Env, origin Point)
```

---

## Stage 1 — Define new types (no behavioral changes)

**Goal:** All new types in place and compiling. No existing behavior touched.

**Steps:**
1. Add `StorageDef` sub-interface to `game/storage.go`.
2. Add `StorageResource()` and `StorageCapacity()` to `logStorageDef` in `game/log_storage.go`.
3. Create `game/storage_manager.go`: `StorageState`, `StorageManager`, `NewStorageManager()`,
   `Register`, `DepositAt`, `FindByOrigin`, `Total`, `SaveData`, `LoadFrom`.
4. Create `game/env.go`: `Env` struct.
5. Verify: `make check` passes.
6. Commit: `feat: add StorageState, StorageManager, Env, StorageDef types`.

**Exit criteria:** `make check` passes; no existing files changed.

---

## Stage 2 — Wire everything together

**Goal:** `StorageManager` is live on `Game`. `State` drops raw storage maps.
`StructureDef` methods take `*Env`. All tests pass.

**Steps:**
1. Update `StructureDef` interface in `game/structure.go` to take `*Env`.
2. Update `log_storage.go`: use `env.Stores.Register`, `env.Stores.DepositAt`, `env.Stores.FindByOrigin`,
   `env.State.*` for player/world/foundation access.
3. Update `game/state.go`:
   - Remove `Storage`, `StorageByOrigin`, `getStorage()`, `TotalStored()`.
   - `Harvest(env *Env)` passes env to `maybeSpawnFoundation`.
   - `TickAdjacentStructures(env *Env, now time.Time)` passes env to structure methods.
4. Update `game/progression.go`: `maybeSpawnFoundation(env *Env)` and `def.ShouldSpawn(env)`.
5. Update `game/game.go`: add `Stores *StorageManager`; `NewWithClockAndRNG` initializes it;
   add `env()` helper; `Tick()` creates env and passes it.
6. Update `game/state_test.go`: replace direct `State` storage field construction with
   `NewStorageManager()` + `stores.Register(...)`. Replace `s.TotalStored` with `stores.Total`.
   Replace bare method calls with env-parameterized versions.
7. Update `e2e_tests/log_storage_test.go`: `g.Stores.Total(game.Wood)` instead of `g.State.TotalStored`.
8. Verify: `make check` passes.
9. Commit: `refactor: wire StorageManager+Env; StructureDef methods take *Env`.

**Exit criteria:** `make check` passes; no storage maps on `State`.

---

## Stage 3 — LoadFrom round-trip test

**Goal:** Prove `SaveData` / `LoadFrom` faithfully reconstructs the manager.

**Steps:**
1. Create `game/storage_manager_test.go` with `TestStorageManagerRoundTrip`:
   - Build a world with two `LogStorage` instances via `SetStructure` + `IndexStructure`.
   - Register both with a manager; deposit into one.
   - Call `saved := manager.SaveData()`.
   - Create a new empty manager; call `manager2.LoadFrom(saved, world)`.
   - Assert `manager2.Total(Wood)` matches original, and `FindByOrigin` returns correct `Stored`.
2. Verify: `make check` passes.
3. Commit: `test: add StorageManager SaveData/LoadFrom round-trip test`.

**Exit criteria:** `make check` passes; round-trip test green.
