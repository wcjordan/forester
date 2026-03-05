# STATUS: Road Formation

## Stages
- [ ] Stage 1: Core road logic + traffic counting
- [ ] Stage 2: Movement speed + pathfinding cost
- [ ] Stage 3: Rendering (TUI + Ebiten)
- [ ] Stage 4: Documentation + PR

## Current
Starting Stage 1.

## Decisions
- No decay (roads permanent once formed)
- 2 levels: Trodden (>=20 steps) and Road (>=100 steps)
- Grassland only (Forest not eligible; extensible via isRoadEligible)
- Player gets speed benefit; villager speed unchanged (deferred)
- A* admissibility: normalize MoveCost by roadMoveCooldown (90ms), not defaultMoveCooldown (150ms)
