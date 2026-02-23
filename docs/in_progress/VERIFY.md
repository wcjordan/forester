# Verify

## Primary gate
```
make check   # lint + test — must pass after every stage
```

## After Stage 1 (rename)
- `make check` passes
- `grep -r 'GhostLogStorage\|ghostOriginFor\|GhostType\|findDefForGhostStructureType' game/` → no hits

## After Stage 2 (foundation mechanic)
- `make check` passes
- `grep -r 'BuildOperation\|AdvanceBuild\|checkGhostContact\|nudgePlayerOutside\|\.Building\b' game/` → no hits
- Manual smoke-test with `make run`:
  - Cut 10 wood; foundation appears
  - Player cannot walk onto foundation tiles
  - Standing adjacent deposits 1 wood per ~500ms
  - After 20 deposits foundation becomes LogStorage
  - Player can deposit into built LogStorage as before

## Key behavioral checks (test coverage)
- `TestFoundationBlocksMovement` — player cannot move onto foundation tile
- `TestFoundationDepositBuilds` — 20 adjacent deposits complete the foundation
- `TestFoundationBecomesLogStorageAfterBuild` — tiles change to LogStorage on completion
- `TestFoundationDepositCooldown` — cooldown prevents faster-than-500ms deposits
- `TestDepositIntoBuiltStorage` — existing storage deposit tests still pass (no regression)
