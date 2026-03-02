# CLAUDE.md

## Project

See `docs/PROJECT_PLAN.md` for an idea of the project.

## Repo Map (entrypoints)

See `docs/GETTING_AROUND.md` for a full navigation guide covering every file, the libraries used in each area, and key architectural patterns.

Quick reference:
- `main.go` — Entry point. Creates and runs the game. Blank-imports `game/structures`, `game/upgrades`, `game/resources`.
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
  - `game/geom/` — Pure geometry helpers (`Point`, `findPath`, `spiralSearchDo`)
  - `game/resources/` — `woodDef` (implements `ResourceDef`, registers via `init()`)
  - `game/structures/` — `logStorageDef`, `houseDef` (register via `init()`)
  - `game/upgrades/` — all upgrade cards (register via `init()`)
- `render/model.go` — bubbletea `Model` (TUI + card selection overlay)
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
```

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
- Long chat context can degrade quality. Reset context when appropriate (see below).

### Simplicity Means

- Single responsibility per function/class
- Avoid premature abstractions
- Code should be clear in intent rather than clever.  Be boring and obvious.
- No clever tricks - choose the boring solution
- If you need to explain it, it's too complex

### Transient working files (authoritative)
Store these in: `<PROJECT_ROOT>/docs/in_progress/`

1) `PLAN.md` — required for non-trivial work
- For each stage: goal, constraints/non-goals, steps, exit criteria

2) `VERIFY.md` — how to prove correctness
- Exact commands (tests/lint/build)
- Env assumptions
- What success/failure looks like

3) `STATUS.md` — current state (≤10 bullets)
- What’s done / next
- Current failures/blockers
- Key decisions

4) `NEED_HELP.md` — only when stuck (see rule below)

Commit changes to `docs/in_progress` after planning is complete.
Delete these files when all the work is complete.

---

## Planning

For any non-trivial changes, break down the problem to subtasks and create a plan in `PLAN.md`.
The plan should be concise and actionable (5 stages max).
Add testable outcomes and specific test cases in `VERIFY.md` and status of subtasks to `STATUS.md`
Each stage in `PLAN.md` should include an instruction to commit the work after that stage is complete.

Plans are working documents. Revise as new information is discovered.
Update the status of each stage as you progress and commit progress.
Remove transient files when all stages are done.

When finalizing planning, ask me clarifying questions about anything ambiguous.
Ask me questions one at a time.  Questions should clarify the plan and build on my previous answers.
Revise the plan based on my answers.  Clarify all ambiguity before starting on the implementation steps.

---

## Implementation loop (repeat per stage)

1. Follow existing patterns (find 2–3 similar examples).
2. Add new tests first when feasible; otherwise add coverage before finishing.
3. Implement minimal change.
4. Verify using `VERIFY.md` 
5. Update `STATUS.md` with command + result.
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
- After creating or materially revising `PLAN.md`
- After a vertical slice / stage completion
- After thrash (repeated failures)
- Before final review/polish

After reset, treat only these as authoritative:
- Repo contents
- `PLAN.md`, `VERIFY.md`, `STATUS.md`, `NEED_HELP.md`
- Current diffs + latest verification output

---

## Quality gates

Definition of Done:
- Tests + lint pass (per `VERIFY.md`)
- Implementation matches `PLAN.md` exit criteria
- No new TODOs without adding a plan stage to address them

NEVER:
- Bypass hooks with `--no-verify`
- Disable tests instead of fixing them
- Commit broken code

ALWAYS:
- Commit incrementally
- Update `PLAN.md` / `STATUS.md` as you go
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
- Avoid speculative refactors unless justified in `PLAN.md`.
- Prefer evidence (tests, code, output) over narrative explanation.

### Architecture Principles

- **Composable architecture** - Create composable components with minimal responsibilities
- **Explicit over implicit** - Clear data flow and dependencies
- **Test-driven when possible** - Never disable tests, fix them

### Code Quality

- When committing:
  - Tests should be passing and code should be linted (`make check` should pass)
  - Self-review changes
  - Ensure commit message explains "why"

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
