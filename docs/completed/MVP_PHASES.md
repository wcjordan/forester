# MVP Implementation Record

Completed and substantially-complete implementation phases. Reference only.

---

## Phase 1: Core Loop ✅ COMPLETE

**Goal**: Basic player movement and tree cutting

### Features
- [x] Player character with movement (WASD/arrow keys)
- [x] 100×100 tile map (procedural, cellular automata; target 1000×1000)
- [x] Viewport/camera centered on player
- [x] Forest tiles with tree sizes (1–10); stumps at 0
- [x] Auto-harvesting forward arc (timer-based, 100 ms interval)
- [x] Wood carry counter
- [x] Tree regrowth (probabilistic, 500 ms cooldown, 1-in-40 odds)
- [x] Basic terminal rendering (ASCII glyphs + lipgloss color)
- [x] Game loop (100 ms tick via bubbletea)
- [x] Forest movement slowdown (300 ms vs 150 ms on grassland)

### ASCII glyphs
```
@  Player (blue)
#  Dense tree, size ≥7 (green)
t  Mid tree, size 4–6 (green)
,  Sapling, size 1–3 (green)
%  Stump, size 0 (gray)
.  Grassland
```

### Technical components
- `main.go` — entry point
- `game/game.go` — `Game`, `Tick()` orchestrator
- `game/state.go` — `State` (Player + World)
- `game/player.go` — movement, `HarvestAdjacent()`, cooldowns
- `game/world.go` — grid, regrowth, no-grow zones, structure indexes
- `game/worldgen.go` — procedural terrain (cellular automata, seed 42)
- `game/tile.go` — `Tile`, `TerrainType`, `StructureType`
- `game/clock.go` — `Clock` interface, `RealClock`, `FakeClock`
- `render/model.go` — bubbletea `Model`, `View()`, `Update()`

---

## Phase 2: Structures & Progression ✅ SUBSTANTIALLY COMPLETE

**Goal**: Carry capacity, organic structure growth, and basic village progression

### Design notes
- **Foundation mechanic**: When a spawn condition is met, a `?` foundation tile appears. Player deposits wood while adjacent to complete it.
- **Story beats**: Ordered one-shot triggers check conditions each tick and fire exactly once, in sequence. Prevents out-of-order progression.
- **World conditions**: Per-`StructureDef` `ShouldSpawn()` drives autonomous re-spawning of additional instances (e.g. more houses once first is built).
- **Placement**: Spawn-anchored (near world center) for houses; player-anchored (between player and center) for log storage. Enforces 1-tile buffer between structures; no placement under player.

### Structure progression
1. **Log Storage (4×4)** ✅ — Foundation appears when inventory is full (≥20 wood). Costs 20 wood. Auto-deposits player wood at 100 ms/unit when adjacent. Capacity: 500 wood.
2. **House (2×2)** ✅ — Foundation appears after 50 wood deposited in storage. Costs 50 wood. Villager spawning is card-gated via the XP system.
3. **Resource Depot** — Planned after 4 houses built. See `docs/FOLLOW_THROUGH.md`.

### Features
- [x] Carry capacity (20 wood max; cutting stops when full)
- [x] Status bar: `Wood: X/Y`
- [x] Foundation tile (`?`) appears when spawn condition met
- [x] Adjacent deposit to build; progress bar in status bar
- [x] Log Storage: auto-deposit to storage when adjacent; `StorageManager` aggregates instances
- [x] House: foundation triggered by 50 wood in storage
- [x] Story beat system (`game/story.go`): ordered, one-shot, retry-able
- [x] Milestone upgrade cards offered after key story beats
- [x] Card selection overlay (side-by-side; `1`/`2`/`3` to choose)
- [x] Structure registry (`StructureDef` interface + `init()` registration)
- [x] Resource storage system (`StorageManager`, `StorageInstance`, `ResourceStorage`)
- [x] Tree no-grow zones (radius 8 around spawn and every structure)
- [x] Circular grassland clearing around spawn point
- [x] Placement constraints: 1-tile buffer, no-under-player, spawn-anchored houses
- [x] Road formation: WalkCount → Trodden (≥20) → Road (≥100); speeds 150/120/90ms; A* prefers roads

### ASCII glyphs added
```
?  Foundation (yellow)
L  Log Storage (bold yellow)
H  House (bold magenta)
:  Trodden path
=  Road
```

### Technical components
- `game/story.go` — `StoryBeat`, `storyBeats`, `maybeAdvanceStory()`
- `game/structure.go` — `StructureDef` interface, `StructureEntry`, registry
- `game/structures/log_storage.go` — `logStorageDef` (implements `StorageDef`)
- `game/structures/house.go` — `houseDef`
- `game/storage.go` — `ResourceType`, `StorageDef`, `StorageInstance`, `ResourceStorage`
- `game/storage_manager.go` — `StorageManager`: `Register`, `DepositAt`, `WithdrawFrom`, `Total`, `TotalCapacity`
- `game/env.go` — `Env` (runtime context passed to `StructureDef` methods)
- `game/progression.go` — `spawnFoundationAt`, `findValidLocationNearPlayer`, `findValidLocationNearSpawn`, `isValidArea`
- `game/upgrade.go` — `UpgradeDef` interface, `upgradeRegistry`
- `game/upgrades/carry_upgrade.go` — `carryCapacityUpgrade`
- `game/upgrades/deposit_upgrades.go` — `buildSpeedUpgrade`, `depositSpeedUpgrade`

---

## Phase 3: Villagers & Automation ✅ PARTIALLY COMPLETE

**Goal**: Autonomous villagers that keep wood flowing; XP, upgrades, and foreman

For remaining work, see `docs/FOLLOW_THROUGH.md`.

### Implemented features
- [x] Villager entity (`Villager` struct: position, task, wood inventory, move cooldown)
- [x] Villager spawning: card-gated via "Spawn Villager" XP upgrade; per-house occupancy tracked in `State.HouseOccupancy`
- [x] Autonomous task selection: probabilistic based on log storage fill level
- [x] Chop task: walk to nearest tree → harvest multiple trees until full → carry to log storage
- [x] Deliver task: fetch wood from log storage → carry to nearest house (wood consumed)
- [x] A* pathfinding via `geom.FindPath`; exponential backoff + idle-reset for unreachable targets
- [x] Status bar: `Log: X/Y` (stored/capacity), `Villagers: X/Y` (count/house count), `XP: n/next`
- [x] `StorageManager.WithdrawFrom` for villager fetch
- [x] XP tracking: +1/wood chopped, +1/wood deposited, +10/structure by player, +20/structure by villager
- [x] XP milestones (50, 125, 225, 350, 500, …) pause game and present 3-card upgrade offer
- [x] Upgrade cards: faster harvesting, depositing, movement, building (stackable)
- [x] Spawn Villager upgrade card (offered when unoccupied house exists)

### ASCII glyphs added
```
v  Villager (cyan)
```

### Technical components
- `game/villager.go` — `Villager`, `VillagerTask`, `Tick()`, `pickTask()`, `move()`, helpers
- `game/xp.go` — `AwardXP()`, milestone thresholds, card offer selection
