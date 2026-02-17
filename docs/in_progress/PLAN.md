# Bootstrap Plan: Dev Environment, Linting, Testing

## Stage 1: Project Initialization
- Go module init, .gitignore, minimal main.go + game package with test
- Exit: `go build` and `go test ./...` pass

## Stage 2: Development Tooling
- golangci-lint config, air config, Makefile
- Exit: `make build`, `make test`, `make lint`, `make check` all pass

## Stage 3: Scaffold Core Packages
- game/state.go, player.go, world.go, tile.go with tests
- Exit: `make check` passes, all packages have tests

## Stage 4: Documentation
- Update CLAUDE.md repo map + verification commands
- Create README.md with setup instructions
- Exit: Docs complete, transient files cleaned up
