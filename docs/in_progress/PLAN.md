# PLAN.md — Phase 1: TUI Rendering + Player Movement

## Goal
Make the game visible and interactive in the terminal. After this phase, `make run` opens
a full-screen TUI showing a map with the player character, and arrow/WASD keys move the
player around. This is the foundation everything else (trees, cutting, roads) builds on.

## Non-Goals
- No tree cutting or wood counter yet (Phase 1b)
- No map generation / procedural content yet (Phase 1b)
- No AI or villagers
- No persistence

---

## Stages

### Stage 1 — Integrate bubbletea and render a static world
**Goal**: `make run` opens a TUI, renders a viewport of the world, player visible, quit with `q`.

Steps:
1. Add `bubbletea` and `lipgloss` to `go.mod`
2. Create `game/input.go` — define `Msg` types (MoveMsg, QuitMsg)
3. Create `render/model.go` — bubbletea `Model` wrapping `*game.Game`; implement `Init`, `Update`, `View`
4. Render a viewport of fixed size (e.g. 40 cols × 20 rows) centered on player
5. Render `@` for player, `.` for grassland, `#` for forest terrain
6. Wire `main.go` to launch the bubbletea program
7. Commit

**Exit criteria**:
- `make run` opens TUI, player visible as `@` on a `.` grid
- `q` / `ctrl+c` exits cleanly
- `make check` passes

---

### Stage 2 — Player movement + viewport scrolling
**Goal**: Arrow keys and WASD move the player. Viewport follows. Bounds enforced.

Steps:
1. Add `MovePlayer(dx, dy int)` to `game/player.go` (bounds-checked against World)
2. Map key presses → `MoveMsg{dx, dy}` in bubbletea `Update`
3. Viewport always centers on player (clamped at world edges)
4. Add a status line below the viewport: `Player: (x, y)  Wood: 0`
5. Commit

**Exit criteria**:
- All four directions move `@` visibly
- Player cannot move out of world bounds
- Viewport scrolls smoothly as player approaches edges
- Status line shows correct coordinates
- `make check` passes

---

## Architecture Notes

```
main.go
  └─ launches bubbletea program with render.NewModel(game.New())

render/
  model.go    ← bubbletea Model: Init / Update / View
               wraps *game.Game, owns viewport offset

game/
  player.go   ← add MovePlayer(dx, dy)
  input.go    ← MoveMsg, QuitMsg types (no bubbletea import here)
```

Key constraint: `game/` must NOT import `bubbletea`. All TUI concerns live in `render/`.

---

## Decisions (locked)
- **Viewport**: responsive — fills the full terminal, handles `tea.WindowSizeMsg`
- **Key bindings**: both WASD and arrow keys move the player
- **Status bar**: single line pinned to bottom — `Player: (x, y)  Wood: 0`
