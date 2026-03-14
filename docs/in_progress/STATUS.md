# Status — S4 House Building Footprint

## Stages

- [x] Stage 1 — Embed spritesheets & pre-slice frames
- [x] Stage 2 — NW-anchor detection + roof overlay
- [x] Stage 3 — Tune sprite crop coordinates to match reference image

## Current state

All stages complete. `make check` passes.
Awaiting visual review — run `make run` and build a house to inspect.
Further coordinate tuning may be needed based on visual result.

## Key decisions

- Building rendered as a single pre-composed 64×96 `*ebiten.Image` (roof 64×64 top,
  wall 64×32 bottom), drawn at NW anchor with `offsetY: -tileSize` (–32px overflow).
- Other 3 footprint tiles: grass base only, no building sprite.
- Foundation tiles: unchanged (per-tile dirt).
- Sources: `thatched-roof.png` (roof), `cottage.png` (wall), `windows-doors.png` (door + windows).
