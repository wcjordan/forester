# Plan: Player Sprite Animation (Walk, Slash, Thrust)

## Goal

Animate the player character using the Universal LPC spritesheet:
- Walk animation while moving (4-directional)
- Slash animation while chopping trees
- Thrust animation while building a foundation

Villagers remain static (deferred — see `docs/BETTER_SPRITES.md`).

## Constraints / Non-goals

- Game logic is unchanged except two small fields added to `game.Player`
- No villager animation changes
- No changes to other sprite stages (S2–S4)

## Spritesheet Layout (Universal LPC, 64×64 per frame)

Row groups — each group has 4 rows (up / left / down / right):

| Rows  | Animation | Frames | Used |
|-------|-----------|--------|------|
| 0–3   | Spellcast | 7      | no   |
| 4–7   | Thrust    | 8      | yes — building |
| 8–11  | Walk      | 9      | yes — movement |
| 12–15 | Slash     | 6      | yes — chopping |
| 16–19 | Shoot     | 13     | no   |
| 20    | Hurt      | 6      | no   |

Direction within each group: row+0=up, row+1=left, row+2=down, row+3=right.

Walk: frame 0 = idle, frames 0–7 cycle while moving (~8fps).
Slash: 6 frames over 500ms (resets each harvest tick, so loops while actively chopping).
Thrust: 8 frames over 500ms (resets each build tick, so loops while actively building).

Animation priority: Slash > Thrust > Walk > Idle.

---

## Stage 1 — Game layer: LastHarvestAt + LastBuildAt

**Goal:** Add timing signals to `Player` so the render layer can trigger slash/thrust.

**Steps:**
1. Add `LastHarvestAt time.Time` and `LastBuildAt time.Time` fields to `Player` in `game/player.go`
2. In `game/resources/wood.go` `Harvest()`: set `p.LastHarvestAt = now` when `harvest > 0` for any target tile
3. In `game/structures/house.go` `houseOnPlayerInteraction()`: set `p.LastBuildAt = now` when a build deposit fires (alongside `p.QueueCooldown`)
4. Add test cases verifying both fields are set on the relevant events
5. `make check` passes; commit

**Exit criteria:**
- `player.LastHarvestAt` is set when wood is actually harvested (harvest > 0)
- `player.LastBuildAt` is set when a build deposit fires
- `make check` passes

---

## Stage 2 — Asset loading + docs

**Goal:** Load `player-spritesheet.png`; update README and BETTER_SPRITES.md.

**Steps:**
1. Add `//go:embed sprites/player-spritesheet.png` and `PlayerSheet *ebiten.Image` to `assets/assets.go`
2. Update `README.md` Setup section: add "Player Spritesheet" subsection with generate URL and placement instructions
3. Update `docs/BETTER_SPRITES.md`: update S1 to reflect static villagers decision; add villager animation future task
4. `make check` passes; commit

**Exit criteria:**
- Game compiles with `PlayerSheet` loaded
- README documents how to obtain the spritesheet
- `make check` passes

---

## Stage 3 — Walk animation

**Goal:** Player animates through directional walk frames while moving; idle when still.

**Steps:**
1. Add `playerMoving bool` and `animTick int` to `EbitenGame` in `render/ebiten_model.go`
2. In `Update()`: set `playerMoving` from key state; increment `animTick` while moving, reset to 0 when not
3. Add `dirFrom(dx, dy int) int` helper (0=up, 1=left, 2=down, 3=right) in `render/sprites.go`
4. Replace `assets.Player` with `assets.PlayerSheet` in `render/sprites.go`; remove static `playerImg` pre-slice
5. Update `spriteForPlayer(dir, animRow, frame int) drawArgs` to crop 64×64 from `PlayerSheet`
6. In `Draw()`: compute dir from `player.FacingDX/FacingDY`; compute walk frame from `animTick`; pass to `spriteForPlayer`
7. `make check` passes; commit

**Exit criteria:**
- Player cycles walk frames (8 frames ~8fps) in the correct direction while WASD/arrow keys are held
- Player shows idle (frame 0, facing direction) when no key held
- `make check` passes

---

## Stage 4 — Slash + Thrust animations

**Goal:** Slash plays while chopping; thrust plays while building.

**Steps:**
1. Define `animKind` type (`animIdle`, `animWalk`, `animSlash`, `animThrust`) in `render/sprites.go`
2. Add `playerAnimState(player, now)` helper: checks `LastHarvestAt` (< 500ms → slash), `LastBuildAt` (< 500ms → thrust), `playerMoving` → walk, else idle
3. Compute slash frame: `int(elapsed.Milliseconds() * 6 / 500) % 6`
4. Compute thrust frame: `int(elapsed.Milliseconds() * 8 / 500) % 8`
5. Wire into `Draw()` replacing the walk-only logic from Stage 3
6. `make check` passes; commit

**Exit criteria:**
- Slash plays (6 frames, ~500ms window) after each harvest; loops while actively chopping
- Thrust plays (8 frames, ~500ms window) after each build deposit; loops while actively building
- Walk resumes once both windows expire and no key held
- `make check` passes

---

## Stage 5 — PR

- Final self-review of all diffs
- Open PR against `main`
