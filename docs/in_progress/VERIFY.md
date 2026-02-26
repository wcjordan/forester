# Verification: Villager Implementation

## Primary gate
```bash
make check   # lint + test (must pass after every commit)
```

## Per-stage checks

### Stage 1
```bash
make check
# Expected: all tests pass, no lint errors

make run
# Manually: build first house (deposit 50 wood into log storage, then build house foundation)
# Expected: a cyan 'v' appears near the house; status bar shows "Villagers: 1/1"
```

**Test cases (automated):**
- `TestVillagerSpawnsOnHouseBuilt`: build a house via OnBuilt; assert `len(state.Villagers) == 1`
- `TestVillagerStatusBar`: verify status bar string includes "Villagers: 1/1"

### Stage 2
```bash
make check

make run
# Expected:
# - Villager 'v' visibly moves around the map
# - Log storage fill slowly increases when villager is chopping
# - When log storage is fairly full, villager starts delivering to houses
# - Status bar "Log: X/500" updates over time
```

**Test cases (automated):**
- `TestWithdrawFrom`: deposit 10, withdraw 5 → stored = 5, returned = 5; withdraw 10 → returned = 5
- `TestTotalCapacity`: register 2 instances each capacity 100 → TotalCapacity = 200
- `TestVillagerChopTask`: tick villager near a tree; assert it moves toward tree, harvests, then moves toward storage
- `TestVillagerDeliverTask`: pre-fill storage, tick villager; assert it fetches from storage and delivers to house

## Assumptions
- `e2e_tests/` blank-imports `game/structures` which registers all defs; villager spawning will fire naturally
- World seed 42 (DefaultSeed) is used in all automated tests
- Tests use `FakeClock` and seeded RNG for determinism

## Success
- `make check` exits 0
- No new TODOs without a plan stage to address them
- Status bar displays correctly for edge cases: 0 villagers, 0 log storage capacity
