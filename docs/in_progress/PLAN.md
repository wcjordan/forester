# Plan: Foundation-based structure building

## Goal
Replace the time-based ghost/build mechanic with a resource-deposit model:
- Rename "ghost" → "foundation" throughout
- Foundations block movement (like built structures)
- Player deposits wood adjacent to a foundation to build it (20 wood = complete)
- Remove `AdvanceBuild`, `checkGhostContact`, `nudgePlayerOutside`

## Non-goals
- No UI/render changes
- No new resource types
- No new structure types

---

## Stage 1 — Rename ghost → foundation

Pure textual rename; zero logic change.

**Files to update:**
- `tile.go`: `GhostLogStorage` → `FoundationLogStorage`, update comment
- `structure.go`: `GhostType()` → `FoundationType()`, `findDefForGhostStructureType` → `findDefForFoundationType`
- `log_storage.go`: `GhostType()` → `FoundationType()`, return `FoundationLogStorage`
- `progression.go`: `ghostOriginFor` → `foundationOriginFor`, `def.GhostType()` → `def.FoundationType()`
- `build.go`: `def.GhostType()` → `def.FoundationType()`, update comments
- `player.go`: update comment on `MovePlayer`
- `state_test.go`: all `GhostLogStorage` → `FoundationLogStorage`
- `player_test.go`: rename subtest label + `GhostLogStorage` → `FoundationLogStorage`

**Exit criteria:** `make check` passes, no remaining "ghost" references in code (comments ok).

**Commit:** `Rename: ghost → foundation for structure building`

---

## Stage 2 — Foundation mechanics (blocks movement + resource deposit build)

All behavior changes in one stage.

### 2a. Foundation blocks movement
- `player.go` `MovePlayer`: change guard from `tile.Structure != LogStorage` to
  `tile.Structure != LogStorage && tile.Structure != FoundationLogStorage`

### 2b. Index foundation on placement
- `progression.go` `maybeSpawnGhosts`: after `SetStructure`, call
  `s.World.IndexStructure(cx, cy, w, h, def)` so `TickAdjacentStructures` can find the foundation

### 2c. Replace `BuildTicks()` with `BuildCost()` in interface
- `structure.go`: replace `BuildTicks() int` with `BuildCost() int` in `StructureDef`
- Remove `BuildOperation` struct + `Progress()`/`Done()` methods
- Remove `findDefForFoundationType` helper that was only used by old build.go

### 2d. Resource tracking + foundation deposit mechanic
- `state.go`: remove `Building *BuildOperation`; add `FoundationDeposited map[Point]int`
- `state.go` `newState()`: initialize `FoundationDeposited`
- `log_storage.go`:
  - Remove `LogStorageBuildTicks`; add `LogStorageBuildCost = 20`
  - `BuildCost() int` returns `LogStorageBuildCost`
  - `OnPlayerInteraction`: check tile at origin; if `FoundationLogStorage`, deposit 1 wood per
    cooldown toward `BuildCost()`; when complete call `SetStructure` → `IndexStructure` → `OnBuilt`

### 2e. Remove old build machinery
- `build.go`: delete file (was `checkGhostContact`, `nudgePlayerOutside`, `AdvanceBuild`)
- `state.go` `Move()`: remove `s.checkGhostContact()` call
- `game.go` `Tick()`: remove `g.State.AdvanceBuild()` call
- `structure.go`: remove `findDefForFoundationType` (was only used by deleted build.go)

### 2f. Update tests
- `state_test.go`: replace `TestBuildMechanic` with new foundation deposit tests:
  - Player cannot move onto foundation
  - Adjacent player deposits wood (cooldown applies)
  - Foundation completes after 20 deposits → becomes LogStorage
  - Foundation tiles replaced by LogStorage after completion
- `player_test.go`: update walkable-foundation subtest → now blocked (not walkable)

**Exit criteria:** `make check` passes; no references to `BuildOperation`, `AdvanceBuild`,
`checkGhostContact`, `nudgePlayerOutside`, `Building`.

**Commit:** `Feat: foundation-based structure building with resource deposit`

---

## Stage 3 — Push + PR

- `git push origin structure_building`
- Open PR against `main`
