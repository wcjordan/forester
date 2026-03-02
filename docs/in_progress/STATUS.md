# G0 Status

## Done
- Stage 1: Ebitengine dependency added, skeleton EbitenGame ✅
- Stage 2: EbitenGame Update/Draw/Layout fully implemented ✅
- Stage 3: main.go rewired; --tui flag; main_tui.go (!js) + main_wasm.go (js) ✅
- Stage 4: WASM build verified; `make wasm` target added ✅

## All exit criteria met
- `make build` passes
- `make check` (lint + tests) passes
- WASM build compiles: `GOOS=js GOARCH=wasm go build -o forester.wasm .`
- `./forester` opens Ebitengine window (1280×720, colored tiles, WASD movement)
- `./forester --tui` runs bubbletea TUI as before
