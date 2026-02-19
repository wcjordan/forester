# Verification

## Commands
```bash
make check   # lint + tests (primary gate)
make run     # manual verification
```

## Manual checks (make run)
- Walk adjacent to trees: Wood counter increments
- Stay adjacent over multiple turns: tree chars progress # → t → , → %
- Stump (%) persists, does not disappear or revert
- Grassland (.) tiles not affected
- Status bar shows updated Wood count

## Test cases (automated)
- All Forest tiles have TreeSize in [4, 10] after GenerateWorld
- HarvestAdjacent: Wood increases, TreeSize decreases
- HarvestAdjacent: tile becomes Stump when TreeSize hits 0
- HarvestAdjacent: no effect on Stump tiles
- HarvestAdjacent: no effect on Grassland tiles
- HarvestAdjacent: safe at world edges (nil tile)
