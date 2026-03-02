# Future Idea: XP & Card Upgrade System (Remaining Work)

The core XP and card system is implemented (see `game/xp.go`, `game/upgrade.go`, `game/upgrades/`).
This file tracks what is **not yet built**.

---

## What's implemented

- XP earned from chopping (+1/wood), depositing (+1/wood), completing structures (+10 player / +20 villager)
- XP milestones with growing gaps (50, 75, 100, 125, …); game pauses and presents 3-card offer
- Card pool: faster harvesting, depositing, movement, building (stackable)
- Spawn Villager card (conditionally offered when an unoccupied house exists)

---

## Remaining work

### Village improvement upgrade cards
Cards that improve the village rather than the player directly:
- Villager move speed upgrade
- Lower structure spawn thresholds (e.g. house foundation triggers earlier)
- Villager carry capacity increase
- Storage capacity upgrade

### Rare / milestone-triggered cards
Distinct from regular XP milestone cards — triggered by specific village achievements:
- "4 houses built" milestone card
- "First resource depot built" milestone card
- Mechanism to distinguish rare cards from regular ones in the UI

### New resource types as XP sources
Stone, berries, fish, etc. would each earn XP when gathered — keeping the XP flow natural as the game expands.

### Unlock mechanics
Cards that unlock new behaviors entirely:
- Villagers autonomously cut trees (currently they don't — only the player does)
- New resource types become available to gather

---

## Open Questions

- Common vs. rare card rarity — how to signal rarity in the ASCII UI?
- Milestone card interaction: if XP milestone and village milestone fire simultaneously, queue both or merge into one offer?
- Upgrade stacking caps — should any upgrade have a maximum stack count?
