# STATUS.md — Phase 1: TUI Rendering + Player Movement

## State: COMPLETE

## Completed
- Stage 1: bubbletea integrated, responsive viewport, player visible as `@`
- Stage 2: WASD + arrow key movement, bounds enforced, status bar at bottom
- `make check` passing (lint + tests)

## Next
- Phase 1b: map generation (trees, forest patches) + auto-cut mechanic

## Blockers
- None

## Key Decisions
- `game/` will NOT import bubbletea — clean separation of logic and rendering
- `render/` package will own the bubbletea model and viewport state
- Viewport: responsive, fills terminal (WindowSizeMsg)
- Keys: both WASD and arrow keys
- Status bar: bottom line, `Player: (x, y)  Wood: 0`
- Two stages planned: static render first, then movement
