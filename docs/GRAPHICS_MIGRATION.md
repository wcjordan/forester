# Graphics Migration Plan

## Vision

Move Forester from a terminal UI to a graphical renderer suitable for web and mobile distribution, without breaking or rewriting the game logic.

Target aesthetic: readable, colorful, 2D top-down — in the ballpark of **Oxygen Not Included** and **Warcraft 2** (rich tile palette, distinct unit sprites, clear grid). Isometric perspective is a stretch goal once the sprite pipeline is established.

Reference games: Warcraft 2, Oxygen Not Included, Cult of the Lamb, Banner Saga

---

## Key Architectural Advantage

The game logic (`game/`) and rendering (`render/model.go`) are already fully decoupled:

- `game.Game` owns all state and runs `Tick()` — no rendering knowledge
- `render.Model` is a stateless, read-only view of `game.State`
- All input handling lives in the render layer

This means the migration is **a renderer swap, not a rewrite**. The `game/` package stays untouched.

---

## Framework Decision: Ebitengine

**Selected: [Ebitengine](https://ebitengine.org/)** — a 2D game engine written in Go.

### Why Ebitengine

| Criterion | Assessment |
|---|---|
| Language | Pure Go — keep the entire `game/` package as-is |
| Web export | Compiles to WASM; runs in browser via `<canvas>` |
| Mobile | iOS/Android via gomobile (later phase) |
| 2D rendering | Sprites, tilemaps, draw ops — all first-class |
| Isometric | DIY but well-documented in the community |
| Maturity | Active, ships real games, stable API |
| AI art fit | Load PNG spritesheets directly; no special tooling |

### Why Not the Alternatives

- **Godot**: Best isometric tooling but requires porting all game logic to GDScript/C#. Not worth it while the game is still evolving rapidly.
- **Unity**: Overkill for a 2D grid game; C# port required; large WebGL bundles.
- **Phaser/Pixi.js + WASM**: Go↔JS interop via WASM is clunky; adds a JS build pipeline.
- **Full 3D**: The reference games are all 2D (even their "isometric" views are pre-rendered sprites). AI 3D model quality still lags significantly behind 2D generation. Overkill for tile-based movement.

---

## Art Style Decisions

### Phase 1: Flat top-down (immediate)
Match **Oxygen Not Included** — square tiles, clear grid, simple colored sprites. This is the fastest path to something visually coherent.

### Phase 2: Isometric sprites (later)
Match **Warcraft 2 / Banner Saga** — diamond-shaped tiles, sprites rendered at ~30° angle. Richer visual depth; compatible with the same Ebitengine renderer via a coordinate transform.

### Viewport
Tiles at **32×32 px** for phase 1. Isometric tiles are conventionally **64×32 px** (2:1 ratio). Picking 32px now means isometric tiles slot in cleanly as a later upgrade.

---

## AI Art Tooling

No need to hand-draw assets. Recommended pipeline:

| Tool | Use |
|---|---|
| **LPC (Liberated Pixel Cup)** | Starting point — free CC-licensed 32px tilesets for grass, forest, buildings. Already matches target aesthetic. [OpenGameArt LPC collection](https://opengameart.org/content/liberated-pixel-cup-lpc-base-assets-sprites-map-tiles) |
| **Midjourney** | Character concepts, building references, style exploration |
| **Stable Diffusion + isometric LoRA** | Generate tileset variations once isometric view begins |
| **Aseprite** | Pixel art cleanup and animation editing (inexpensive, well worth it) |

Suggested first assets to source/generate:
- Grass tile (single 32×32 sprite, tileable)
- Forest tile (dense canopy, mid-canopy, sapling, stump variants)
- Player character (idle + 4-directional walk, 32×32)
- Villager (simpler variant of player)
- Log storage (4×4 tile footprint = 128×128 or drawn as a single 64×64+ sprite)
- House (2×2 tile footprint)
- Foundation overlay (transparent highlight)

---

## Migration Phases

### Phase G0 — Ebitengine skeleton (colored rectangles, no sprites)

**Goal**: Wire Ebitengine into the project and render the world grid using solid-color rectangles. Prove the integration works and the game loop runs correctly.

**What changes**:
- Add `ebitengine` dependency (`go get github.com/hajimehoshi/ebiten/v2`)
- Create `render/ebiten_model.go` implementing `ebiten.Game`:
  - `Update() error` — drive `game.Tick()` at 100ms intervals (same as current bubbletea loop)
  - `Draw(*ebiten.Image)` — draw one colored rectangle per visible tile
  - `Layout(w, h int) (int, int)` — return logical resolution (e.g. 1280×720)
- Update `main.go` to use `ebiten.RunGame()` instead of bubbletea
- Keep `render/model.go` (bubbletea) but wire it behind a `--tui` flag so TUI mode still works during development

**Color palette for rectangles** (matching current TUI glyphs):
```
Grassland      → #7EC850 (mid green)
Forest/dense   → #2D6A2D (dark green)
Forest/mid     → #4A8A4A
Forest/sapling → #6DAA6D
Stump          → #6B5A3E (brown)
Foundation     → #D4A840 (yellow)
Log Storage    → #C8920A (amber)
House          → #A040C0 (purple)
Player         → #4080FF (blue)
Villager       → #40C0C0 (cyan)
```

**Exit criteria**:
- Game window opens at 1280×720
- World grid renders as colored tiles
- Player moves with WASD
- Game tick runs at correct interval
- `make check` passes
- WASM build compiles: `GOOS=js GOARCH=wasm go build -o forester.wasm .`

**Does not include**: sprites, animations, camera smoothing, UI/HUD

---

### Phase G1 — Sprite rendering (first real art)

**Goal**: Replace colored rectangles with 32×32 PNG sprites for terrain and structures.

**What changes**:
- Add `assets/sprites/` directory with PNG files (sourced from LPC or generated)
- Load spritesheets via `ebiten.NewImageFromFile` or embedded via `go:embed`
- Map `TerrainType` + `TreeSize` + `StructureType` → sprite frame
- Player and villager rendered as sprites (static for now)

**Exit criteria**:
- All terrain types and structure types have distinct sprites
- Player and villager visible as sprites
- No colored rectangles remain
- `make check` passes, WASM still compiles

---

### Phase G2 — HUD and camera polish

**Goal**: Basic in-game UI and smooth camera.

**What changes**:
- Status bar (wood count, storage level, villager count) drawn as text overlay using Ebitengine's text package
- Card selection screen rendered graphically (not terminal box-drawing)
- Camera interpolates smoothly toward player position (lerp)
- Viewport scales with window resize

**Exit criteria**:
- All game information from current TUI status bar is visible
- Card selection is playable
- Resize/fullscreen works without visual glitches

---

### Phase G3 — Web deployment

**Goal**: Playable in a browser, shareable via URL.

**What changes**:
- `wasm_exec.js` + minimal HTML shell page in `web/`
- Makefile target: `make web` → builds WASM and copies to `web/`
- Test in Chrome and Firefox

**Exit criteria**:
- Game runs in browser without Go installed
- Input (keyboard) works correctly in browser context
- Performance acceptable on a mid-range laptop

---

### Phase G4 — Isometric view (stretch goal)

**Goal**: Re-render the world with isometric perspective.

**Approach**:
- Coordinate transform: `screenX = (x - y) * tileW/2`, `screenY = (x + y) * tileH/2`
- Swap 32×32 square sprites for 64×32 isometric tile sprites
- Entities (player, villagers, trees) get isometric-angle sprites
- Render order: sort tiles by `y` then `x` to get correct depth layering

**This does not touch game logic** — only the Draw() path changes.

---

## What Stays Unchanged

- All of `game/` — zero changes required
- All of `e2e_tests/` — game logic tests remain valid
- `make check`, `make test`, `make lint` — same verification commands

---

## Immediate Next Step

**Start Phase G0.**

1. Add Ebitengine: `go get github.com/hajimehoshi/ebiten/v2`
2. Create `render/ebiten_model.go` with colored-rectangle rendering
3. Add `--tui` flag to `main.go` to preserve bubbletea mode during transition
4. Verify WASM compilation
5. Commit

This is a self-contained change (~200 lines) with no artwork dependency and a clear pass/fail outcome.
