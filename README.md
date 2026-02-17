# Forester

A terminal-based city builder where you play as a character who develops a village through organic, emergent gameplay. Roads form where you walk, structures appear where you work, and villagers learn by following your example.

See [docs/PROJECT_PLAN.md](docs/PROJECT_PLAN.md) for the full design document.

## Prerequisites

- **Go 1.23+** — [install](https://go.dev/dl/)
- **golangci-lint** — `brew install golangci-lint`
- **air** (optional, for hot-reload) — `go install github.com/air-verse/air@latest`

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
game/                # Core game logic
  game.go            # Game orchestrator
  state.go           # Game state (player + world)
  player.go          # Player entity
  world.go           # World grid and tile access
  tile.go            # Tile and terrain types
docs/
  PROJECT_PLAN.md    # Game design document
```

## Tech Stack

- **Go** — game logic and rendering
- **bubbletea** — terminal UI (planned)
- **golangci-lint** — linting
- **air** — hot-reload during development
