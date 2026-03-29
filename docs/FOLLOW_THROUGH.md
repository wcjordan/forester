# Follow-Through

Remaining items from substantially-completed projects. These are the next concrete tasks.

---

## Sprites: Stage S5 — Villager Animation

**From**: `docs/completed/BETTER_SPRITES.md`

**Goal:** Villagers animate while moving (walk cycle, facing direction) and optionally play task-specific animations while chopping or delivering.

**Work:**
- Generate a villager spritesheet from the [Universal LPC Character Generator](https://liberatedpixelcup.github.io/Universal-LPC-Spritesheet-Character-Generator/) with a different outfit/color than the player. Place at `assets/sprites/villager-spritesheet.png`.
- Embed and load it in `assets/assets.go` (same pattern as `PlayerSheet`).
- Add `VillagerDir [maxVillagers]int` and `VillagerPrevPos [maxVillagers]game.Point` to `EbitenGame`; derive direction from position delta each tick.
- Pre-slice villager walk frames in `render/sprites.go` (same `[4][8]` layout as `playerWalkFrames`).
- Add `spriteForVillager(dir, frame int)` using `dirFrom` + walk-row lookup (same pattern as player).
- Optionally: drive slash (chopping) and thrust (delivering) animations from `VillagerTask` state.

**Exit criteria:**
- Villagers animate through walk frames in the correct facing direction while moving.
- Villagers show idle (frame 0, last facing direction) when stationary.
- `make check` passes.

---

## Graphics: Phase G3 — Web Deployment

**From**: `docs/completed/GRAPHICS_MIGRATION.md`

**Goal:** Playable in a browser, shareable via URL. WASM build already compiles (`make wasm`).

**Work:**
1. Add `wasm_exec.js` + minimal HTML shell page in `web/`
2. Add `make web` Makefile target (builds WASM and copies to `web/`)
3. Test in Chrome, Firefox, and Safari
4. Verify keyboard input works in browser context

**Exit criteria:**
- Game runs in browser without Go installed
- Input (keyboard) works correctly in browser context
- Performance acceptable on a mid-range laptop

---

## Graphics: Phase G4 — Isometric View *(stretch goal)*

**From**: `docs/completed/GRAPHICS_MIGRATION.md`

**Goal:** Re-render the world with isometric perspective. No game logic changes — only the `Draw()` path changes.

**Approach:**
- Coordinate transform: `screenX = (x - y) * tileW/2`, `screenY = (x + y) * tileH/2`
- Swap 32×32 square sprites for 64×32 isometric tile sprites
- Entities (player, villagers, trees) get isometric-angle sprites
- Draw order: depth-sort tiles (tiles farther from camera drawn first)

---

## Villagers & Automation — Phase 3 Remaining

**From**: `docs/PROJECT_PLAN.md`

- **Village improvement upgrade cards**: villager move speed, lower structure spawn thresholds, villager carry capacity, storage capacity upgrade
- **Following behavior**: villagers trail player and mirror current task
- **Foreman promotion**: player promotes a villager; foreman works autonomously at a building
- **Foreman influence**: foreman encourages nearby villagers to join its task; influence radius scales with upgrades
- **Resource Depot structure**: triggered after 4 houses; gates Tier 1 progression; shifts village center anchor

---

## Road Formation — Remaining

**From**: `docs/future_ideas/road_formation.md`

Current state: 2-level system done (Trodden=20 WalkCount, Road=100). Speeds: Grassland=150ms, Trodden=120ms, Road=90ms.

- **Better Road** (3rd level): high traffic + villager contribution threshold
- **Road decay**: WalkCount degrades on unused paths over time
- **Road formation on Forest terrain**: currently only Grassland tiles accumulate WalkCount
- **Villager movement speed benefit**: villagers use A* road weighting but don't apply the speed bonus

---

## XP & Cards — Remaining

**From**: `docs/future_ideas/xp_and_upgrades.md`

Core system is implemented. Remaining card types:

### Village improvement cards
- Villager move speed upgrade
- Lower structure spawn thresholds (e.g. house foundation triggers earlier)
- Villager carry capacity increase
- Storage capacity upgrade

### Rare / milestone-triggered cards
- "4 houses built" milestone card
- "First Resource Depot built" milestone card
- UI distinction between rare and regular cards

### New resource types as XP sources
Stone, berries, fish, etc. earn XP when gathered — keeps XP flow natural as game expands.

### Open questions
- Common vs. rare card rarity — how to signal rarity in the UI?
- Milestone card interaction: if XP milestone and village milestone fire simultaneously, queue both or merge into one offer?
- Upgrade stacking caps — should any upgrade have a maximum stack count?
