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
- **UI**: Terminal-based (using bubbletea or similar TUI library)
- **Hot-reloading**: air (for rapid iteration)
- **Architecture**: Clean separation of game logic from rendering

### Key Libraries to Consider
- `bubbletea` - Elm-inspired TUI framework (primary choice)
- `tcell` - Low-level terminal handling (if we need more control)
- `air` - Live reload for Go applications

## Core Mechanics

### 1. Player Character
- Moves around a top-down map
- Auto-interacts with nearby objects (trees, resources)
- Carries resources on their back
- Earns XP from gathering resources (not villagers)
- Triggers card/upgrade choices through XP levels

### 2. World & Map
- **Size**: 1000x1000 tiles
- **Generation**: Procedurally generated
- **Terrain Types** (MVP): Forest, Grassland
- **Boundaries**: Fixed (not infinite)

### 3. Tree Cutting (Primary Mechanic)
- **Auto-cut**: When player is near trees, automatically cut them
- **Resource gain**: Wood accumulates on player's back
- **Space clearing**: Tree tile becomes empty/grassland
- **Regrowth**: Trees can regrow in forests if no structures built

### 4. Road Formation
- **Progressive states**: Grassland → Trodden Grassland → Road
- **Frequency-based**: Requires repeated/frequent travel
- **Quality levels**: More traffic = better road quality
- **Benefit**: Better roads = faster movement speed
- **Contributors**: Both player and villagers contribute

### 5. Structures
- **Organic development**: Appear from repeated harvesting in an area
- **Progression example**: Wood Storage → Lumber Mill
- **Block regrowth**: Trees won't regrow where structures exist

### 6. Experience & Upgrades
- **XP Source**: Gathering resources (player only)
- **Card Triggers**:
  - Regular cards: XP milestones/level ups
  - Rare cards: Village progression milestones
- **Upgrade Types**:
  - Player abilities (move faster, cut faster, carry more)
  - Village improvements (villagers work faster, structures upgrade sooner)
  - Unlock mechanics (villagers can cut trees, new resource types)

### 7. Villagers
- **Population growth**: Grows over time as village expands
- **Behavior**: Follow player and help with current task
- **Contribution**: Generate resources (not XP)
- **Foreman system**:
  - Can promote a villager to "foreman"
  - Foreman continues task autonomously
  - Encourages other villagers to do the same task
  - Allows player to move on to new activities
- **No direct control**: Lead by example only

## MVP Implementation Phases

### Phase 1: Core Loop (Foundation) ✅ COMPLETE
**Goal**: Get basic player movement and tree cutting working

#### Features
- [x] Player character with movement (WASD/arrow keys)
- [x] 1000x1000 tile map data structure
- [x] Procedural map generation (cellular automata — forest + grassland)
- [x] Viewport/camera (show portion of map around player)
- [x] Tree entities on map
- [x] Auto-detection of nearby trees
- [x] Auto-cutting mechanic (timer-based, when near tree)
- [x] Resource tracking (wood counter)
- [x] Tree removal from map
- [x] Tree regrowth system (timer-based in forests)
- [x] Basic terminal rendering (ASCII)
- [x] Game loop (tick-based)
- [x] Forest movement slowdown (half speed through forest tiles)

#### ASCII Representation (Phase 1)
```
@ = Player
T = Tree
. = Grassland
# = Forest floor
```

#### Technical Components
- `main.go` - Entry point, hot-reload setup
- `game/state.go` - Core game state
- `game/player.go` - Player entity and movement
- `game/map.go` - Map data structure and generation
- `game/tree.go` - Tree entities and cutting logic
- `game/tick.go` - Game loop and update logic
- `render/terminal.go` - Terminal UI rendering
- `util/math.go` - Distance calculations, etc.

### Phase 2: Structures & Progression ✅ PARTIALLY COMPLETE
**Goal**: Add carry capacity, organic structure growth, and a basic village progression loop

#### Design Decisions
- **Village center**: Player spawn point. Houses and depot appear near here.
- **Foundation structures**: When conditions are met, a foundation tile appears on the map. Player deposits wood while adjacent to construct it. (Original design was "walk into ghost to build"; implementation uses adjacent deposit instead.)
- **Carry capacity**: Player carries max 20 wood. Cutting stops when full. Auto-deposit when adjacent to log storage.

#### Structure Progression
1. **Log Storage (4×4)** ✅ — Foundation appears when player inventory is full (≥20 wood). Deposit 20 wood while adjacent to build it. Auto-deposits wood when player is adjacent to built storage.
2. **House** — Available when 50 wood has been deposited into the log storage. Visual milestone; hooks into future villager spawning.
3. **Resource Depot** — Available when 4 houses have been built. Details TBD.

#### Features
- [x] Carry capacity (max 20 wood; cutting stops when full)
- [x] Status bar shows `Wood: 14/20`
- [x] Foundation structure indicator on map when conditions met (trigger: player inventory full)
- [x] Structure construction: deposit wood while adjacent to foundation to build
- [x] Build progress bar shown in status bar while adjacent to foundation
- [x] Log Storage (4×4): auto-deposits wood when adjacent; capacity 500 wood
- [x] Structure registry pattern (`StructureDef` interface + `init()` registration)
- [x] Resource storage system (`StorageManager`, `StorageInstance`, `ResourceStorage`)
- [x] Tree no-grow zones (trees don't regrow under/near structures)
- [x] Circular clearing around spawn point
- [ ] House: triggered by 50 wood deposited in log storage; visual only for now
- [ ] Resource Depot: triggered by 4 houses built; details TBD
- [ ] Road formation (grassland → trodden → road) — deferred, post-structures
- [ ] XP / card upgrade system — deferred, see Phase 3

#### ASCII Representation (Phase 2)
```
@ = Player
T = Tree
. = Grassland
? = Foundation (deposit wood while adjacent to build)
L = Log Storage
H = House
D = Resource Depot
```

#### Technical Components
- `game/structure.go` — `StructureDef` interface + structure registry (`structures` slice)
- `game/log_storage.go` — `logStorageDef` implementation, registered via `init()`
- `game/storage.go` — `ResourceType`, `StorageInstance`, `ResourceStorage`, `StorageDef` sub-interface
- `game/storage_manager.go` — `StorageManager` (runtime storage state), `StorageState` (serializable)
- `game/env.go` — `Env` (runtime context passed to StructureDef methods)
- `game/progression.go` — `maybeSpawnFoundation`, `findValidLocation`, `isValidArea`
- Player carry capacity and harvest stop logic live in `game/player.go`

### Phase 3: XP & Upgrades → Villagers & Automation
**Goal**: Add XP tracking, card-based upgrade system, then villagers and foreman

#### Features
- [ ] XP tracking (earned from harvesting wood)
- [ ] XP level-up thresholds and card selection screen
- [ ] Upgrade cards: player abilities (move faster, cut faster, carry more)
- [ ] Upgrade cards: village improvements (structures upgrade sooner)
- [ ] Unlock cards: new mechanics (villagers cut trees, new resource types)
- [ ] Villager entities
- [ ] Population growth system (tied to village milestones)
- [ ] Villager spawning
- [ ] Following behavior (villagers follow player)
- [ ] Task detection (villagers see what player is doing)
- [ ] Villager resource gathering (contribute to village)
- [ ] Foreman promotion mechanic
- [ ] Foreman autonomous behavior (continue task)
- [ ] Foreman influence (encourage other villagers)
- [ ] Milestone tracking (trigger rare cards)

#### ASCII Representation (Phase 3)
```
@ = Player
v = Villager
V = Foreman
T = Tree
. = Grassland
: = Trodden path
= = Road
S = Storage structure
M = Mill structure
```

#### Technical Components
- `game/villager.go` - Villager entity and behavior
- `game/population.go` - Population growth system
- `game/foreman.go` - Foreman promotion and autonomous behavior
- `game/milestone.go` - Milestone tracking for rare cards

## Architecture Design

### Game Loop
```
while running:
    1. Handle Input (movement commands)
    2. Update Game State (tick)
        - Move player
        - Auto-interact (cut trees)
        - Update villagers
        - Update roads (traffic)
        - Update structures
        - Check for level ups
        - Regrow trees
    3. Render (terminal output)
    4. Sleep until next tick (60 FPS target)
```

### Entity System
- **Player**: Position, inventory, stats, XP
- **Tree**: Position, growth stage, regrowth timer
- **Villager**: Position, assigned task, following target
- **Foreman**: Position, autonomous task, influence radius
- **Structure**: Position, type, level, area of effect

### Map/Tile System
```go
type Tile struct {
    Terrain     TerrainType  // Grassland, Forest
    WalkCount   int          // Traffic counter
    RoadLevel   int          // 0=none, 1=trodden, 2=road, 3+=better road
    Entity      Entity       // Tree, Structure, etc
    ActivityMap map[string]int // Track activity types (tree cutting, etc)
}
```

### Separation of Concerns
- **Game Logic**: Pure Go functions, no rendering dependencies
- **Rendering**: Swappable (terminal now, web later)
- **State Management**: Single source of truth for game state
- **Input Handling**: Abstract input layer (keyboard now, mouse later)

## Development Workflow

### Setup
1. Initialize Go module: `go mod init forester`
2. Install dependencies: `bubbletea`, `tcell`
3. Install air: `go install github.com/cosmtrek/air@latest`
4. Create `.air.toml` configuration
5. Run with hot-reload: `air`

### Testing Strategy
- Unit tests for game logic (tree cutting, road formation, etc.)
- Manual playtesting for balance and feel
- Debug view to inspect game state (tile info, entity counts, etc.)

### Iteration Approach
1. Build minimal feature
2. Test in game
3. Tune parameters (cutting speed, road thresholds, etc.)
4. Iterate quickly with hot-reload

## Future Enhancements (Post-MVP)

### Additional Resources & Activities
- Berry harvesting (from bushes)
- Mining (stone, ore)
- Farming (planting and harvesting)
- Fishing (near water)

### Combat System
- Enemies/threats
- Combat mechanics (auto-attack like Vampire Survivors)
- Defensive structures
- Combat upgrades

### Crafting
- Combine resources
- Create tools/equipment
- Unlock new structures

### Advanced Features
- Multiple biomes (desert, mountains, swamps)
- Weather system
- Day/night cycle
- Seasons affecting growth/gameplay
- Trade with other settlements
- Quests/objectives

### Rendering Upgrades
- Web-based renderer (HTML5 canvas)
- Sprite graphics (2D pixel art)
- Isometric view (like original vision)
- Animations
- Particle effects

## Success Metrics

### Phase 1 Complete When:
- Can move player around map
- Trees auto-cut when nearby
- Wood counter increases
- Trees regrow over time
- Map generates with varied forest patches

### Phase 2 Complete When:
- ✅ Player has a carry capacity (20 wood) and status bar reflects it
- ✅ Cutting stops when player is full
- ✅ Log storage foundation appears when player is full
- ✅ Depositing wood while adjacent builds the log storage (20 wood cost)
- ✅ Wood auto-deposits when player is adjacent to built log storage
- [ ] House foundation appears after 50 wood deposited; builds on contact
- [ ] Resource depot foundation appears after 4 houses built

### Phase 3 Complete When:
- Villagers spawn and follow player
- Villagers contribute to resource gathering
- Can promote villagers to foremen
- Foremen work autonomously
- Village feels "alive" with activity

## Design Principles

1. **Minimal Input, Maximum Expression**: Like Vampire Survivors, focus on positioning and movement, not micro-management
2. **Organic Over Explicit**: Let systems emerge from interaction rather than explicit building
3. **Lead by Example**: Player teaches through action, not commands
4. **Rapid Iteration**: Hot-reload and simple rendering enable fast experimentation
5. **Separation of Concerns**: Keep game logic independent of rendering for future flexibility

## Open Questions / Future Decisions

- Exact XP curve and level up frequency?
- How many upgrade options per level (3? 4?)?
- Structure upgrade thresholds (how many trees = storage?)?
- Villager population growth rate?
- Road traffic thresholds (how many walks = road?)?
- Tree regrowth timer (how long until respawn?)?
- Foreman mechanics details (influence radius, task switching?)?

These will be determined through playtesting and iteration.

---

**Next Steps**: Begin Phase 3 — XP tracking and card-based upgrade system, then house/depot structures to complete Phase 2's progression chain.
