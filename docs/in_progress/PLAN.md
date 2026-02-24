# House Building — Implementation Plan

## Summary
Add the House structure following the same StructureDef registry pattern as Log Storage.
- Spawn trigger: current wood stored in Log Storage ≥ 50
- Footprint: 2×2, build cost: 50 wood
- OnBuilt: present a 2-card milestone offer (Build Speed vs. Deposit Speed)
- One house max (same guard as Log Storage — skip if foundation or built instance exists)

## Constraints / Non-goals
- No multiple houses in this feature (capped at 1)
- No villager spawning (visual milestone only, mechanic hook for later)
- No Resource Depot (out of scope)

---

## Stage 1: Tile types + house structure + render glyphs

**Goal:** House can spawn, be built, and render correctly.

**Steps:**
1. `game/tile.go`: add `FoundationHouse`, `House` to `StructureType` constants.
2. `game/house.go`: implement `houseDef{}` (register via `init()`):
   - `FoundationType()` → `FoundationHouse`, `BuiltType()` → `House`
   - `Footprint()` → 2×2, `BuildCost()` → 50
   - `ShouldSpawn(env)` → `env.Stores.Total(Wood) >= 50`
   - `OnPlayerInteraction`: deposit into foundation (same pattern as logStorageDef);
     when adjacent to built house do nothing (no storage).
   - `OnBuilt`: queue 2-card offer `["build_speed", "deposit_speed"]`
3. `game/state.go`: add `FoundationProgress() (float64, bool)` helper that returns
   the progress ratio and existence of any active foundation, using `StructureIndex`
   to look up `BuildCost()`. Renderer uses this instead of the hardcoded constant.
4. `render/model.go`:
   - Add glyphs: `FoundationHouse` → `?` (same yellow style), `House` → `H` (bold yellow).
   - Replace the hardcoded `LogStorageBuildCost` status-bar progress with `FoundationProgress()`.

**Commit:** `feat: House tile types, houseDef skeleton, render glyphs`

---

## Stage 2: Separate build and deposit intervals on Player

**Goal:** Make "build speed" and "deposit speed" mechanically distinct.

**Steps:**
1. `game/player.go`:
   - Add `Build CooldownType` constant (distinct from `Deposit`).
   - Add `BuildInterval`, `DepositInterval time.Duration` to `Player`; both default
     to `DepositTickInterval` in `NewPlayer`.
2. `game/log_storage.go`:
   - Foundation deposit path: check `Build` cooldown, queue with `p.BuildInterval`.
   - Storage deposit path: check `Deposit` cooldown, queue with `p.DepositInterval`.
3. `game/house.go`:
   - Foundation deposit path: check `Build` cooldown, queue with `p.BuildInterval`.
4. Tests: update any existing tests that directly set/check the `Deposit` cooldown
   for foundation deposits to use `Build` instead.

**Commit:** `refactor: separate Build and Deposit cooldowns; player-level deposit intervals`

---

## Stage 3: Upgrade cards + 2-card milestone screen

**Goal:** Build/deposit speed upgrades are selectable when house completes.

**Steps:**
1. `game/deposit_upgrades.go`: two upgrade types registered via `init()`:
   - `build_speed`: reduces `Player.BuildInterval` by 10%.
   - `deposit_speed`: reduces `Player.DepositInterval` by 10%.
2. `render/model.go` — `renderCardScreen`:
   - When `len(offer) == 1`: existing single-card layout (unchanged).
   - When `len(offer) == 2`: render two card boxes stacked vertically inside the
     outer border; each card shows its key binding (`[ 1 ]` / `[ 2 ]`); footer
     says "Press 1 or 2 to choose an upgrade".
3. `render/model.go` — `Update`:
   - When `HasPendingOffer()` and key is `"2"`: call `SelectCard(1)`.

**Commit:** `feat: build_speed/deposit_speed upgrades + 2-card milestone screen`

---

## Stage 4: E2E test + verification pass

**Goal:** Prove the full house workflow end-to-end; ensure `make check` passes.

**Steps:**
1. `e2e_tests/house_test.go`: `TestHouseWorkflow` — full scenario:
   - Build a Log Storage first (reuse the state after log storage E2E setup).
   - Deposit ≥ 50 wood into Log Storage (via adjacent ticks) until house foundation appears.
   - Verify foundation blocks movement into its footprint.
   - Deposit 50 wood to complete the house.
   - Verify `House` tile renders as `H`.
   - Verify 2-card offer appears; select one and verify the corresponding interval shrinks.
2. Run `make check`; fix any failures.

**Commit:** `test: E2E house workflow`

---

## Exit Criteria
- `make check` (lint + tests with race detector) passes clean.
- House foundation spawns only when ≥ 50 wood is stored in Log Storage.
- House builds by depositing 50 wood while adjacent to foundation.
- Completed house renders as `H` on the map.
- Two-card milestone screen appears; each key applies the correct upgrade.
- No regressions in the Log Storage E2E test.
