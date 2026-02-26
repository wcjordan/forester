# Plan: Story Progression + World Condition Spawning

## Goal
Separate structure spawning into two systems:
1. **Story beats** – fire exactly once each, in order, for early-game tutorial/narrative triggers
2. **World condition rules** – use `StructureDef.ShouldSpawn` and may fire an unlimited number of times

## Context / design decisions
- Story beats are defined centrally in `game/story.go` (single slice, package-level, injectable in tests)
- Each beat has `ID string`, `Condition func(*Env) bool`, `Action func(*Env)`
- Fired beats are recorded in `State.CompletedBeats map[string]bool`
- `maybeAdvanceStory` iterates beats in order; skips completed; fires the first one whose condition is met; returns (one beat per tick)
- Beat actions look up structure defs via `findStructureDefByFoundationType` (registry lookup inside `game` package — avoids import cycles)
- `maybeSpawnFoundation` becomes the **world condition loop**: removes the `HasStructureOfType` guard; each `ShouldSpawn` is now responsible for its own gate logic
- `logStorageDef.ShouldSpawn` → `return false` (story beat owns the initial spawn; multi-instance future work)
- `houseDef.ShouldSpawn` → world condition: true when `len(House) >= 1 && len(FoundationHouse) == 0`
- Placement helper `(s *State) spawnFoundationAt(def StructureDef)` extracted and shared by both systems

## Non-goals
- Multiple log storage instances (future)
- Villager-occupancy conditions for house (future — no villagers yet)
- Event-driven (push) beats — polling per tick is fine for now

## Initial story beats (in order)
1. `initial_log_storage`: condition `player.Wood >= player.MaxCarry` → spawn log storage foundation near player
2. `initial_house`: condition `stores.Total(Wood) >= 50` → spawn house foundation near spawn

---

## Stage 1 – Story beat infrastructure

**Goal**: Add the beat system; wire it into `Harvest`; behavior is additive (existing `maybeSpawnFoundation` still runs).

**Steps**:
1. Add `CompletedBeats map[string]bool` to `State`; init in `newState()`
2. Create `game/story.go`:
   - `StoryBeat` struct
   - `storyBeats []StoryBeat` package-level var
   - `findStructureDefByFoundationType(StructureType) StructureDef` helper (searches `structures` slice)
   - Extract `(s *State) spawnFoundationAt(def StructureDef)` from `maybeSpawnFoundation` body
   - `(s *State) maybeAdvanceStory(env *Env)` method
   - Define `storyBeats` with the two initial beats
3. Update `maybeSpawnFoundation` to use `spawnFoundationAt`
4. Call `s.maybeAdvanceStory(env)` from `Harvest` (before `maybeSpawnFoundation`)
5. `make check` — must pass
6. Commit: `Feat: story beat infrastructure (initial_log_storage + initial_house beats)`

**Exit criteria**: `make check` passes; both story beats exist in `storyBeats`; `maybeAdvanceStory` is called from `Harvest`.

---

## Stage 2 – Update ShouldSpawn + remove HasStructureOfType guard

**Goal**: Wire `ShouldSpawn` as world condition; log storage no longer self-spawns.

**Steps**:
1. `logStorageDef.ShouldSpawn` → `return false`
2. `houseDef.ShouldSpawn` → `len(env.State.World.StructureTypeIndex[game.House]) >= 1 && len(env.State.World.StructureTypeIndex[game.FoundationHouse]) == 0`
3. Remove `HasStructureOfType` guard from `maybeSpawnFoundation`
4. `make check` — must pass (existing e2e tests should still work; same observable behavior)
5. Commit: `Feat: log storage spawn moved to story beat; house ShouldSpawn becomes world condition`

**Exit criteria**: `make check` passes; house foundation spawns via world condition after first house is built.

---

## Stage 3 – Update unit tests

**Goal**: Align `state_test.go` with the new system; add targeted tests for story beats and world conditions.

**Steps**:
1. Add `CompletedBeats: make(map[string]bool)` to every manually constructed `State{}` in `state_test.go`
2. Update `testLogStorageDef.ShouldSpawn` → `return false` (matches production behavior)
3. Add `withTestStoryBeats(t, beats []StoryBeat)` helper that saves/restores `storyBeats`
4. Rename `TestFoundationDoesNotSpawnTwice` → `TestStoryBeatFiresOnce`: verify `maybeAdvanceStory` called twice only spawns one foundation
5. Keep `TestFoundationSpawnsWhenInventoryFull` — now exercises story beat via `Harvest` loop (should still pass)
6. Update `TestFoundationLocationIsAllGrassland` / `TestFoundationLocationBetweenPlayerAndSpawn` to call `s.spawnFoundationAt(testLogStorageDef{})` directly (tests placement logic, not the trigger)
7. Add `TestHouseWorldConditionSpawnsAfterBuild`: set up a built house in world, verify `maybeSpawnFoundation` spawns a new house foundation; also verify it does NOT spawn when `FoundationHouse` already exists
8. `make check` — must pass
9. Commit: `Test: update state tests for story beat + world condition systems`

**Exit criteria**: `make check` passes; all new and updated tests are meaningful; no duplicate coverage.

---

## Stage 4 – Final review + PR

**Steps**:
1. `make check` clean pass
2. Self-review diff for scope creep, dead code, missing edge cases
3. Remove `docs/in_progress/` files
4. Commit cleanup if needed
5. Open PR against `main`
