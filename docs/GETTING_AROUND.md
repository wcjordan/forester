# Getting Around the Forester Codebase

A navigation guide covering file layout, responsibilities, and the libraries used in each area.

---

## Top-level files

| File | Purpose |
|---|---|
| `main.go` | Entry point. Defaults to Ebitengine window; `--tui` flag runs bubbletea TUI. Blank-imports `game/structures`, `game/upgrades`, and `game/resources` to trigger `init()` registration. |
| `main_tui.go` | (`//go:build !js`) `shouldRunTUI()` / `runTUI()` — bubbletea startup. Excluded from WASM builds. |
| `main_wasm.go` | (`//go:build js`) Stubs for `shouldRunTUI`/`runTUI` so WASM compiles without bubbletea. |
| `go.mod` / `go.sum` | Module `forester`, Go 1.24. Direct deps: `bubbletea`, `lipgloss`, `ebiten/v2`. |
| `Makefile` | All dev commands — see [Verification commands](#verification-commands). |
| `.air.toml` | Hot-reload config for `air` (`make dev`). |
| `.golangci.yml` | Linter rules consumed by `golangci-lint` (`make lint`). |

---

## Package map

### `game/` — all game logic (no I/O, no rendering)

**Core entities**

| File | Key types / funcs | Notes |
|---|---|---|
| `game.go` | `Game`, `New()`, `NewWithClock()`, `NewWithClockAndRNG()`, `Tick()`, `SelectCard()`, `XPInfo()` | Top-level orchestrator. Owns `State`, `StorageManager`, `VillagerManager`. `Tick()` runs one logical frame: harvest → story → spawn → adjacent-structure interactions → villagers → regrow. Returns early when a card offer is pending. |
| `state.go` | `State` | Serializable game state: `Player`, `World`, `FoundationDeposited`, `HouseOccupancy`, `XP`, `XPMilestoneIdx`, `pendingOfferIDs`, `completedBeats`. Designed to be a dumb data bag; orchestration lives elsewhere. |
| `player.go` | `Player`, `HarvestAdjacent()` | Harvests the three-tile forward arc. Carry capacity is dynamic via `Player.MaxCarry`. Move cooldowns are terrain-dependent. |
| `tile.go` | `Tile`, `TerrainType` (`Grassland`, `Forest`), `StructureType` (alias for `core.StructureType`) | Pure data. `StructureType` is a string type so external packages can define new values without editing this file. `Tiles[y][x]` indexing convention (row-major). |
| `world.go` | `World`, `NewWorld()`, `TileAt()`, `InBounds()`, `SetStructure()`, `AddStructure()`, `IsAdjacentToStructure()`, `Regrow()`, `IsBlocked()`, `isHarvestable()` | 2D grid. `Regrow()` is probabilistic (1-in-40 per eligible Forest tile). `AddStructure()` combines `SetStructure` + index update. |
| `worldgen.go` | `GenerateWorld()`, `defaultSeed` | Procedural terrain via cellular automata (5 iterations). Same seed → same map. Uses its own local `*rand.Rand`; does **not** share the game RNG. |
| `spawn.go` | `maybeSpawnFoundation()`, `findValidLocationNearPlayer()`, `findValidLocationNearSpawn()`, `isValidArea()` | Foundation spawn logic: checks each `StructureDef.ShouldSpawn()`, finds valid grassland area, places the foundation tile. Free functions with explicit params. |
| `story.go` | `StoryBeat`, `storyBeats`, `maybeAdvanceStory()` | Ordered one-shot triggers. Fire at most one per tick; retry until the action succeeds. Drives first-log-storage and first-house story beats. |
| `structure.go` | `StructureDef` interface, `StructureCallbacks`, `FinalizeFoundation`, `RegisterStructure()`, `IterateStructures()` | Extension point for new structures. Each type registers itself via `init()` in `game/structures/`. |
| `resource.go` | `ResourceDef` interface, `RegisterResource()`, `IterateResources()` | Extension point for harvestable resource types (wood, future: stone, etc.). Registered from `game/resources/`. |
| `storage.go` | `ResourceType`, `StorageDef`, `StorageInstance`, `ResourceStorage` | `StorageDef` extends `StructureDef` for structures that hold resources. `StorageInstance` tracks one structure's fill level. |
| `storage_manager.go` | `StorageManager`, `StorageState` | Runtime owner of all storage amounts. `Register()` called on `OnBuilt`; `DepositAt()` / `WithdrawFrom()` used by interaction handlers and villagers. |
| `villager.go` | `Villager`, `VillagerManager`, `VillagerTask`, `RegisterVillagerDepositType()`, `RegisterVillagerDeliveryType()` | Autonomous agents. `VillagerManager.Tick()` steps each villager: probabilistic task selection → move toward target → harvest/deposit. Multi-tree chop: villager fills carry cap before returning to storage. |
| `upgrade.go` | `UpgradeDef` interface, `upgradeRegistry`, `RegisterUpgrade()` | Upgrade extension point. Each upgrade registers itself via `init()` in `game/upgrades/`. `Apply(*Env)` mutates state. |
| `xp.go` | `AwardXP()`, `xpMilestoneAt()`, `pickCardOffer()` | XP tracking and milestone logic. `AwardXP` adds XP and enqueues 3-card offers for each milestone crossed. Milestone gaps grow: 50, 75, 100, 125, … |
| `env.go` | `Env` | Runtime context (`State`, `Stores`, `Villagers`, `RNG`) passed to all `StructureDef`, `ResourceDef`, and `UpgradeDef` methods. Separates serializable state from derived runtime state. |
| `clock.go` | `Clock` interface, `RealClock{}`, `FakeClock` + `NewFakeClock()` + `Advance()` | Inject `*FakeClock` in tests for deterministic time control. Starts at 2024-01-01 so zero-value cooldowns are always expired. |
| `input.go` | `MoveMsg` | Thin message type bridging bubbletea keys → game moves. |

**Subpackages**

| Package | Key types / funcs | Notes |
|---|---|---|
| `game/core` | `StructureType`, `NoStructure` | Leaf package (no imports from `game/`). Lets `game/structures` and `game/upgrades` define `StructureType` constants without creating import cycles. |
| `game/geom` | `Point`, `FindPath()`, `SpiralSearchDo()`, `FootprintBorderDo()` | Pure geometry helpers. No game logic. Used by villagers, spawn, and structure placement. |
| `game/resources` | `woodDef{}` | Implements `ResourceDef` for wood. Registers via `init()`. Handles harvesting (`Harvest`) and regrowth (`Regrow`). |
| `game/structures` | `logStorageDef{}`, `houseDef{}` | Implements `StructureDef` for each structure. Registers via `init()`. Also calls `RegisterVillagerDepositType` / `RegisterVillagerDeliveryType`. |
| `game/upgrades` | `carryUpgrade`, `buildSpeedUpgrade`, `depositSpeedUpgrade`, `harvestSpeedUpgrade`, `moveSpeedUpgrade`, `spawnVillagerUpgrade` | Implements `UpgradeDef` for each upgrade card. Registers via `init()`. |
| `game/internal/gametest` | `WallDef`, `NewGame()` | Test helpers. `WallDef` is a minimal `StructureDef` stub. Only importable by packages inside `game/`. |

---

### `render/` — presentation layer

Two renderers share the `render` package. `game/` has no I/O knowledge; both renderers are read-only views over `*game.Game`.

| File | Key types / funcs | Notes |
|---|---|---|
| `tui_model.go` | `Model` (implements `tea.Model`), `NewModel()`, `NewModelWithClock()`, `Init()`, `Update()`, `View()` | (`//go:build !js`) Bubbletea Elm-Architecture model. `Update` dispatches `tea.KeyMsg` / `TickMsg` / `tea.WindowSizeMsg`. `View` renders the viewport centered on the player plus a status bar. When a card offer is pending, renders a full-screen card selection overlay. |
| `ebiten_model.go` | `EbitenGame` (implements `ebiten.Game`), `NewEbitenGame()`, `Update()`, `Draw()`, `Layout()` | Ebitengine renderer. `Update` handles WASD/arrow movement and game ticks at 100ms intervals. `Draw` renders the world as solid-color rectangles (1280×720, 32px tiles) with a spec-defined color palette. |
| `util.go` | `clamp()` | Shared integer clamp helper used by both renderers. |

**Rendering glyphs** (TUI — defined in `tui_model.go`)

| Glyph | Meaning |
|---|---|
| `@` | Player (blue) |
| `#` | Dense forest, TreeSize ≥ 7 (green) |
| `t` | Mid-size tree, TreeSize 4–6 (green) |
| `,` | Sapling, TreeSize 1–3 (green) |
| `%` | Cut tree / stump, TreeSize 0 (dark gray) |
| `?` | Foundation footprint — deposit wood while adjacent to build (yellow) |
| `L` | Built Log Storage (bold yellow) |
| `H` | Built House (bold magenta) |
| `v` | Villager (cyan) |
| `.` | Grassland |

---

### `e2e_tests/` — end-to-end tests

| File | Purpose |
|---|---|
| `log_storage_test.go` | `TestLogStorageWorkflow` — full scenario: navigate → harvest → trigger foundation → build log storage → deposit. |
| `house_test.go` | `TestHouseWorkflow` — builds log storage, then first/second house; verifies villager spawning via XP card path. |
| `xp_test.go` | `TestXPMilestones` — verifies XP accumulates correctly, milestones fire, and card offers are queued. |
| `helpers_test.go` | Shared helpers: `driveToXPOffer()`, card selection, house build helpers. |
| `visual_test.go` | `renderFrame` / `announcePhase` for `E2E_VISUAL=1` playback mode. No-ops in CI. |

Run visually: `make e2e_viz` (set `E2E_VISUAL=1 E2E_VISUAL_DELAY=150ms`).

---

## Key libraries

| Library | Used in | Purpose |
|---|---|---|
| `github.com/charmbracelet/bubbletea` | `main_tui.go`, `render/tui_model.go`, `e2e_tests/` | Elm-Architecture TUI framework. Drives the event loop (`Init` / `Update` / `View`). Excluded from WASM builds. |
| `github.com/charmbracelet/lipgloss` | `render/tui_model.go` | Terminal color and style (ANSI). Used for per-tile glyph styles. Excluded from WASM builds. |
| `github.com/hajimehoshi/ebiten/v2` | `main.go`, `render/ebiten_model.go` | 2D game engine. Provides the window, frame loop, input, and drawing primitives. Compiles for native and WASM. |
| `math/rand` | `game/` | Seeded RNG for worldgen and regrowth. Injected via `*rand.Rand` for test determinism. |
| Standard library (`time`, `math`, `fmt`, `strings`) | Throughout | No other runtime dependencies. |

---

## Architectural patterns to know

### Clock injection
`Clock` interface → `RealClock{}` in production, `*FakeClock` in tests.
All time-dependent game logic (`Tick`, move cooldowns, deposit cooldown) accepts a `Clock`.

### RNG injection
`game.NewWithClockAndRNG(clock, rng)` gives tests full determinism.
`worldgen.go` uses its own local RNG (seeded separately by `defaultSeed = 42`).

### Structure registry
Each structure file calls `RegisterStructure(myDef{}, game.StructureCallbacks{ ... })` in `init()` inside `game/structures/`.
`main.go` and `e2e_tests/` blank-import `game/structures` to trigger registration.
**To add a new structure:** implement `StructureDef`, add a `StructureType` constant in `game/core/core.go`, create a new file in `game/structures/`, and in `init()` call `RegisterStructure(myDef{}, game.StructureCallbacks{ ... })`.

### Upgrade registry
Each upgrade file calls `RegisterUpgrade(myUpgrade{})` in `init()` inside `game/upgrades/`. The upgrade ID comes from `myUpgrade.ID()`.
`main.go` and `e2e_tests/` blank-import `game/upgrades` to trigger registration.

### XP and card offers
`AwardXP(env, n)` in `game/xp.go` adds XP and enqueues a 3-card offer for each milestone crossed.
`Game.HasPendingOffer()` causes `Tick()` to return early (game pauses). `Game.SelectCard(idx)` applies the chosen upgrade and pops the offer.

### Tile coordinate convention
`World.Tiles[y][x]` — row first, then column.
Always use `World.TileAt(x, y)` to access tiles safely (returns `nil` for out-of-bounds).

### Dual renderer
Two renderers exist side by side in the `render/` package:
- **TUI** (`tui_model.go`, `//go:build !js`): bubbletea drives a `TickMsg` every `GameTickInterval` (100ms) which calls `game.Tick()`. Player input via `tea.KeyMsg`.
- **Ebitengine** (`ebiten_model.go`): Ebitengine calls `Update()` ~60/s. `Update` accumulates time and calls `game.Tick()` every 100ms. WASD held keys call `player.Move()`; the player's 150ms move cooldown throttles actual movement.

Select renderer at runtime: `./forester` → Ebitengine window; `./forester --tui` → bubbletea TUI.

### WASM build isolation
`render/tui_model.go` and `main_tui.go` carry `//go:build !js` so bubbletea/lipgloss are excluded from WASM builds. `main_wasm.go` (`//go:build js`) provides stubs. Build with `make wasm`.

---

## Verification commands

```bash
make check    # lint + test (primary gate)
make test     # go test -race ./...
make lint     # golangci-lint run ./...
make build    # compile binary
make run      # build and run
make dev      # hot-reload with air
make e2e_viz  # visual E2E playback in terminal
make clean    # remove build artifacts
make format   # format code w/ gofmt
make wasm     # compile WASM binary (forester.wasm)
```
