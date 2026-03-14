# Verification — S4 House Building Footprint

## Primary gate

```bash
make check   # lint + tests — must pass after every stage
```

## Visual checks (run with `make run`)

After Stage 2:
- [ ] Three non-anchor house tiles show plain grass (no building sprite).
- [ ] NW anchor tile shows a roof sprite that overflows above the north footprint row.
- [ ] Foundation tiles (`?`) still render as per-tile dirt, not grass.

After Stage 3:
- [ ] House renders as a single coherent thatched cottage spanning the 2×2 footprint.
- [ ] Roof is visible above the north row (overflow, ~32px above footprint).
- [ ] Front wall face (half-timber, cream/brown) visible on south row.
- [ ] Wooden door centered on wall face.
- [ ] Flower-box windows flanking the door.
- [ ] No repeated-tile stamp artifact on any of the 4 footprint tiles.
- [ ] Building does not visually bleed into unrelated tiles.

## Regression checks

- [ ] Grassland tiles render correctly.
- [ ] Trees (sapling / young / mature / stump) unaffected.
- [ ] Roads and trodden paths unaffected.
- [ ] Log storage unaffected.
- [ ] Player and villager sprites unaffected.
- [ ] Foundation tiles unaffected.

## Env assumptions

- All sprite packs present under `assets/sprites/` (gitignored, must be downloaded manually).
- `assets/sprites/lpc-thatched-roof-cottage/thatched-roof.png` and `cottage.png` present.
- `assets/sprites/lpc-windows-doors-v2/windows-doors.png` present.
