# Verification: Player Sprite Animation

## Primary gate

```bash
make check   # lint + test — must pass after every stage
```

## Per-stage checks

### Stage 1 (game layer)
```bash
make test
```
- `player.LastHarvestAt` is non-zero after a harvest fires with wood available
- `player.LastBuildAt` is non-zero after a build deposit fires
- Both remain zero-value if no harvest/build occurred

### Stage 2 (asset loading)
```bash
make build   # confirms PlayerSheet embeds and compiles
make run     # visually: game launches, player renders (any frame)
```

### Stage 3 (walk animation)
```bash
make run
```
Visual checks:
- Hold W/A/S/D or arrow keys: player cycles through 8 directional frames
- Release all keys: player shows frame 0 of the current facing direction
- Direction matches key: W=up, S=down, A=left, D=right

### Stage 4 (slash + thrust)
```bash
make run
```
Visual checks:
- Stand facing a forest tile with TreeSize > 0: slash animation plays continuously
- Stand adjacent to a house foundation with wood in inventory: thrust animation plays continuously
- Walk to an area with no trees and no foundations: walk animation resumes

## Environment assumptions

- `assets/sprites/player-spritesheet.png` is present (generated from the LPC Character Generator URL in README)
- `assets/sprites/lpc_base_assets/` is present (existing requirement)
- Go 1.23+, golangci-lint installed

## Success / failure signals

- Success: `make check` exits 0, visual animations match expected behavior above
- Failure: compile error on missing spritesheet → add the file; test failure → fix before next stage
