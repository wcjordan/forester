# Plan: Villager Implementation

## Context

Add autonomous villagers that spawn when houses are built and perform two tasks:
- **Chop**: find a tree → harvest wood → walk to log storage → deposit
- **Deliver**: withdraw wood from log storage → walk to a house → consume (wood disappears)

Task is selected probabilistically each time a villager goes idle:
- P(Chop) = 1.0 − fill_ratio  (prefer chop when storage is low)
- P(Deliver) = fill_ratio      (prefer deliver when storage is high)

Fall back to the available task when the preferred one has no valid target.

Status bar additions:
- Log storage: `Log: 45/500`
- Villagers: `Villagers: 1/2` (count / house count)

---

## Stage 1 — Villager entity, spawning, and rendering

**Goal**: Villagers exist, are rendered, and spawn 1-per-house-built.

### Steps
1. Add `Villager` struct to new `game/villager.go`:
   - Fields: `X`, `Y`, `moveCooldown time.Time`
   - No tasks yet — just position + movement cooldown stub
2. Add `Villagers []*Villager` to `State` in `state.go`; initialize as `nil` slice in `newState()`
3. Add `SpawnVillager(x, y int)` on `*State` (appends a new `Villager` to the slice)
4. Modify `houseDef.OnBuilt` in `game/structures/house.go` to call `env.State.SpawnVillager(origin.X - 1, origin.Y - 1)` (one tile above/left of origin; find first in-bounds, non-structure tile)
5. Render `v` (cyan) for each villager in `render/model.go` (after player, before terrain)
6. Add `Villagers: X/Y` to status bar (X = len(Villagers), Y = house count)
7. Add unit test: house built → villager spawned

**Exit criteria**: `make check` passes; one `v` appears after building first house; status bar shows `Villagers: 1/1`.

**Commit**: `Feat: villager entity, spawning per house, rendering`

---

## Stage 2 — Task system, movement, and storage withdrawal

**Goal**: Villagers autonomously pick tasks and execute them.

### Steps
1. Add `WithdrawFrom(origin Point, amount int) int` to `StorageManager`
2. Add `TotalCapacity(r ResourceType) int` to `StorageManager` (sum across all instances)
3. Expand `Villager` struct in `game/villager.go`:
   - `Task VillagerTask` (enum: Idle, ChopWalking, ChopHarvesting, CarryingToStorage, FetchingFromStorage, DeliveringToHouse)
   - `TargetX, TargetY int` — current movement target
   - `Wood int` — inventory (max `VillagerMaxCarry = 5`)
4. Implement `pickTask(env *Env, rng *rand.Rand)` on `*Villager`:
   - Compute fill ratio from `Stores.Total(Wood) / Stores.TotalCapacity(Wood)` (clamp 0–1)
   - Roll `rng.Float64()` against fill ratio
   - Chop path: find nearest tree (scan world for Forest tiles with TreeSize > 0)
   - Deliver path: check `Stores.Total(Wood) > 0` and house exists in `StructureTypeIndex[House]`
   - Fall back if preferred target unavailable
5. Implement `Tick(env *Env, rng *rand.Rand, now time.Time)` on `*Villager`:
   - Gate on `moveCooldown` (use `DefaultMoveCooldown`)
   - Per task state, move one cardinal step toward target OR interact when adjacent
   - **ChopWalking**: move toward target tree; when adjacent harvest 1 wood; when `Wood == VillagerMaxCarry` or tree exhausted → transition to CarryingToStorage
   - **CarryingToStorage**: move toward nearest log storage origin; when adjacent call `DepositAt` for all carried wood; → Idle
   - **FetchingFromStorage**: move toward log storage origin; when adjacent call `WithdrawFrom` for up to `VillagerMaxCarry`; if got any wood → find nearest house, set target → DeliveringToHouse; else → Idle
   - **DeliveringToHouse**: move toward target house origin; when adjacent, drop wood (set `Wood = 0`); → Idle
   - **Idle**: call `pickTask`
6. Add `TickVillagers(env *Env, rng *rand.Rand, now time.Time)` to `State`; call it from `Game.Tick()` after other ticks
7. Update status bar in `render/model.go`: add `Log: X/Y` (total stored / total capacity)
8. Add tests: task selection logic, withdraw, delivery cycle

**Exit criteria**: `make check` passes; villager visibly moves and harvests trees; wood flows to log storage and gets delivered to houses; status bar shows log storage fill.

**Commit**: `Feat: villager task system — chop, carry, deliver`

---

## Non-goals (out of scope)
- BFS pathfinding (villagers may briefly get stuck on structures)
- Foreman mechanic
- Villager XP / productivity upgrades
- Villager rendering distinction between task states (e.g. carrying vs not)
