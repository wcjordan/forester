# Status: Player Sprite Animation

## Stages

- [ ] Stage 1 — Game layer (LastHarvestAt, LastBuildAt)
- [ ] Stage 2 — Asset loading + docs
- [ ] Stage 3 — Walk animation
- [ ] Stage 4 — Slash + Thrust animations
- [ ] Stage 5 — PR

## Current state

Planning complete. Ready to implement Stage 1.

## Key decisions

- Villager animation deferred (static villagers remain using soldier_altcolor.png)
- Option A for harvest trigger: `LastHarvestAt time.Time` on `Player` (set in `game/resources/wood.go`)
- Option A for build trigger: `LastBuildAt time.Time` on `Player` (set in `game/structures/house.go`)
- Building animation = Thrust (rows 4-7), Chopping animation = Slash (rows 12-15)
- No deposit animation
