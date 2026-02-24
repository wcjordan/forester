# Plan: Foundation Placement Constraints

## Goal
Tighten the rules for where foundations may spawn:
1. A foundation cannot overlap the player's current tile.
2. A foundation must have a full 1-tile gap (all 8 directions) from every existing structure.
3. Houses are placed as close as possible to the world spawn point (50,50) — the player's
   current position plays no role in house placement.

Log storage keeps its existing player-toward-center path algorithm.

## Non-goals
- Do not change log storage placement algorithm.
- Do not add new structures.
- Do not change build mechanics (deposit, completion, upgrade cards).

## Single stage — one commit

### Steps

1. **`game/progression.go` — `isValidArea`**
   - Add `playerX, playerY int` parameters.
   - Return false if `(playerX, playerY)` is within the w×h footprint.
   - Return false if any tile in the 1-tile Chebyshev border around the footprint
     (`x-1..x+w`, `y-1..y+h` minus the footprint itself) has `Structure != NoStructure`.
     Out-of-bounds border tiles are skipped (no structure there by definition).

2. **`game/progression.go` — `findValidLocation`**
   - Pass `env.Player.X, env.Player.Y` down to `isValidArea` calls.

3. **`game/progression.go` — `findValidLocationNearSpawn` (new function)**
   - Enumerate all in-bounds top-left positions for a w×h footprint.
   - Sort by Euclidean distance² from world spawn center to footprint center (x+w/2, y+h/2).
   - Return first position for which `isValidArea` passes.

4. **`game/progression.go` — `maybeSpawnFoundation`**
   - Check if `def` implements a new optional interface `SpawnAnchoredPlacer`.
   - If yes: use `findValidLocationNearSpawn`; otherwise use `findValidLocation`.

5. **`game/house.go` — `houseDef`**
   - Implement `UseSpawnAnchoredPlacement() bool { return true }` to satisfy
     `SpawnAnchoredPlacer`.

6. **`e2e_tests/house_test.go`**
   - Update Phase 6 navigation and Phase 7/8 position assertions to match the new
     spawn location (determined empirically after implementation, likely ~(49,51)).
   - Remove the comment/expectation that the house spawns at the player's position.

7. **Verify**: `make check` passes.

8. **Commit**: "Feat: placement constraints — no-under-player, 1-tile buffer, house near spawn"

## Exit criteria
- `make check` passes (lint + tests, including race detector).
- Log storage e2e test: foundation still at (48,46)–(51,49), no behaviour change.
- House e2e test: house spawns near (50,50), not at player's position; test passes.
