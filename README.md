# Forester

A terminal-based city builder where you play as a character who develops a village through organic, emergent gameplay. Roads form where you walk, structures appear where you work, and villagers learn by following your example.

See [docs/PROJECT_PLAN.md](docs/PROJECT_PLAN.md) for the full design document.

## Prerequisites

- **Go 1.23+** — [install](https://go.dev/dl/)
- **golangci-lint** — `brew install golangci-lint`
- **air** (optional, for hot-reload) — `go install github.com/air-verse/air@latest`

## Setup

### Sprite Packs (required)

All sprite packs live under `assets/sprites/` and are gitignored — download each one manually before building.

| Pack | Source | Local path |
|---|---|---|
| LPC Base Assets | [OpenGameArt](https://opengameart.org/content/liberated-pixel-cup-lpc-base-assets-sprites-map-tiles) | `assets/sprites/lpc_base_assets/` |
| LPC Trees | [OpenGameArt](https://opengameart.org/content/lpc-trees) | `assets/sprites/lpc-trees/` |
| LPC Terrains | [OpenGameArt](https://opengameart.org/content/lpc-terrains) | `assets/sprites/lpc-terrains/` |
| LPC Thatched Roof Cottage | [OpenGameArt](https://opengameart.org/content/lpc-thatched-roof-cottage) | `assets/sprites/lpc-thatched-roof-cottage/` |
| LPC Adobe Building Set | [OpenGameArt](https://opengameart.org/content/lpc-adobe-building-set) | `assets/sprites/lpc-adobe-building-set/` |
| LPC Containers | [OpenGameArt](https://opengameart.org/content/lpc-containers) | `assets/sprites/lpc-containers/` |
| LPC Windows & Doors | [OpenGameArt](https://opengameart.org/content/lpc-windows-doors) | `assets/sprites/lpc-windows-doors/` |

For each pack: download the archive from the link above, extract it, and place the contents at the local path shown. Attribution files (`CREDITS-*.txt`) are included in each archive. All packs use CC-BY or CC0 licenses.

Key files after extraction:
- `lpc_base_assets/`: `tiles/` dir (grass.png, dirt.png, house.png, …) and `sprites/people/` dir (soldier.png, …)
- `lpc-trees/`: `trees-green.png`
- `lpc-terrains/`: `terrain-v7.png`
- `lpc-thatched-roof-cottage/`, `lpc-adobe-building-set/`, `lpc-containers/`: building and container tile PNGs

### Player Spritesheet (required)

The player character uses a custom spritesheet generated from the [Universal LPC Character Generator](https://liberatedpixelcup.github.io/Universal-LPC-Spritesheet-Character-Generator/).

1. Open the generator URL (or re-generate with your preferred layers)
2. Export and save the PNG as:
   ```
   assets/sprites/player-spritesheet.png
   ```
3. Save Credits & Attribution to `assets/sprites/player-spritesheet.credits.txt`

This file is `.gitignore`d and must be present before `make build` / `make test` (it is embedded at compile time via `//go:embed`). The spritesheet uses the Universal LPC layout: rows 4–7 thrust (64×64), rows 8–11 walk (64×64), and the Slash128 section at y≈3488+ (128×128 frames, 6 per direction) for the axe-chop animation.

Currently prefer: https://liberatedpixelcup.github.io/Universal-LPC-Spritesheet-Character-Generator/#sex=male&expression=Neutral_light&shoes=Basic_Boots_brown&legs=Fur_Pants_fur_tan&apron=Apron_green&overalls=Suspenders_brown&clothes=Longsleeve_brown&bauldron=Bauldron_tan&beard=Winter_Beard_dark%20brown&head=Rabbit_light&shadow=Shadow_shadow&gloves=Gloves_brass&weapon=Smash_axe

## Quick Start

```bash
git clone <repo-url> && cd forester
make run      # build and run
make dev      # run with hot-reload (requires air)
```

## Development

```bash
make check    # lint + test (run before committing)
make test     # run tests with race detector
make lint     # run golangci-lint
make build    # compile binary
make clean    # remove build artifacts
```

## Project Structure

```
main.go              # Entry point (Ebitengine by default; --tui for terminal mode)
main_tui.go          # bubbletea startup (non-WASM builds only)
main_wasm.go         # WASM stubs (js builds only)
game/                # Core game logic (no I/O, no rendering)
  game.go            # Game orchestrator (Tick loop)
  state.go           # Serializable game state
  player.go          # Player entity, movement, harvesting
  world.go           # World grid, tile access, regrowth
  worldgen.go        # Procedural map generation (cellular automata)
  tile.go            # Tile, TerrainType, StructureType definitions
  structure.go       # StructureDef interface + registry
  storage.go         # StorageInstance, storageDef, resourceStorage
  storage_manager.go # StorageManager (deposit/withdraw aggregation)
  villager.go        # Villager, VillagerManager; autonomous chop/deliver behavior
  spawn.go           # Foundation spawning logic
  env.go             # Env (runtime context for structure methods)
  clock.go           # Clock interface + FakeClock for tests
  xp.go              # XP tracking, milestone thresholds, card offer selection
  story.go           # Ordered one-shot story beats
  core/              # StructureType leaf package (no upstream deps)
  geom/              # Pure geometry helpers (Point, FindPath A*, SpiralSearchDo)
  resources/         # woodDef (implements ResourceDef, registers via init())
  structures/        # logStorageDef, houseDef (register via init())
  upgrades/          # All upgrade cards (register via init())
render/
  ebiten_model.go    # Ebitengine renderer (sprites, camera, input)
  tui_model.go       # bubbletea Model (TUI mode, --tui flag)
  hud.go             # HUD/status bar drawing
  sprites.go         # Sprite loading and tile-to-sprite mapping
  util.go            # Shared render utilities
e2e_tests/           # End-to-end tests with injected clock + RNG
docs/
  PROJECT_PLAN.md    # Game design document
  GETTING_AROUND.md  # Codebase navigation guide
```

## Tech Stack

- **Go** — game logic and renderer
- **Ebitengine** — 2D graphical renderer (default; compiles to WASM for web)
- **bubbletea** — terminal UI (available via `--tui` flag)
- **lipgloss** — terminal color/style (TUI mode)
- **golangci-lint** — linting
- **air** — hot-reload during development
