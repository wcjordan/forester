# Future Idea: XP & Card Upgrade System

## Concept

Vampire Survivors-style progression: gathering resources earns XP, level-ups trigger card selection events where the player picks from a few upgrades. Rare cards unlock at village milestones.

## Mechanics

### XP
- Player earns XP from gathering resources (wood, future: stone, berries, etc.)
- Villagers do NOT earn the player XP (they contribute to the village, not the player's growth)
- XP curve and level-up frequency to be tuned via playtesting

### Card Selection
- On level-up: pause and present 3 upgrade cards to choose from
- Regular cards: triggered by XP milestones
- Rare cards: triggered by village progression milestones (e.g. first house built, 4 houses built)

### Upgrade Types
- **Player abilities**: move faster, cut faster, carry more wood
- **Village improvements**: villagers work faster, structures upgrade sooner
- **Unlock mechanics**: villagers can cut trees, new resource types become available

## Open Questions

- Exact XP curve (how fast do levels come)?
- How many cards offered per level (3 is the classic Vampire Survivors number)?
- Which upgrades are common vs. rare?
- How do rare milestone cards interact with regular level-up cards?

## Why Deferred

Deprioritized in favor of structure progression (Phase 2), which provides a more tangible sense of building. Upgrades work best once there are multiple systems (roads, villagers, multiple resource types) worth upgrading.

## Implementation Notes

- `game/xp.go` - XP tracking and level-up logic
- `game/upgrades.go` - Upgrade definitions and application to game state
- `render/cards.go` - Card selection UI (pause game, show 3 options, apply choice)
