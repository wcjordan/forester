# Status: Villager Implementation

## Current state
- Planning complete; ready to implement Stage 1

## Stages
- [x] Stage 1: Villager entity, spawning, rendering — make check passes
- [x] Stage 2: Task system, movement, storage withdrawal — make check passes

## Done
All stages complete. Ready to clean up in_progress files.

## Key decisions
- Task selection: probabilistic (P(chop) = 1 - fill_ratio, P(deliver) = fill_ratio)
- 1 villager spawned per house built (via houseDef.OnBuilt)
- Simple cardinal movement toward target (no pathfinding)
- Wood delivered to house disappears (no house storage)
- VillagerMaxCarry = 5
