# VERIFY: Road Formation

## Primary gate
```bash
make check   # lint + test (must pass after every stage)
```

## Stage 1 exit checks
```bash
make test    # TestRoadLevelFor, TestPlayerMove_IncrementsWalkCount, TestVillagerMove_IncrementsWalkCount pass
```
- `RoadLevelFor` returns 0 for WalkCount < 20, 1 for 20-99, 2 for >= 100
- `isRoadEligible` returns true for Grassland, false for Forest
- Player move on Grassland tile increments WalkCount by 1
- Player move on Forest tile does NOT increment WalkCount
- Villager move on Grassland tile increments WalkCount

## Stage 2 exit checks
```bash
make test    # TestMoveCooldownFor_RoadLevels, TestMoveCost_RoadLevels pass
```
- `MoveCooldownFor` returns 90ms for WalkCount >= 100 Grassland, 120ms for >= 20, 150ms for < 20
- `World.MoveCost` returns 1.0 for road, 1.33 for trodden, 1.67 for grassland, 3.33 for forest (approx)
- All MoveCost values >= 1.0 (A* admissibility preserved)

## Stage 3 exit checks
```bash
make build   # compiles without errors
make run     # game launches; walk around to verify road rendering
```
Manual:
- In TUI: walk on same grassland tile 20+ times → `:` glyph appears; 100+ times → `=`
- In Ebiten: walk on same tile → color shifts tan then gray-brown

## Stage 4 exit checks
```bash
make check   # all clean
gh pr view   # PR open with correct description
```
