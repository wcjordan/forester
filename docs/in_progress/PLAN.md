# Resource Depot — Phase 1 Plan

## Goal

Add the Resource Depot as the Tier 0→1 progression gate: a 5×4 structure that spawns after 4 houses are built, costs 800 wood to build, registers as large wood storage, and triggers a +100 carry capacity card on completion.

## Constraints / Non-goals

- **No anchor shift** (deferred to Phase 2 / stacked PR): houses continue spawning near world center regardless of depot location.
- **Placeholder Ebitengine sprite**: colored rectangle only; real sprite asset is future work.
- **No Tier 1 buildings or resources**: the depot is a milestone structure; nothing unlocks yet.

---

## Stages

### Stage 1 — Game logic *(game/)*

**Files:**
- `game/structures/resource_depot.go` — `resourceDepotDef` implementing `StructureDef` + `StorageDef`
- `game/upgrades/large_carry_upgrade.go` — `largeCarryCapacityUpgrade` (+100 MaxCarry, ID: `"large_carry_capacity"`)

**Details:**
- Footprint: 5×4
- Build cost: 800 wood
- Placement: spawn-anchored (`UseSpawnAnchoredPlacement() = true`)
- Storage: registers wood storage (capacity 2000) in `OnBuilt`
- Story beats:
  - Order 500 `"initial_resource_depot"`: condition = 4 houses built AND no depot pending/built; action = `SpawnFoundationByType(FoundationResourceDepot)`
  - Order 600 `"first_resource_depot_built"`: condition = depot built; action = `AddOffer(["large_carry_capacity"])`
- `RegisterVillagerDepositType(ResourceDepot)` — villagers can deposit harvested wood here
- `RegisterVillagerDeliveryType(FoundationResourceDepot)` — villagers can deliver wood toward the build cost
- `OnPlayerInteraction`: same build-deposit pattern as house (adjacent deposit, 1 wood per build-cooldown tick)

**Exit criteria:** `make check` passes; unit tests in `game/structures` cover trigger conditions.

---

### Stage 2 — Rendering *(render/)*

**TUI (`render/tui_model.go`):**
- Add `resourceDepotStyle` (bold cyan)
- Add `structures.FoundationResourceDepot` to existing foundation `?` case
- Add `structures.ResourceDepot` → `D` in bold cyan

**Ebitengine (`render/sprites.go` + draw loop):**
- Add `ResourceDepot` placeholder: a solid-color 32×32 rectangle (amber/gold) at the NW anchor tile
- Non-anchor tiles draw terrain only (same pattern as log storage / house)

**Exit criteria:** `make check` passes; `D` appears in TUI; colored rectangle appears in Ebitengine.

---

### Stage 3 — E2E test

**File:** `e2e_tests/resource_depot_test.go`

**Scenario:**
1. Build log storage + accept carry upgrade.
2. Build 4 houses (drive enough ticks, auto-drain XP offers).
3. After 4th house: verify `FoundationResourceDepot` spawns.
4. Navigate adjacent to the foundation; deposit 800 wood (tick loop).
5. Verify `ResourceDepot` structure exists.
6. Verify `large_carry_capacity` card offer is pending; select it; verify MaxCarry increased by 100.

**Exit criteria:** `make check` passes.

---

### Stage 4 — Commit + PR

- `make check` passes (lint + tests).
- Commit with clear message.
- Open PR against `main`.

---

## Phase 2 (stacked PR — separate branch)

After Phase 1 merges:
- Modify house spawn placement to use the depot's origin as the anchor when a depot exists.
- Houses spawn near the depot rather than near world center.
- Requires modifying `findValidLocationNearSpawn` or `houseDef.ShouldSpawn` / `SpawnFoundationByType` to accept an optional anchor override.
