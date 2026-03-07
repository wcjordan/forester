# Verification — S3 Road Autotiling

## Commands
```
make check   # lint + test (primary gate)
make build   # ensure binary compiles
```

## Success criteria
- `make check` passes with no errors
- `make build` produces a binary
- `assets.TerrainSheet` is non-nil at runtime (panics loudly if embed fails)
- `soilAutotile` and `gravelAutotile` arrays have 16 non-nil entries each (verified by init())
- No references to removed `assets.TroddenPath` or `assets.Road`

## What success looks like
- Trodden path tiles render with Soil (terrain 14) lpc-terrains texture
- Road tiles render with Gravel_1 (terrain 18) lpc-terrains texture
- Edge tiles (1–3 road neighbors) show partial road / grass blend texture
- Fully surrounded road tiles show solid gravel/soil fill
