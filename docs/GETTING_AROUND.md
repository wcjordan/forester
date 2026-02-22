# Getting Around the Forester Codebase

A navigation guide covering file layout, responsibilities, and the libraries used in each area.

---

## Top-level files

| File | Purpose |
|---|---|
| `main.go` | Entry point. Creates `game.Game` and `render.Model`, then hands control to bubbletea. |
| `go.mod` / `go.sum` | Module `forester`, Go 1.24. Direct deps: `bubbletea`, `lipgloss`. |
| `Makefile` | All dev commands — see [Verification commands](#verification-commands). |
| `.air.toml` | Hot-reload config for `air` (`make dev`). |
| `.golangci.yml` | Linter rules consumed by `golangci-lint` (`make lint`). |

---

## Package map

### `game/` — all game logic (no I/O, no rendering)

**Core entities**

| File | Key types / funcs | Notes |
|---|---|---|
| `game.go` | `Game`, `New()`, `NewWithClock()`, `NewWithClockAndRNG()`, `Tick()` | Top-level orchestrator. `Tick()` runs one logical frame: harvest → advance build → deposit → maybe regrow. |
| `state.go` | `State` (owns `Player` + `World`), `Move()`, `Harvest()`, `AdvanceBuild()` | Coordinates all subsystems. Ghost spawning and build completion live here. |
| `player.go` | `Player`, `MovePlayer()`, `HarvestAdjacent()`, `TryDeposit()` | Harvests the three-tile forward arc. Carries wood up to `MaxWood=20`. Move cooldowns are terrain-dependent. |
| `tile.go` | `Tile`, `TerrainType` (`Grassland`, `Forest`), `StructureType` (`NoStructure`, `GhostLogStorage`, `LogStorage`) | Pure data. `Tiles[y][x]` indexing convention (row-major). |
| `world.go` | `World`, `NewWorld()`, `TileAt()`, `InBounds()`, `SetStructure()`, `IsAdjacentToStructure()`, `Regrow()` | 2D grid. `Regrow()` is probabilistic (1-in-40 per eligible Forest tile). |
| `worldgen.go` | `GenerateWorld()`, `DefaultSeed` | Procedural terrain via cellular automata (5 iterations). Same seed → same map. Uses its own local `*rand.Rand`; does **not** share the game RNG. |

**Structures subsystem** (extensible via interface + registry)

| File | Key types / funcs | Notes |
|---|---|---|
| `structure.go` | `StructureDef` interface, `structures []StructureDef`, `BuildOperation` | `StructureDef` is the extension point for new structures. Each type registers itself via `init()`. |
| `log_storage.go` | `logStorageDef{}` | Implements `StructureDef`. Spawns at 10 wood cut; 4×4 footprint; 30-tick build; deposits 1 wood/tick when player is adjacent. Registers itself with `init()`. |

**Resources**

| File | Key types / funcs | Notes |
|---|---|---|
| `storage.go` | `ResourceType`, `StorageInstance`, `ResourceStorage` | Aggregates multiple storage instances per resource. `Deposit()` fills the first non-full instance. |

**Testability infrastructure**

| File | Key types / funcs | Notes |
|---|---|---|
| `clock.go` | `Clock` interface, `RealClock{}`, `FakeClock` + `NewFakeClock()` + `Advance()` | Inject `*FakeClock` in tests for deterministic time control. Starts at 2024-01-01 so zero-value cooldowns are always expired. |
| `input.go` | `MoveMsg` | Thin message type bridging bubbletea keys → game moves. |

---

### `render/` — TUI presentation layer

| File | Key types / funcs | Notes |
|---|---|---|
| `model.go` | `Model` (implements `tea.Model`), `NewModel()`, `NewModelWithClock()`, `Init()`, `Update()`, `View()` | Bubbletea Elm-Architecture model. `Update` dispatches `tea.KeyMsg` / `TickMsg` / `tea.WindowSizeMsg`. `View` renders the viewport centered on the player plus a status bar. |

**Rendering glyphs** (defined in `model.go`)

| Glyph | Meaning |
|---|---|
| `@` | Player (blue) |
| `#` | Dense forest, TreeSize ≥ 7 (green) |
| `t` | Mid-size tree, TreeSize 4–6 (green) |
| `,` | Sapling, TreeSize 1–3 (green) |
| `%` | Cut tree / stump, TreeSize 0 (dark gray) |
| `?` | Ghost structure footprint (yellow) |
| `L` | Built Log Storage (bold yellow) |
| `.` | Grassland |

---

### `e2e_tests/` — end-to-end tests

| File | Purpose |
|---|---|
| `log_storage_test.go` | `TestLogStorageWorkflow` — full scenario: navigate → harvest → trigger ghost → walk onto ghost → build → deposit. Uses injected `FakeClock` and seeded RNG for determinism. |
| `visual_test.go` | Helpers `renderFrame` / `announcePhase` for `E2E_VISUAL=1` playback mode. No-ops in CI. |

Run visually: `make e2e_viz` (set `E2E_VISUAL=1 E2E_VISUAL_DELAY=150ms`).

---

## Key libraries

| Library | Used in | Purpose |
|---|---|---|
| `github.com/charmbracelet/bubbletea` | `main.go`, `render/model.go`, `e2e_tests/` | Elm-Architecture TUI framework. Drives the event loop (`Init` / `Update` / `View`). |
| `github.com/charmbracelet/lipgloss` | `render/model.go` | Terminal color and style (ANSI). Used for per-tile glyph styles. |
| `math/rand` | `game/` | Seeded RNG for worldgen and regrowth. Injected via `*rand.Rand` for test determinism. |
| Standard library (`time`, `math`, `fmt`, `strings`) | Throughout | No other runtime dependencies. |

---

## Architectural patterns to know

### Clock injection
`Clock` interface → `RealClock{}` in production, `*FakeClock` in tests.
All time-dependent game logic (`Tick`, move cooldowns, deposit cooldown) accepts a `Clock`.

### RNG injection
`game.NewWithClockAndRNG(clock, rng)` gives tests full determinism.
`worldgen.go` uses its own local RNG (seeded separately by `DefaultSeed = 42`).

### Structure registry
Each structure file calls `structures = append(structures, myDef{})` in `init()`.
`state.go` iterates `structures` to check spawn conditions and dispatch `OnAdjacentTick` / `OnBuilt`.
**To add a new structure:** implement `StructureDef`, add constants to `tile.go`, create a new file, register via `init()`.

### Tile coordinate convention
`World.Tiles[y][x]` — row first, then column.
Always use `World.TileAt(x, y)` to access tiles safely (returns `nil` for out-of-bounds).

### Bubbletea tick loop
`render.TickMsg` fires every `HarvestTickInterval` (100 ms). Each tick calls `game.Tick()`.
Player input goes through `tea.KeyMsg` → `state.Move()`.

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
make format  # format code w/ gofmt
```
