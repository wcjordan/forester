# Forester - Game Design & Implementation Plan

## Game Concept

A city builder/simulation game where you play as a character who develops a village through organic, emergent gameplay. Instead of explicitly building structures, they develop naturally where you work and interact with the world. Features a Vampire Survivors-style auto-interaction system and roguelike card upgrade mechanics.

## Core Vision

- **Top-down character movement** - Factorio-style player traversing the map
- **Organic growth** - Roads form where you travel, structures appear where you work
- **Auto-interaction** - Minimal button presses, auto-cut trees when near (like Vampire Survivors)
- **Lead by example** - Cannot directly assign villagers, they follow and learn from you
- **Card upgrades** - Progression through XP and milestone-based rare cards

## Technology Stack

- **Language**: Go
- **Primary renderer**: Ebitengine (2D, compiles to WASM)
- **TUI fallback**: bubbletea (`--tui` flag), lipgloss
- **Hot-reloading**: air

---

## Core Mechanics

### Player Character
- Moves around a top-down map; auto-interacts with nearby objects (trees, resources)
- Carries resources on their back (carry cap 20, upgradable)
- Earns XP from chopping (+1/wood), depositing (+1/wood), and completing structures (+10 player / +20 villager)
- XP milestones (50, 125, 225, 350, 500, …) pause gameplay and present a 3-card upgrade offer

### World & Map
- **Target size**: 1000×1000 tiles *(current: 100×100)*
- **Generation**: Procedurally generated (cellular automata)
- **Terrain Types**: Forest, Grassland
- **Boundaries**: Fixed (not infinite)

### Tree Cutting (Primary Mechanic)
- **Auto-cut**: Player automatically cuts trees in a forward arc
- **Resource gain**: Wood accumulates on player's back; cutting stops when full
- **Regrowth**: Trees regrow probabilistically in forests, suppressed near structures and spawn

### Road Formation
- **Progressive states**: Grassland → Trodden Path → Road (→ Better Road planned)
- **Frequency-based**: Tile WalkCount increments on each player/villager step; thresholds at 20 (trodden) and 100 (road)
- **Benefit**: Movement speeds — Grassland=150ms, Trodden=120ms, Road=90ms
- **A\* weighting**: `World.MoveCost` normalizes by road cost so pathfinding naturally prefers roads

### Structures
- **Organic development**: Foundations appear when gameplay conditions are met; player deposits wood while adjacent to complete
- **Current chain**: Log Storage (4×4, triggered by full inventory) → House (2×2, triggered by 50 wood in storage) → Resource Depot (planned)
- **Block regrowth**: Trees won't regrow within noGrowRadius of any structure

### Experience & Upgrades
- **XP Sources**: Chopping (+1/wood), depositing (+1/wood), completing structures (+10/+20)
- **XP Milestones**: Growing gaps (50, 125, 225, 350, 500, …); game pauses for 3-card offer
- **Card pool** (stackable, Vampire Survivors-style): Faster harvesting/depositing/movement/building, Spawn Villager, village improvement cards (planned)

### Villagers
- **Spawning**: Card-gated — "Spawn Villager" offered at XP milestone when an unoccupied house exists; per-house occupancy tracked in `State.HouseOccupancy`
- **Behavior**: Autonomous probabilistic task selection — P(chop) = 1 − fill_ratio, P(deliver) = fill_ratio
- **Movement**: A* pathfinding via `geom.FindPath` (Manhattan heuristic, terrain-cost-aware); exponential backoff for unreachable targets
- **Following behavior**: Not yet implemented — villagers will trail player and mirror current task
- **Foreman system**: Not yet implemented — see `docs/FOLLOW_THROUGH.md`

---

## Active Work

Phase 3 (Villagers & Automation) is partially complete. Remaining items:

- [ ] Village improvement upgrade cards (villager speed, carry capacity, storage, spawn thresholds)
- [ ] Following behavior (villagers trail player and mirror current task)
- [ ] Foreman promotion and influence radius
- [ ] Resource Depot structure (triggered after 4 houses; gates Tier 1 progression)

See `docs/FOLLOW_THROUGH.md` for the full list including sprite, web deployment, road, and card backlog.

---

## Architecture

### Game loop
```
Ebitengine tick (100 ms):
  game.Tick()
    State.Harvest()                — player auto-harvests forward arc
    State.TickAdjacentStructures() — player interacts with adjacent structures
    State.TickVillagers()          — each villager takes one step / interaction
    World.Regrow()                 — probabilistic tree regrowth (500 ms cooldown)
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

```bash
make check    # lint + test (primary gate)
make test     # go test -race ./...
make lint     # golangci-lint run
make build    # compile binary
make run      # build and run
make dev      # hot-reload with air
make e2e_viz  # visual E2E playback
make format   # gofmt
make wasm     # compile WASM binary
```

Testing strategy:
- Unit tests for all game logic (`game/` package)
- End-to-end tests with injected clock + RNG (`e2e_tests/`)
- `make e2e_viz` for manual visual playback

---

## Roadmap

| Where | What |
|---|---|
| `docs/FOLLOW_THROUGH.md` | Near-term concrete tasks (S5 villager animation, G3 web deploy, Phase 3 remaining, road/card backlog) |
| `docs/future_ideas/VILLAGE_PROGRESSION.md` | Full village tier design (Tiers 1–4, industries, population types) |
| `docs/completed/MVP_PHASES.md` | Phase 1, 2, 3 implementation record |
| `docs/completed/GRAPHICS_MIGRATION.md` | Ebitengine migration reference (G0–G2 complete) |
| `docs/completed/BETTER_SPRITES.md` | Sprite improvement reference (S1–S4 complete) |

### Long-term / post-Tier-1
- Berry harvesting, mining, farming, fishing
- Multiple biomes, weather, day/night cycle, seasons
- Combat system (enemies, defensive structures)
- Trade with other settlements
- Map scale-up to 1000×1000

---

## Open Questions

- XP curve and level-up frequency as game expands?
- Villager population cap / housing ratio?
- Foreman mechanics (eligibility criteria, influence radius, task switching)?
- Resource Depot trigger threshold and exact function?
