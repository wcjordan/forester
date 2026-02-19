# Plan: Auto-cut + Wood Counter

## Goal
Trees auto-harvest when player is adjacent. Trees have variable sizes (4–10 wood). Stumps replace fully-harvested trees and look distinct.

## Stages

### Stage 1: Tile + Worldgen [ ]
- Add `Stump` to `TerrainType`
- Add `TreeSize int` to `Tile`
- Assign random TreeSize 4–10 to Forest tiles during worldgen
- Update worldgen tests

### Stage 2: Player Harvesting + State.Move [ ]
- Add `HarvestAdjacent(w *World)` to `Player` (1 wood/adjacent Forest/turn)
- Add `Move(dx, dy int)` to `State` (MovePlayer + HarvestAdjacent)
- Update renderer to call `State.Move()`
- Add tests for harvesting

### Stage 3: Rendering [ ]
- Size-based tree chars: `#` (7+), `t` (4–6), `,` (1–3)
- Stump glyph: `%`
- Colors: green trees, dark gray stumps
