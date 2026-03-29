# Forester — Game Design Reference

Current mechanics as implemented.

---

## Player Character

- Moves around a top-down map; auto-interacts with nearby objects (trees, resources)
- Carries resources on their back (carry cap 20, upgradable)
- Earns XP from chopping (+1/wood), depositing (+1/wood), and completing structures (+10 player / +20 villager)
- XP milestones (50, 125, 225, 350, 500, …) pause gameplay and present a 3-card upgrade offer

---

## World & Map

- **Target size**: 1000×1000 tiles *(current: 100×100)*
- **Generation**: Procedurally generated (cellular automata, seed 42)
- **Terrain Types**: Forest, Grassland
- **Boundaries**: Fixed (not infinite)

---

## Tree Cutting (Primary Mechanic)

- **Auto-cut**: Player automatically cuts trees in the three-tile forward arc
- **Resource gain**: Wood accumulates on player's back; cutting stops when full
- **Regrowth**: Trees regrow probabilistically in forests (1-in-40 per tick), suppressed near structures and spawn

---

## Road Formation

- **Progressive states**: Grassland → Trodden Path → Road (→ Better Road planned)
- **Frequency-based**: Tile `WalkCount` increments on each player/villager step; thresholds at 20 (trodden) and 100 (road)
- **Benefit**: Movement speeds — Grassland=150ms, Trodden=120ms, Road=90ms
- **A\* weighting**: `World.MoveCost` normalizes by road cost so pathfinding naturally prefers roads

---

## Structures

- **Organic development**: Foundations appear when gameplay conditions are met; player deposits wood while adjacent to complete
- **Current chain**: Log Storage (4×4, triggered by full inventory) → House (2×2, triggered by 50 wood in storage) → Resource Depot (planned)
- **Block regrowth**: Trees won't regrow within `noGrowRadius` of any structure

---

## Experience & Upgrades

- **XP Sources**: Chopping (+1/wood), depositing (+1/wood), completing structures (+10 player / +20 villager)
- **XP Milestones**: Growing gaps (50, 125, 225, 350, 500, …); game pauses for a 3-card offer
- **Card pool** (stackable, Vampire Survivors-style): Faster harvesting / depositing / movement / building, Spawn Villager, village improvement cards (planned)

---

## Villagers

- **Spawning**: Card-gated — "Spawn Villager" offered at XP milestone when an unoccupied house exists; per-house occupancy tracked in `State.HouseOccupancy`
- **Behavior**: Autonomous probabilistic task selection — P(chop) = 1 − fill_ratio, P(deliver) = fill_ratio
- **Movement**: A* pathfinding via `geom.FindPath` (Manhattan heuristic, terrain-cost-aware); exponential backoff for unreachable targets
- **Following behavior**: Not yet implemented — villagers will trail player and mirror current task
- **Foreman system**: Not yet implemented — see `docs/FOLLOW_THROUGH.md`

---

## Game Loop

```
Ebitengine tick (100 ms):
  game.Tick()
    State.Harvest()                — player auto-harvests forward arc
    State.TickAdjacentStructures() — player interacts with adjacent structures
    State.TickVillagers()          — each villager takes one step / interaction
    World.Regrow()                 — probabilistic tree regrowth (500 ms cooldown)
```

### Tile data
```go
type Tile struct {
    Terrain   TerrainType   // Grassland, Forest
    TreeSize  int           // Forest only: 1–10 alive, 0 = stump
    WalkCount int           // traffic counter; drives road formation (trodden >=20, road >=100)
    Structure StructureType // overlay: NoStructure, Foundation*, LogStorage, House, …
}
```
