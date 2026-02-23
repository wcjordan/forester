# Status

## Current
- Planning complete; ready to implement

## Stages
- [ ] Stage 1: Rename ghost → foundation
- [ ] Stage 2: Foundation mechanics (blocks movement + resource deposit build)
- [ ] Stage 3: Push + PR

## Decisions
- Build cost: 20 wood
- Deposit rate: reuse existing 500ms Deposit cooldown
- Foundation indexed in StructureIndex on placement (so TickAdjacentStructures can find it)
- `OnPlayerInteraction` detects foundation vs built by checking tile type at origin
- `FoundationDeposited map[Point]int` added to State for tracking deposit progress
