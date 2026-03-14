# Status — S4 House Building Footprint

## Stages

- [x] Stage 1 — Embed spritesheets & pre-slice frames
- [x] Stage 2 — NW-anchor detection + roof overlay
- [ ] Stage 3 — Tune sprite crop coordinates to match reference image

## Current state

Stages 1 + 2 complete and combined into one commit. `make check` passes.
Building renders at NW anchor with grass on other 3 tiles.
Sprite crop coordinates are approximate placeholders — Stage 3 tunes them visually.

## Key decisions

- Building rendered as a single pre-composed 64×96 `*ebiten.Image` (roof 64×64 top,
  wall 64×32 bottom), drawn at NW anchor with `offsetY: -tileSize` (–32px overflow).
- Other 3 footprint tiles: grass base only, no building sprite.
- Foundation tiles: unchanged (per-tile dirt).
- Sources: `thatched-roof.png` (roof), `cottage.png` (wall), `windows-doors.png` (door + windows).
