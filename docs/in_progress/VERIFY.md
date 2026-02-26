# Verify

## Primary gate
```bash
make check   # lint + test (run after every stage)
```

## Stage-by-stage success criteria

### Stage 1
- `make check` passes
- `game/story.go` exists with `StoryBeat`, `storyBeats`, `maybeAdvanceStory`
- `State.CompletedBeats` initialized in `newState()`
- `maybeAdvanceStory` called from `Harvest`
- `spawnFoundationAt` helper extracted and used by both systems

### Stage 2
- `make check` passes (including e2e tests)
- `logStorageDef.ShouldSpawn` returns false
- `houseDef.ShouldSpawn` returns true only when House >= 1 and FoundationHouse == 0
- No `HasStructureOfType` guard in `maybeSpawnFoundation`
- `TestLogStorageWorkflow` e2e still passes (log storage spawns via story beat)
- `TestHouseWorkflow` e2e still passes (first house via story beat, then world condition enables future houses)

### Stage 3
- `make check` passes
- All `State{}` literals include `CompletedBeats: make(map[string]bool)`
- `testLogStorageDef.ShouldSpawn` returns false
- `TestStoryBeatFiresOnce` present and passing
- `TestHouseWorldConditionSpawnsAfterBuild` present and passing
- Placement tests use `spawnFoundationAt` directly

### Stage 4
- `make check` clean
- No `docs/in_progress/` files remain
- PR open against `main`

## What failure looks like
- E2e test fails → story beat condition or placement logic doesn't match old behavior
- `TestFoundationSpawnsWhenInventoryFull` fails → story beat not wired into `Harvest`
- Lint fails → unused import or exported type naming issue
