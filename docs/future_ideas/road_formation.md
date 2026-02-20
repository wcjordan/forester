# Future Idea: Road Formation

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

## Implementation Notes

- Add `WalkCount int` to `Tile` struct
- Increment on each player/villager move
- Thresholds for upgrades tuned via playtesting
- Movement speed modifier applied in player update logic
