# CLAUDE.md

## Project

See `docs/GAME_DESIGN.md` for core mechanics and `README.md` for project overview.

## Repo Map (entrypoints)

See `docs/GETTING_AROUND.md` for a full navigation guide covering every file, the libraries used in each area, and key architectural patterns.

Quick reference:
- `main.go` — Entry point. Defaults to Ebitengine window; `--tui` flag runs bubbletea TUI. Blank-imports `game/structures`, `game/upgrades`, `game/resources`.
- `main_tui.go` — (`//go:build !js`) `shouldRunTUI()` / `runTUI()` — bubbletea startup.
- `main_wasm.go` — (`//go:build js`) Stubs so WASM builds compile without bubbletea.
- `game/` — Core game package (all game logic lives here)
  - `game.go` — `Game` struct, `New()`, `Tick()` orchestrator; owns `State`, `StorageManager`, `VillagerManager`
  - `state.go` — `State` struct (serializable: Player, World, XP, HouseOccupancy, pendingOfferIDs)
  - `player.go` — `Player` entity (position, inventory, cooldowns)
  - `world.go` — `World` grid, `NewWorld()`, bounds/tile access, `AddStructure()`
  - `tile.go` — `Tile`, `TerrainType`; `StructureType` aliased from `game/core`
  - `structure.go` — `StructureDef` interface + `RegisterStructure()`
  - `resource.go` — `ResourceDef` interface + `RegisterResource()`
  - `upgrade.go` — `UpgradeDef` interface + `RegisterUpgrade()`
  - `villager.go` — `Villager`, `VillagerManager`; autonomous chop/deliver behavior
  - `xp.go` — `AwardXP()`, milestone thresholds, card offer selection
  - `spawn.go` — `maybeSpawnFoundation()`, placement helpers
  - `story.go` — ordered one-shot story beats
  - `clock.go` — `Clock` interface for test time injection
  - `storage.go` — `ResourceStorage` / `StorageInstance`
  - `storage_manager.go` — `StorageManager` (deposit/withdraw aggregation)
  - `game/core/` — `StructureType` leaf package (no upstream deps)
  - `game/geom/` — Pure geometry helpers (`Point`, `FindPath`, `SpiralSearchDo`)
  - `game/resources/` — `woodDef` (implements `ResourceDef`, registers via `init()`)
  - `game/structures/` — `logStorageDef`, `houseDef` (register via `init()`)
  - `game/upgrades/` — all upgrade cards (register via `init()`)
- `render/tui_model.go` — (`//go:build !js`) bubbletea `Model` (TUI + card selection overlay)
- `render/ebiten_model.go` — `EbitenGame` Ebitengine renderer (solid-color tile grid, WASD input)
- `render/util.go` — Shared render utilities (`clamp`)
- `e2e_tests/` — End-to-end tests with injected clock + RNG
- `docs/PROJECT_PLAN.md` — Full game design document
- `docs/GRAPHICS_MIGRATION.md` — Ebitengine renderer migration plan

---

## Verification commands

```bash
make check   # lint + test (primary gate — run this)
make test    # tests only (with race detector)
make lint    # golangci-lint only
make build   # compile binary
make run     # build and run
make dev     # hot-reload with air
make e2e_viz  # visual E2E playback in terminal
make clean   # remove build artifacts
make format  # format code w/ gofmt
make wasm    # compile WASM binary (forester.wasm)
```

---

## Git Worktrees

**Use the `EnterWorktree` tool for all non-trivial work.** Each task should run in an isolated git worktree to keep changes contained and reviewable. Worktrees are created at `.claude/worktrees/<name>`.

When spawning subagents via the Agent tool, pass `isolation: "worktree"` to give each subagent its own isolated repo copy. The worktree is automatically cleaned up if the subagent makes no changes.

A `WorktreeCreate` hook automatically copies `assets/sprites/` into each new worktree. This is required because sprites are gitignored but embedded at compile time via `go:embed`, which does not follow symlinks — without the copy, any package importing `forester/assets` (including `e2e_tests`) will fail to compile.

---

## Agent workflow (Plan → Implement → Verify)

Use a verification-driven iterative looping workflow
Favor small testable steps, externalized state, and explicit verification.

### Principles (keep in mind always)
- Small diffs; avoid rewrites.
- Use repo code + test output as truth.
- Tests and linting must pass after each implementation step.
- Don’t expand scope silently.
- Be pragmatic to keep changes small.  Adapt to the project's current state.
- Externalize state (plans, decisions, work in progress) into transient files (see below).
- Use `TaskCreate` / `TaskUpdate` to track in-conversation steps for multi-part work.
- Long chat context can degrade quality. Reset context when appropriate (see below).

### Simplicity Means

- Single responsibility per function/class
- Avoid premature abstractions
- Code should be clear in intent rather than clever.  Be boring and obvious.
- No clever tricks - choose the boring solution
- If you need to explain it, it's too complex

### Transient working files

Use native **plan mode** (`EnterPlanMode`) for planning and tracking non-trivial work. Two transient files in `<PROJECT_ROOT>/docs/in_progress/` supplement the plan:

1) `STATUS.md` — current state (≤10 bullets)
- What’s done / next
- Current failures/blockers
- Key decisions

2) `NEED_HELP.md` — only when stuck (see rule below)

Delete these files when all work is complete.

---

## Planning

For any non-trivial changes, use `EnterPlanMode` to break down the problem into stages and track progress. The plan should be concise and actionable (5 stages max). Each stage should have explicit testable outcomes and specific test cases that prove correctness, plus an instruction to commit after completion.

Plans are working documents. Revise as new information is discovered. Remove transient files (`STATUS.md`, `NEED_HELP.md`) when all work is done.

When finalizing planning, ask me clarifying questions about anything ambiguous.
Ask me questions one at a time. Questions should clarify the plan and build on my previous answers.
Revise the plan based on my answers. Clarify all ambiguity before starting on implementation.

---

## Implementation loop (repeat per stage)

1. Follow existing patterns (find 2–3 similar examples).
2. Add new tests first when feasible; otherwise add coverage before finishing.
3. Implement minimal change.
4. Run `make check` to verify.
5. Update plan status and `STATUS.md` with command + result.
6. Cleanup once tests are passing.
7. Commit with clear message describing the change.

---

## Stuck rule (hard stop)

Maximum **3 attempts** per issue. If still blocked:
- STOP and update `NEED_HELP.md` with:
  - What you tried
  - Exact errors/output
  - 2–3 alternative approaches, libraries, abstractions, patterns, etc
  - A simpler reframing / smaller subproblem

---

## Context resets

Reset / restart from files at boundaries:
- After completing or materially revising a plan stage
- After thrash (repeated failures)
- Before final review/polish

After reset, treat only these as authoritative:
- Repo contents
- Current plan (plan mode)
- `STATUS.md`, `NEED_HELP.md`
- Current diffs + latest verification output

---

## Quality gates

Definition of Done:
- Tests + lint pass (`make check`)
- Implementation matches plan exit criteria
- No new TODOs without adding a plan stage to address them

NEVER:
- Bypass hooks with `--no-verify`
- Disable tests instead of fixing them
- Commit broken code

ALWAYS:
- Commit incrementally
- Update plan / `STATUS.md` as you go
- Prefer boring, readable code

If verification fails:
- Treat failures as data
- Reassess the plan if needed
- Avoid rationalizing or ignoring failing signals

If issues are found:
- Update the plan or status
- Re-enter the implementation loop

---

## General Guidelines

- Do not silently change behavior without updating the plan.
- Do not expand scope without noting it explicitly.
- Avoid speculative refactors unless justified in the plan.
- Prefer evidence (tests, code, output) over narrative explanation.

### Architecture Principles

- **Composable architecture** - Create composable components with minimal responsibilities
- **Explicit over implicit** - Clear data flow and dependencies
- **Test-driven when possible** - Never disable tests, fix them
- **Optional interface duck-typing for StructureDef extensions** — new placement or behavior variants are wired as optional interfaces checked via type assertion (e.g. `spawnAnchoredPlacer`, `spawnAnchorOverrider`), not by widening the `StructureDef` base interface. Keep the core interface minimal.

### Code Quality

- When committing:
  - Tests should be passing and code should be linted (`make check` should pass)
  - Self-review changes
  - Ensure commit message explains "why"
- **Name all multi-value return types** — `golangci-lint` (gocritic `unnamedResult`) rejects unnamed returns on exported methods and interface signatures. Always write `(x, y int)` not `(int, int)`.
- **Avoid shadowing built-ins** — `golangci-lint` rejects variable names like `cap`, `len`, `new`, `make`.

### Error Handling

- Fail fast with descriptive messages
- Include context for debugging
- Handle errors at appropriate level
- Never silently swallow exceptions

### Decision Framework

When multiple valid approaches exist, choose based on:

1. **Testability** - Can I easily test this?
2. **Readability** - Will someone understand this in 6 months?
3. **Consistency** - Does this match project patterns?
4. **Simplicity** - Is this the simplest solution that works?
5. **Reversibility** - How hard to change later?

### Test Guidelines

- Test behavior, not implementation
- Prefer snapshot testing to complex assertions
- Use existing test utilities/helpers
- Tests should be deterministic

#### E2E test gotchas (`e2e_tests/`)

- **`tick` vs `tickDraining`**: Use plain `tick` when you need to inspect a pending upgrade offer. `tickDraining` calls `drainOffers()` which immediately accepts and clears any pending offer — you won't see it.
- **Player position affects near-player spawn**: `findValidLocationNearPlayer` walks from the player toward world center. If the player is already at world center, the walk has zero steps and only considers the player's own tile (which is blocked). Move the player off center before triggering any beat that uses near-player foundation placement.
- **Cooldown stacking near multiple structures**: A player adjacent to both a foundation and a storage structure will fire both build and deposit cooldowns each tick, burning through inventory twice as fast. When testing focused building, lock the other cooldown: `g.State.Player.SetCooldown(game.Deposit, clock.Now().Add(time.Hour))`.
- **Story beats are sequential and unexported**: `completedBeats` is unexported; you cannot skip beats. E2E tests must advance through all preceding beats naturally — factor this into test design time estimates.
