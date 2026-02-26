# Status: Villager Implementation

## Current state
- Planning complete; ready to implement Stage 1

## Stages
- [ ] Stage 1: Villager entity, spawning, rendering
- [ ] Stage 2: Task system, movement, storage withdrawal

## Next
- Begin Stage 1: create `game/villager.go`, update `state.go`, `house.go`, `render/model.go`

## Key decisions
- Task selection: probabilistic (P(chop) = 1 - fill_ratio, P(deliver) = fill_ratio)
- 1 villager spawned per house built (via houseDef.OnBuilt)
- Simple cardinal movement toward target (no pathfinding)
- Wood delivered to house disappears (no house storage)
- VillagerMaxCarry = 5
