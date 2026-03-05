# Road Formation

> **Status**: 2-level implementation complete (Trodden + Road). See notes below for what remains.

## Concept

Roads form organically where the player (and eventually villagers) travel repeatedly. This creates a natural record of movement patterns and rewards efficient routing.

## Mechanics

- Each tile tracks a walk count (traffic counter)
- Thresholds trigger terrain upgrades:
  - Grassland → Trodden Path (low traffic)
  - Trodden Path → Road (moderate traffic)
  - Road → Better Road (high traffic, villager contribution)
- Roads increase movement speed for all who travel them
- Both player and villagers contribute to traffic counts

## ASCII Representation

```
. = Grassland
: = Trodden path
= = Road
```

## Why Deferred

Deprioritized in favor of structure progression (Phase 2), which creates a more immediate sense of building a village. Roads work best once there are multiple destinations worth connecting (log storage, houses, depot).

## Implemented

- `WalkCount int` in `Tile` struct (existed already)
- Incremented on each player/villager move to Grassland tiles (`isRoadEligible`)
- Thresholds: Trodden=20, Road=100
- Player movement speed: Grassland=150ms, Trodden=120ms, Road=90ms
- `World.MoveCost` normalizes by road cost (90ms) so A* prefers roads (admissibility preserved)
- TUI: `:` trodden, `=` road; Ebiten: solid-color tan/gray tiles

## Not Yet Done

- Better Road (3rd level, high traffic + villager contribution)
- Road decay (WalkCount degrades unused paths over time)
- Road formation on Forest terrain (after trees cleared)
- Villager movement speed benefit from roads
