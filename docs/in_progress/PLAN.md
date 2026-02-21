# PLAN: Structures Refactor

## Goal
Introduce a `StructureDef` interface so each structure type encapsulates its own
trigger condition and player-interaction behavior. Refactor `state.go` to dispatch
generically through a registry. Split LogStorage-specific logic into its own file.

## Non-Goals
- No new structures yet (House, Depot) — this prepares the pattern only
- No package changes — stays in `game` package, organised by file naming convention
- No changes to `Tile.Structure` or the `StructureType` int enum

---

## Stage 1 — Define `StructureDef` interface and registry (additive only)

**File:** `game/structure.go` (extend existing)

Add to the existing file:
```go
// StructureDef describes the behavior of one structure type.
type StructureDef interface {
    GhostType()     StructureType
    BuiltType()     StructureType
    Footprint()     (w, h int)
    BuildTicks()    int
    ShouldSpawn(s *State) bool      // domain trigger condition only
    OnAdjacentTick(s *State)        // called each tick player is adjacent to built structure
}

// structures is the registry of all known structure definitions.
var structures []StructureDef
```

Exit criteria:
- `make check` passes with no changes to behavior

Commit: `Add StructureDef interface and registry to structure.go`

---

## Stage 2 — Implement `logStorageDef` in a new file (additive only)

**New file:** `game/log_storage.go`

- Unexported `logStorageDef` struct implementing `StructureDef`
- Move `LogStorageBuildTicks` constant here from `structure.go`
- Register via `init()`:

```go
func init() { structures = append(structures, logStorageDef{}) }
```

Method implementations:
| Method | Value |
|---|---|
| `GhostType()` | `GhostLogStorage` |
| `BuiltType()` | `LogStorage` |
| `Footprint()` | `4, 4` |
| `BuildTicks()` | `LogStorageBuildTicks` |
| `ShouldSpawn(s)` | `s.TotalWoodCut >= 10` |
| `OnAdjacentTick(s)` | deposit 1 wood: `s.Player.Wood--; s.LogStorageDeposited++` (if Wood > 0) |

Exit criteria:
- `make check` passes; no behavior change (nothing in `state.go` calls the registry yet)

Commit: `Add logStorageDef implementing StructureDef`

---

## Stage 3 — Generalize `state.go` dispatch + update tests

Replace all type-specific dispatch with registry iteration.

### Changes to `state.go`

| Old | New | Notes |
|---|---|---|
| `maybeSpawnGhost()` | `maybeSpawnGhosts()` | loop over registry; generic "already placed" guard; calls `findValidLocation(w, h)` |
| `findGhostLocation()` | `findValidLocation(w, h int)` | takes footprint dims instead of hardcoded 4×4 |
| `isValid4x4(x, y)` | `isValidArea(x, y, w, h int)` | generalise dimensions |
| `checkGhostContact()` | same name | use `findDefForGhost(st)` helper to look up def; use `def.Footprint()`, `def.BuildTicks()`, `def.BuiltType()` |
| `GhostOrigin()` | `ghostOriginFor(st StructureType)` | unexported; takes which ghost type to find |
| `TryDeposit() bool` | `TickAdjacentStructures()` | loop over registry; calls `def.OnAdjacentTick(s)` when adjacent to `def.BuiltType()` |
| `AdvanceBuild()` | same name | fix hardcoded `LogStorage` → use `s.Building.Target` |

New helper (unexported, in `state.go`):
```go
// findDefForGhost returns the StructureDef whose GhostType matches st, or nil.
func findDefForGhost(st StructureType) StructureDef { ... }
```

Generic `maybeSpawnGhosts` loop pattern:
```go
for _, def := range structures {
    if !def.ShouldSpawn(s) { continue }
    if s.HasStructureOfType(def.GhostType()) || s.HasStructureOfType(def.BuiltType()) { continue }
    w, h := def.Footprint()
    cx, cy := s.findValidLocation(w, h)
    if cx >= 0 { s.World.SetStructure(cx, cy, w, h, def.GhostType()) }
}
```

### Test updates (`state_test.go`)

| Old call | New call |
|---|---|
| `s.maybeSpawnGhost()` | `s.maybeSpawnGhosts()` |
| `s.TryDeposit()` | `s.TickAdjacentStructures()` |

`TryDeposit` tests check a bool return — rewrite those assertions to check state:
- "returns false when Wood is 0" → assert `s.Player.Wood` still 0 after call
- "returns false when not adjacent" → assert `s.LogStorageDeposited` still 0

Exit criteria:
- `make check` passes
- All existing test cases still covered

Commit: `Generalize state.go to dispatch through StructureDef registry`

---

## Constraints
- Keep `HasStructureOfType` on `State` (still needed for generic spawn guard and tests)
- `nudgePlayerOutside` already takes `(rx, ry, rw, rh int)` — no change needed
- `BuildOperation` stays in `structure.go` (generic build infrastructure)
- `StructureType` int enum stays in `tile.go`
