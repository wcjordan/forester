# Status: Player Sprite Animation

## Stages

- [x] Stage 1 — Game layer (LastHarvestAt, LastBuildAt)
- [x] Stage 2 — Asset loading + docs
- [x] Stage 3 — Walk animation
- [x] Stage 4 — Slash + Thrust animations
- [ ] Stage 5 — PR

## Current state

All implementation stages complete. Opening PR.

## Key decisions

- Villager animation deferred (static villagers remain using soldier_altcolor.png)
- Option A for harvest trigger: `LastHarvestAt time.Time` on `Player` (set in `game/resources/wood.go`)
- Option A for build trigger: `LastBuildAt time.Time` on `Player` (set in `game/structures/house.go`)
- Building animation = Thrust (rows 4-7), Chopping animation = Slash (rows 12-15)
- No deposit animation
