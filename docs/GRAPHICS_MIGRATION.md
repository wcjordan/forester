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
  - `Update() error` — drive `game.Tick()` at `game.GameTickInterval` (currently 100ms, same as current bubbletea loop)
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
- Load spritesheets via `ebitenutil.NewImageFromFile` or embed via `go:embed`, decode with `image/png`, then wrap with `ebiten.NewImageFromImage`
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

---

## Appendix: Alternatives Considered

The options below were explored during planning. Each has real value — they were not dismissed, just deferred or deprioritized relative to the chosen path.

---

### Isometric Grid / View

**What it is**: A 2D rendering projection where the camera sits at roughly a 30° angle, giving a pseudo-3D look. Tiles become diamonds (2:1 aspect ratio). Used by Warcraft 2, Age of Empires, Banner Saga, and the classic Diablo games.

**Value**: Significantly more visual depth and richness than flat top-down. Characters and structures can have front/side faces. Trees read as trees rather than canopy blobs. It is the natural evolution of this game's visual style given the reference aesthetics.

**Why deferred (not rejected)**: Isometric requires two things flat top-down does not — a coordinate transform and correctly depth-sorted draw order (tiles farther from camera must be drawn first). Both are straightforward but add scope to the initial integration. The plan explicitly includes isometric as Phase G4. Starting flat top-down means the sprite pipeline and Ebitengine integration are proven before adding the projection complexity. The game grid and coordinate system need no changes — only the `Draw()` path is affected.

---

### Godot 4

**What it is**: A full-featured, open-source game engine with a built-in editor, scene system, 2D/3D rendering, animation tools, physics, and web/mobile export. Uses GDScript (Python-like) or C# as its primary languages.

**Value**:
- Best-in-class built-in isometric tilemap editor — you paint tiles visually rather than in code
- Animation state machines for characters (idle, walk, chop) with visual tooling
- First-class web (HTML5) and mobile export with better packaging than WASM DIY
- Large, active community with many 2D city-builder / strategy tutorials
- Free, no licensing risk
- AI-assisted asset import works well (drag PNG in, configure atlas, done)

**Why not chosen**: Requires porting all game logic from Go to GDScript or C#. The `game/` package is ~2000 lines of well-tested, evolving Go. At this stage of development, that port would be a large, risky undertaking that duplicates already-working logic and removes the ability to run `make check` as a verification gate. Godot becomes more attractive if and when the game design stabilizes and a more permanent technology commitment makes sense.

---

### Phaser.js (web-first)

**What it is**: A JavaScript/TypeScript 2D game framework that runs natively in the browser via HTML5 canvas/WebGL. Mature, widely used for browser games, with built-in tilemap support (Tiled map format), sprite animation, camera, and input handling.

**Value**:
- Runs in the browser with zero WASM overhead — native JS performance
- Excellent Tiled (.tmx) integration for designing maps visually
- Large tutorial ecosystem for exactly this type of top-down grid game
- Easy deployment — just static files, no Go runtime or WASM loader
- TypeScript gives reasonable type safety

**Why not chosen**: Requires maintaining a Go↔browser boundary. Two options exist: (a) compile game logic to WASM and call it from JS — workable but the interop layer (syscall/js, JSON marshalling) is tedious and fragile; (b) move game logic to TypeScript — a full rewrite. Neither is attractive while the Go game logic is actively evolving. Revisit if the game ever becomes a pure browser product with a stable, frozen backend.

---

### Midjourney

**What it is**: An AI image generation service accessed via Discord or web. Produces high-quality, stylistically coherent images from text prompts.

**Value**:
- Fastest way to generate concept art, mood boards, and reference sheets
- Pixel art mode (`--style raw`, `--ar 1:1`, pixel art prompting) produces usable 2D sprites
- Excellent for character concepts and building designs before committing to final sprites
- Consistency across a session can be improved with `--sref` (style reference) and `--cref` (character reference)
- No local GPU required

**Limitations / why it's not the only tool**: Output resolution is limited and upscaling degrades pixel art. Individual frames for animation require significant manual work to maintain consistency. Tilesets that need to tile seamlessly are hit-or-miss. Better used for reference and hero assets than for systematic tileset generation.

---

### Stable Diffusion (local / ComfyUI)

**What it is**: Open-source image generation model, runnable locally via ComfyUI or Automatic1111. Community LoRA models exist for pixel art, isometric tiles, and specific game aesthetics.

**Value**:
- Free, runs locally (no per-image cost)
- Isometric LoRAs (e.g., `isometric-diffusion`) reliably produce diamond-format tiles
- Tileable texture workflows produce seamless ground tiles
- Inpainting allows fixing problem areas of generated sprites
- Can fine-tune on a small set of reference images to maintain stylistic consistency across an asset set

**Limitations**: Requires a capable GPU (8GB+ VRAM) or cloud instance. Setup is more involved than Midjourney. Pixel art quality varies by model — `pixel-art-xl` and similar LoRAs are needed. Best suited for bulk tileset generation once an aesthetic direction is locked in.

---

### Adobe Firefly / DALL-E 3

**What it is**: Commercial AI image generation APIs with strong prompt-following and safety filtering.

**Value**:
- DALL-E 3 has excellent prompt adherence for specific requests ("a 32x32 pixel art grass tile, top-down view, bright green, simple")
- Firefly is integrated into Photoshop, making edit/refine cycles fast
- Both are accessible via API for batch asset generation workflows

**Limitations**: Less controllable for systematic tileset consistency than fine-tuned SD models. Firefly's pixel art quality is weaker than Midjourney. DALL-E 3 output cannot be used for commercial purposes in all jurisdictions without review. Better for one-off hero assets than bulk tileset work.

---

### Gemini 2.5 Flash Image ("Nano Banana")

**What it is**: Google's native multimodal image generation model (released August 2025, codename "Nano Banana"), available via Google AI Studio and the Gemini API. Unlike diffusion models bolted onto a language model, image generation runs natively within the same model as text reasoning.

**Value**:
- **Multi-turn iterative editing** — describe changes conversationally ("make the tree darker, add a shadow on the right") rather than re-prompting from scratch; particularly good for refining sprites to exact spec
- **Character consistency** — explicitly designed to maintain the same character's appearance across multiple generated images, which is a known weak point of Midjourney for animation frames
- **Image blending** — combine multiple reference images into a single output; useful for mixing LPC base sprites with a custom aesthetic
- **API-first** — clean REST API and Python SDK; straightforward to build batch asset generation pipelines
- **Low latency and low cost** — Flash-class model; fast iteration cycles without burning budget
- **World knowledge in the loop** — reasoning capabilities help it understand game-specific concepts ("isometric top-down tile, 2:1 diamond ratio") more reliably than pure diffusion models

**Comparison to Midjourney**:
- Midjourney produces higher raw aesthetic quality and has stronger community-tuned style prompts
- Gemini 2.5 Flash Image is more controllable for iteration and consistency — better for a systematic asset pipeline, worse for one-shot hero art
- For game sprites specifically, the character consistency and API access make it a strong complement: use Midjourney for initial style exploration, Gemini for generating consistent animation frames and tileset variants

**Limitations**: Pixel art output quality is less proven than Midjourney; community prompt recipes are less mature. Not ideal as a sole tool for final production assets yet, but the API-driven iteration loop is genuinely faster for systematic work.

---

### Summary Table

| Option | Best For | Key Trade-off | Status |
|---|---|---|---|
| **Ebitengine** | Go-native 2D, WASM web export | DIY isometric; smaller community than Unity/Godot | **Chosen** |
| **Isometric view** | Visual depth matching WC2/Banner Saga | Coordinate transform + draw-order complexity | Phase G4 |
| **Godot 4** | Polished isometric tooling, visual editor | Full Go→GDScript/C# port required | Revisit if design stabilizes |
| **Phaser.js** | Pure-browser, no WASM overhead | Go↔JS interop or full JS rewrite | Revisit for browser-only product |
| **Midjourney** | Concept art, character refs, hero assets | Animation consistency; tileset seams | Use now for concepts |
| **Stable Diffusion** | Bulk tilesets, isometric LoRAs, seamless textures | GPU required; setup overhead | Use for systematic tileset gen |
| **Gemini 2.5 Flash Image** | Iterative editing, character consistency, API pipelines | Pixel art quality less proven than MJ | Strong supplemental |
| **DALL-E 3 / Firefly** | One-off assets, prompt-precise sprites | Less consistent across asset sets | Supplemental |

---

### Recommended Split: Midjourney + Gemini 2.5 Flash Image

These two tools are complementary rather than competing, covering opposite ends of the asset production pipeline:

**Phase 1 — Style definition (Midjourney)**

Use Midjourney to establish the visual language before producing any final assets:
1. Generate 10–20 mood board images with prompts like `"pixel art village tileset, top-down view, warm palette, Warcraft 2 style, 32x32 tiles"`
2. Pick the 2–3 results that feel right and save them as style references (`--sref`)
3. Generate key hero assets — player character, a tree, a house — to lock in proportions and color palette
4. Export the approved references; these become the aesthetic contract for everything that follows

Midjourney's strength here is aesthetic judgment and creative range across many options quickly. Its weakness is reproducibility — getting the *same* character twice is hard.

**Phase 2 — Systematic asset production (Gemini 2.5 Flash Image)**

Use Gemini with the Midjourney references as input images to generate the full asset set:
1. Upload the approved hero sprite and ask for directional variants: `"same character, facing left, same pixel art style"`
2. Generate animation frames conversationally: `"now the same character mid-stride, left foot forward"`
3. Build tileset variants: `"same grass tile but with a dirt path crossing it horizontally"`
4. Iterate in-place: `"the tree canopy reads too dark against the grass — lighten it two shades"`

Gemini's multi-turn editing and character consistency are well-suited to this systematic work. The API access also makes it practical to script batch generation (e.g. all 4 directional frames × all terrain types in one run).

**In practice**: expect to spend ~20% of art time in Midjourney defining the style, ~70% in Gemini producing and refining the full asset catalog, and ~10% in Aseprite for final pixel cleanup and animation timing.
