# STATUS.md — Phase 1: TUI Rendering + Player Movement

## State: PLANNING

## Completed
- (none yet)

## Next
- Resolve open questions (viewport size, key bindings, status bar)
- Stage 1: Integrate bubbletea, render static world

## Blockers
- None

## Key Decisions
- `game/` will NOT import bubbletea — clean separation of logic and rendering
- `render/` package will own the bubbletea model and viewport state
- Viewport: responsive, fills terminal (WindowSizeMsg)
- Keys: both WASD and arrow keys
- Status bar: bottom line, `Player: (x, y)  Wood: 0`
- Two stages planned: static render first, then movement
