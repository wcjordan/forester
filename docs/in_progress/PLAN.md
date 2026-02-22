# Plan: Extract build + progression logic from state.go

## Goal
Reduce `state.go` to core state concerns (struct, constructor, coordination methods).
Move building mechanics and ghost/progression logic to dedicated files.

## Non-goals
- No behavioral changes — pure mechanical refactor
- Do not move `BuildOperation` out of `structure.go` (keep scope small)
- Do not rename or refactor any methods

---

## Stage 1 — Create `progression.go`

Move from `state.go` → new `game/progression.go`:
- `HasStructureOfType`
- `ghostOriginFor`
- `maybeSpawnGhosts`
- `findValidLocation`
- `isValidArea`
- `abs` (utility used only by `findValidLocation`)

Remove those functions from `state.go`.

**Exit criteria:** `make check` passes.
**Commit:** `Refactor: extract ghost/progression logic to progression.go`

---

## Stage 2 — Create `build.go`

Move from `state.go` → new `game/build.go`:
- `checkGhostContact`
- `nudgePlayerOutside`
- `AdvanceBuild`
- `findDefForBuilt`
- `findDefForGhost`

Remove those functions from `state.go`.

**Exit criteria:** `make check` passes.
**Commit:** `Refactor: extract build mechanics to build.go`

---

## Stage 3 — Cleanup + PR

- Self-review: verify `state.go` only contains core state concerns
- Delete `docs/in_progress/` files
- Push branch, open GitHub PR

**Exit criteria:** PR open, CI green.
