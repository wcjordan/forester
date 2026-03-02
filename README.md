# Forester

A terminal-based city builder where you play as a character who develops a village through organic, emergent gameplay. Roads form where you walk, structures appear where you work, and villagers learn by following your example.

See [docs/PROJECT_PLAN.md](docs/PROJECT_PLAN.md) for the full design document.

## Prerequisites

- **Go 1.23+** — [install](https://go.dev/dl/)
- **golangci-lint** — `brew install golangci-lint`
- **air** (optional, for hot-reload) — `go install github.com/air-verse/air@latest`

## Setup

### LPC Base Assets (required)

The Ebitengine renderer uses sprites from the [Liberated Pixel Cup base assets](https://opengameart.org/content/liberated-pixel-cup-lpc-base-assets-sprites-map-tiles).

1. Download the archive from the link above
2. Extract it so that the `tiles/` and `sprites/` directories land at:
   ```
   assets/sprites/lpc_base_assets/
   ```
   Expected layout:
   ```
   assets/sprites/lpc_base_assets/
     tiles/
       treetop.png, trunk.png, grass.png, dirt.png, house.png, barrel.png, ...
     sprites/
       people/
         soldier.png, soldier_altcolor.png, ...
   ```

This directory is `.gitignore`d and must be populated manually before running.

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
main.go              # Entry point
game/                # Core game logic (no I/O, no rendering)
  game.go            # Game orchestrator (Tick loop)
  state.go           # Game state (player + world)
  player.go          # Player entity, movement, harvesting
  world.go           # World grid, tile access, regrowth
  worldgen.go        # Procedural map generation
  tile.go            # Tile, TerrainType, StructureType definitions
  structure.go       # StructureDef interface + registry
  log_storage.go     # Log Storage implementation
  storage.go         # ResourceType, StorageInstance, StorageDef
  storage_manager.go # StorageManager (runtime storage state)
  progression.go     # Foundation spawning logic
  env.go             # Env (runtime context for structure methods)
  clock.go           # Clock interface + FakeClock for tests
  input.go           # MoveMsg (key → game bridge)
render/
  model.go           # bubbletea Model (TUI presentation)
e2e_tests/           # End-to-end tests with injected clock + RNG
docs/
  PROJECT_PLAN.md    # Game design document
  GETTING_AROUND.md  # Codebase navigation guide
```

## Tech Stack

- **Go** — game logic
- **bubbletea** — terminal UI (Elm-architecture event loop)
- **lipgloss** — terminal color/style
- **golangci-lint** — linting
- **air** — hot-reload during development
