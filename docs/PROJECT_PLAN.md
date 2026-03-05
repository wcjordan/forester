# Forester - Game Design & Implementation Plan

## Game Concept

A city builder/simulation game where you play as a character who develops a village through organic, emergent gameplay. Instead of explicitly building structures, they develop naturally where you work and interact with the world. Features a Vampire Survivors-style auto-interaction system and roguelike card upgrade mechanics.

## Core Vision

- **Top-down character movement** - Factorio-style player traversing the map (isometric view when we add graphics)
- **Organic growth** - Roads form where you travel, structures appear where you work
- **Auto-interaction** - Minimal button presses, auto-cut trees when near (like Vampire Survivors)
- **Lead by example** - Cannot directly assign villagers, they follow and learn from you
- **Card upgrades** - Progression through XP and milestone-based rare cards

## Technology Stack

- **Language**: Go
- **UI**: Terminal-based (bubbletea TUI library)
- **Hot-reloading**: air (for rapid iteration)
- **Architecture**: Clean separation of game logic from rendering

### Libraries in use
- `bubbletea` - Elm-inspired TUI framework (primary event loop)
- `lipgloss` - Terminal color and style (ANSI)
- `air` - Live reload for Go applications

---

## Core Mechanics

### 1. Player Character
- Moves around a top-down map
- Auto-interacts with nearby objects (trees, resources)
- Carries resources on their back
- Earns XP from chopping (+1/wood), depositing (+1/wood), and completing structures (+10 player / +20 villager)
- XP milestones (50, 125, 225, 350, 500, …) pause gameplay and present a 3-card upgrade offer

### 2. World & Map
- **Target size**: 1000×1000 tiles *(current: 100×100)*
- **Generation**: Procedurally generated (cellular automata)
- **Terrain Types**: Forest, Grassland
- **Boundaries**: Fixed (not infinite)

### 3. Tree Cutting (Primary Mechanic)
- **Auto-cut**: When player is near trees, automatically cut them
- **Resource gain**: Wood accumulates on player's back (carry cap 20, upgradable to 100)
- **Space clearing**: Tree tile becomes a stump; stays Forest terrain
- **Regrowth**: Trees regrow probabilistically in forests, suppressed near structures and spawn

### 4. Road Formation *(not yet implemented)*
- **Progressive states**: Grassland → Trodden Grassland → Road
- **Frequency-based**: Requires repeated/frequent travel
- **Quality levels**: More traffic = better road quality
- **Benefit**: Better roads = faster movement speed
- **Contributors**: Both player and villagers contribute

### 5. Structures
- **Organic development**: Foundations appear when gameplay conditions are met
- **Progression chain**: Log Storage → House → Resource Depot *(Depot not yet implemented)*
- **Block regrowth**: Trees won't regrow within noGrowRadius of any structure

### 6. Experience & Upgrades
- **XP Source**: Player earns XP from chopping (+1/wood), depositing (+1/wood), and completing structures (+10 player / +20 villager)
- **XP Milestones**: 50, 125, 225, 350, 500, … (threshold grows each milestone); game pauses and presents a 3-card upgrade offer
- **Card pool** (stackable, Vampire Survivors-style): Faster harvesting, depositing, movement, building; Spawn Villager (offered when an unoccupied house exists)
- **Legacy story beats** (still active): First log storage → Expanded Carry Capacity; first house → build/deposit speed options
- **Upgrade Types** (implemented):
  - Player carry capacity (+80 max wood)
  - Foundation build speed (+10%)
  - Storage deposit speed (+10%)
  - Harvest speed (+10%)
  - Move speed (+10%)
  - Spawn Villager (places villager at an unoccupied house)

### 7. Villagers
- **Spawning**: Card-gated — "Spawn Villager" card appears in XP milestone offers when an unoccupied house exists; places a villager adjacent to a random unoccupied house. Per-house occupancy tracked in `State.HouseOccupancy`
- **Behavior**: Autonomous; probabilistic task selection based on log storage fill level
  - P(chop task) = 1 − fill_ratio; P(deliver task) = fill_ratio
  - **Chop task**: Walk to nearest tree → harvest up to 5 wood → carry to log storage
  - **Deliver task**: Fetch up to 5 wood from log storage → carry to nearest house (wood consumed)
- **Movement**: Cardinal step-toward-target (300 ms/step); avoids structures; no BFS pathfinding yet
- **Contribution**: Wood flow — not XP
- **Following behavior**: Not yet implemented (villagers work independently)
- **Foreman system**: Not yet implemented
  - Can promote a villager to "foreman"
  - Foreman continues task autonomously
  - Encourages other villagers to do the same task
  - Allows player to move on to new activities
- **No direct control**: Lead by example only

---

## MVP Implementation Phases

### Phase 1: Core Loop ✅ COMPLETE
**Goal**: Basic player movement and tree cutting

#### Features
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

#### ASCII glyphs
```
@  Player (blue)
#  Dense tree, size ≥7 (green)
t  Mid tree, size 4–6 (green)
,  Sapling, size 1–3 (green)
%  Stump, size 0 (gray)
.  Grassland
```

#### Technical components
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

### Phase 2: Structures & Progression ✅ SUBSTANTIALLY COMPLETE
**Goal**: Carry capacity, organic structure growth, and basic village progression

#### Design notes
- **Foundation mechanic**: When a spawn condition is met, a `?` foundation tile appears. Player deposits wood while adjacent to complete it.
- **Story beats**: Ordered one-shot triggers check conditions each tick and fire exactly once, in sequence. Prevent out-of-order progression.
- **World conditions**: Per-`StructureDef` `ShouldSpawn()` drives autonomous re-spawning of additional instances (e.g. more houses once first is built).
- **Placement**: Spawn-anchored (near world center) for houses; player-anchored (between player and center) for log storage. Enforces 1-tile buffer between structures; no placement under player.

#### Structure progression
1. **Log Storage (4×4)** ✅ — Foundation appears when inventory is full (≥20 wood). Costs 20 wood. Auto-deposits player wood at 100 ms/unit when adjacent. Capacity: 500 wood.
2. **House (2×2)** ✅ — Foundation appears after 50 wood deposited in storage. Costs 50 wood. Villager spawning is card-gated via the XP system (see Phase 3). Subsequent houses spawn automatically when no foundation is pending.
3. **Resource Depot** — Planned after 4 houses built. Details TBD.

#### Features
- [x] Carry capacity (20 wood max; cutting stops when full)
- [x] Status bar: `Wood: X/Y`
- [x] Foundation tile (`?`) appears when spawn condition met
- [x] Adjacent deposit to build; progress bar in status bar
- [x] Log Storage: auto-deposit to storage when adjacent; `StorageManager` aggregates instances
- [x] House: foundation triggered by 50 wood in storage; spawns a villager on completion
- [x] Story beat system (`game/story.go`): ordered, one-shot, retry-able
- [x] Milestone upgrade cards offered after key beats (carry capacity; build/deposit speed)
- [x] Card selection overlay (side-by-side for 2 cards; `1`/`2` to choose)
- [x] Structure registry (`StructureDef` interface + `init()` registration)
- [x] Resource storage system (`StorageManager`, `StorageInstance`, `ResourceStorage`)
- [x] Tree no-grow zones (radius 8 around spawn and every structure)
- [x] Circular grassland clearing around spawn point
- [x] Placement constraints: 1-tile buffer, no-under-player, spawn-anchored houses
- [ ] Resource Depot: triggered after 4 houses; details TBD
- [x] Road formation: Grassland tiles accumulate WalkCount from player/villager steps; thresholds at 20 (trodden) and 100 (road); faster movement (120ms/90ms); TUI glyphs `:` and `=`; villagers naturally prefer roads via A* cost
- [ ] XP-based card triggers (currently milestone/story-beat-based only)

#### ASCII glyphs added
```
?  Foundation (yellow)
L  Log Storage (bold yellow)
H  House (bold magenta)
```

#### Technical components
- `game/story.go` — `StoryBeat`, `storyBeats`, `maybeAdvanceStory()`
- `game/structure.go` — `StructureDef` interface, `StructureEntry`, registry
- `game/structures/log_storage.go` — `logStorageDef` (implements `StorageDef`)
- `game/structures/house.go` — `houseDef`; `OnBuilt` spawns a villager
- `game/storage.go` — `ResourceType`, `StorageDef`, `StorageInstance`, `ResourceStorage`
- `game/storage_manager.go` — `StorageManager`: `Register`, `DepositAt`, `WithdrawFrom`, `Total`, `TotalCapacity`
- `game/env.go` — `Env` (runtime context passed to `StructureDef` methods)
- `game/progression.go` — `spawnFoundationAt`, `findValidLocationNearPlayer`, `findValidLocationNearSpawn`, `isValidArea`
- `game/upgrade.go` — `UpgradeDef` interface, `upgradeRegistry`
- `game/upgrades/carry_upgrade.go` — `carryCapacityUpgrade`
- `game/upgrades/deposit_upgrades.go` — `buildSpeedUpgrade`, `depositSpeedUpgrade`

---

### Phase 3: Villagers & Automation ✅ PARTIALLY COMPLETE
**Goal**: Autonomous villagers that keep wood flowing; later XP, upgrades, and foreman

#### Features
- [x] Villager entity (`Villager` struct: position, task, wood inventory, move cooldown)
- [x] Villager spawning: card-gated via "Spawn Villager" XP upgrade; per-house occupancy tracked in `State.HouseOccupancy`
- [x] Autonomous task selection: probabilistic based on log storage fill
- [x] Chop task: walk to nearest tree → harvest multiple trees until full → carry to log storage
- [x] Deliver task: fetch wood from log storage → carry to nearest house (wood consumed)
- [x] Cardinal movement: step toward target, primary axis first, fallback if blocked
- [x] Status bar: `Log: X/Y` (stored/capacity), `Villagers: X/Y` (count/house count), `XP: n/next`
- [x] `StorageManager.WithdrawFrom` for villager fetch
- [x] XP tracking: +1/wood chopped, +1/wood deposited, +10/structure by player, +20/structure by villager
- [x] XP milestones (50, 125, 225, 350, 500, …) pause game and present 3-card upgrade offer
- [x] Upgrade cards: faster harvesting, depositing, movement, building (stackable)
- [x] Spawn Villager upgrade card (offered when unoccupied house exists)
- [ ] Upgrade cards: village improvements (villager speed, structure thresholds)
- [ ] Following behavior (villagers trail player and mirror current task)
- [ ] Foreman promotion (player promotes a villager; foreman works autonomously)
- [ ] Foreman influence (foreman encourages nearby villagers to join its task)
- [ ] Resource Depot structure

#### ASCII glyphs added
```
v  Villager (cyan)
```

#### Technical components
- `game/villager.go` — `Villager`, `VillagerTask`, `Tick()`, `pickTask()`, `move()`, helpers

---

## Architecture

### Game loop (actual)
```
bubbletea tick (100 ms):
  game.Tick()
    State.Harvest()               — player auto-harvests forward arc
    State.TickAdjacentStructures() — player interacts with adjacent structures
    State.TickVillagers()         — each villager takes one step / interaction
    World.Regrow()                — probabilistic tree regrowth (500 ms cooldown)
```

### Key patterns
- **Clock injection**: `Clock` interface → `RealClock` prod / `*FakeClock` tests
- **RNG injection**: `NewWithClockAndRNG(clock, rng)` for deterministic tests
- **Structure registry**: `RegisterStructure()` called from `init()` in `game/structures/`; blank-imported by `main.go` and `e2e_tests/`
- **Upgrade registry**: `RegisterUpgrade()` called from `init()` in `game/upgrades/`
- **Story beats**: evaluated in strict order; at most one fires per tick; retry until action succeeds
- **Tile indexing**: `World.Tiles[y][x]` (row-major); always use `TileAt(x, y)` for safe access

### Tile data
```go
type Tile struct {
    Terrain   TerrainType   // Grassland, Forest
    TreeSize  int           // Forest only: 1–10 alive, 0 = stump
    WalkCount int           // traffic counter; drives road formation (trodden >=20, road >=100)
    Structure StructureType // overlay: NoStructure, Foundation*, LogStorage, House, …
}
```

---

## Development Workflow

### Commands
```bash
make check    # lint + test (primary gate)
make test     # go test -race ./...
make lint     # golangci-lint run
make build    # compile binary
make run      # build and run
make dev      # hot-reload with air
make e2e_viz  # visual E2E playback (E2E_VISUAL=1 E2E_VISUAL_DELAY=150ms)
make format   # gofmt
```

### Testing strategy
- Unit tests for all game logic (`game/` package)
- End-to-end tests with injected clock + RNG (`e2e_tests/`)
- `make e2e_viz` for manual visual playback

---

## Future Enhancements (Post-MVP)

### Near-term
- Resource Depot (triggered after 4 houses)
- Villager following behavior
- Foreman system
- Graphical renderer (Ebitengine migration — see `docs/GRAPHICS_MIGRATION.md`)

### Medium-term
- Road formation (grassland → trodden → road from walk traffic)
- BFS pathfinding for villagers
- Map scale-up to 1000×1000
- Additional upgrade cards

### Long-term / Post-MVP
- Berry harvesting, mining, farming, fishing
- Combat system (enemies, auto-attack, defensive structures)
- Crafting (combine resources, new tools)
- Multiple biomes, weather, day/night cycle, seasons
- Trade with other settlements
- Web-based renderer (HTML5 canvas), sprite graphics, isometric view

---

## Open Questions / Future Decisions

- XP curve and level-up frequency?
- Number of upgrade options per level (2? 3?)?
- Villager population cap / housing ratio?
- Road traffic thresholds?
- Foreman mechanics (influence radius, task switching)?
- Resource Depot trigger threshold and function?

---

## Success Metrics

### Phase 2 Complete When:
- ✅ Player has carry capacity (20 wood) and status bar reflects it
- ✅ Cutting stops when full
- ✅ Log storage foundation appears when inventory is full
- ✅ Depositing wood while adjacent builds log storage
- ✅ Wood auto-deposits when adjacent to built log storage
- ✅ House foundation appears after 50 wood deposited; builds on adjacent deposit
- ✅ Milestone card offers appear after key story beats
- [ ] Resource Depot foundation appears after 4 houses built

### Phase 3 Complete When:
- ✅ Villagers spawn (card-gated via XP system) and appear on map
- ✅ Villagers autonomously collect and deliver wood (multi-tree chop before returning)
- ✅ Status bar shows log storage fill, villager/house ratio, and XP progress
- ✅ XP tracked and triggers 3-card upgrade offers at milestones
- [ ] Villagers follow player and mirror current activity
- [ ] Foreman can be promoted; works autonomously with influence
- [ ] Village feels "alive" with coordinated activity
