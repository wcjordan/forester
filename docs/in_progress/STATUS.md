# STATUS

- [ ] Stage 1: Add `Game.Tick()` / `Game.RegrowTick()`; move `DepositTickInterval` to game
- [ ] Stage 2: Wire render to new API; remove game logic from `Model`

## Current state
Planning complete. Ready to implement.

## Key decisions
- `depositCooldown` moves from `render.Model` to `game.Game` (time-based, not tick-count)
- `DepositTickInterval` moves from render to game (it's a game-logic rate)
- `State.Move()` stays in render — it's input handling, not game loop orchestration
- Two-timer design (tickMsg / regrowTickMsg) stays in render — scheduling is a render concern
- Individual `State` methods (Harvest, AdvanceBuild, etc.) stay exported/unchanged
