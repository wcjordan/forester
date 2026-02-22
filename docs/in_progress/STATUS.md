# STATUS — E2E Test: Log Storage Workflow

## Current state

- Planning complete. Ready to begin Stage 1.

## Stages

- [ ] Stage 1: Clock interface + FakeClock (`game/clock.go`)
- [ ] Stage 2: Wire clock into Game + Player
- [ ] Stage 3: Wire clock into render.Model + export TickMsg
- [ ] Stage 4: E2E test (`e2e_tests/log_storage_test.go`)

## Next

Start Stage 1: create `game/clock.go`.

## Key decisions

- FakeClock starts at `2024-01-01T00:00:00Z` so zero-value `time.Time` fields (deposit/move cooldowns) are already "in the past" and pass on first use.
- `TryDeposit` signature changes to accept explicit `now time.Time` — caller (Game.Tick) provides clock time.
- `NewModel` / `New` unchanged; `NewModelWithClock` / `NewWithClock` added for tests.
- ANSI escape codes stripped before View() assertions using a regex helper.
