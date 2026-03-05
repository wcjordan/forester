# Village Progression Design

## High-Level Approach

The core progression philosophy is **emergent complexity through resource chains**. Each new tier of the game introduces a resource or building that:

1. Creates new demand for existing resources (wood is always needed — it just competes with more things as the game grows)
2. Unlocks a new category of production that was previously out of reach
3. Raises the floor for what "sustainable" means at that scale

The model is closer to **Anno** or **The Settlers** than Civilization: you are not clicking through a linear tech tree, but building physical infrastructure that must physically produce, transport, and consume goods. The tech/structure tree provides the *unlock order*, but the player's job is managing the *flow*.

### Progression Tension: Order of Magnitude Demand

Each tier transition is designed to feel like a step-change in complexity. When a player stabilizes wood production for ~3 villagers, the introduction of food demand creates a sudden ~3× increase in the number of production systems they must manage. When metal arrives, it creates another step-change: a single Smelter needs coal, ore, a mine, haulers, a forge, and a Smithy before it produces a single tool.

The goal is that the player is always chasing sustainability at the current tier while glimpsing the *shape* of the next tier challenge ahead.

### Structural Unlock Pattern

Buildings unlock in chains — a building rarely appears in isolation. The pattern is:

```
New terrain type → Harvesting building → Processing building → Consumer building → Population need
```

Example for the metal chain:
```
Ore Vein (terrain) → Mine → Smelter → Forge → Smithy
                                             ↓
                                     Iron Tools (global efficiency boost)
```

The Resource Depot acts as the Tier 1→2 transition hub: it does not produce goods itself, but its construction unlocks the ability to store and route non-wood resources, making the subsequent terrain and building unlocks coherent.

---

## Industries

Industries group related terrain types, buildings, and resources into production chains. Each chain has a raw extraction step, one or more processing steps, and a final good consumed by buildings or population.

### Wood & Forestry

The starting industry. Initially a pure extraction loop (chop → store → build). Grows into the fuel and construction materials backbone for all other industries.

| Building | Inputs | Outputs |
|---|---|---|
| (existing) Log Storage | — | Stores raw logs |
| Lumber Mill | Logs | Planks (required for most Tier 2+ buildings) |
| Charcoal Kiln | Logs | Charcoal (fuel for smelting, baking, heating) |
| Carpenter's Workshop | Planks | Furniture (housing quality), tools |
| Paper Mill | Wood pulp, Water | Paper (required for research) |
| Forester's Lodge | — | Plants new trees; sustains harvest radius |

*Key tension*: Wood serves four competing uses (building, charcoal, paper, furniture). Players who don't invest in a Forester's Lodge eventually exhaust the local forest and face a hard shortage.

---

### Stone & Earth

The first non-wood material chain. Unlocks durable buildings and the Brickyard, which produces the construction material required for Tier 3+ civic structures.

| Terrain / Building | Inputs | Outputs |
|---|---|---|
| Rock Outcroppings (terrain) | — | Source for quarrying |
| Quarry | Stone outcroppings | Stone blocks |
| Clay Deposits (terrain) | — | Source for clay |
| Clay Pit | Clay deposits | Raw clay |
| Lime Kiln | Stone + Charcoal | Lime (mortar component) |
| Brickyard | Clay + Charcoal | Bricks (durable construction material) |
| Stonecutter | Stone + Lime | Dressed stone (premium construction) |

*Key tension*: Stone and bricks both require charcoal, creating a direct competition with the metal chain during scale-up.

---

### Food & Agriculture

Food becomes a hard requirement once population exceeds a threshold (likely 3–4 villagers). Until then, foraging sustains the settlement. The food chain has the broadest variety of buildings and the most direct connection to population tier unlocks.

| Building | Inputs | Outputs |
|---|---|---|
| Hunting Lodge | — | Meat (passive; supplements early food) |
| Fishing Hut | Water adjacency | Fish |
| Farmstead | Cleared land | Grain, vegetables, flax |
| Granary | Wood / Stone | Food storage (replaces ad-hoc piles) |
| Windmill | Planks + Stone | Wheat → Flour |
| Bakery | Flour + Charcoal | Bread (satisfies Craftsmen-tier need) |
| Animal Farm | Wood + Planks | Cattle, pigs, chickens, sheep |
| Butcher | Animals | Meat (higher yield than hunting) |
| Dairy | Cattle | Milk, Cheese |
| Brewery | Grain + Water | Ale (happiness; not caloric) |
| Orchard | Land + time | Fruit (slow; passive yield once mature) |

*Key tension*: Grain is consumed by both Bakery (food) and Brewery (happiness). Early game, all grain goes to bread. Later, happiness becomes a real need and the player must choose to grow grain supply rather than divert it.

---

### Metal

The highest-leverage industry. Iron tools accelerate all other production systems by ~50%, but the infrastructure to produce them is the most involved chain in the game. Intentionally feels like a significant capital investment before any return.

| Terrain / Building | Inputs | Outputs |
|---|---|---|
| Iron Ore Vein (terrain) | — | Source for mining |
| Coal Seam (terrain) | — | Source for fuel (alternative to charcoal) |
| Mine | Ore vein | Iron ore |
| Coal Mine | Coal seam | Coal (more efficient fuel than charcoal) |
| Smelter / Bloomery | Iron ore + Charcoal/Coal | Pig iron |
| Forge | Pig iron | Iron bars |
| Smithy | Iron bars + Wood | Tools, hardware |
| Blast Furnace | Iron bars + Coal | Steel |
| Armorer | Steel + Cloth | Armor |
| Weaponsmith | Steel + Wood | Weapons |

*Key tension*: Bootstrapping. You cannot smelt metal without fuel; you cannot mine coal efficiently without iron picks; iron picks require smelting. The intended path is charcoal-first smelting → first tools → coal mine → coal replaces charcoal in smelting → charcoal freed for other uses.

---

### Textiles

A mid-tier chain that addresses the population clothing need. Relatively passive once running, but requires land and specialist labor.

| Building | Inputs | Outputs |
|---|---|---|
| Sheep Farm | Land + Wood | Wool |
| Flax Farm | Cleared land | Flax fiber |
| Spinning Workshop | Wool / Flax | Thread |
| Loom | Thread + Planks | Cloth |
| Tailor | Cloth | Clothing (Craftsmen-tier need) |
| Dye Works | Cloth + Plant dyes | Dyed cloth (luxury tier need) |

---

### Civic & Commerce

Non-production buildings that provide happiness, research, population tier unlocks, and trade access. Each civic building requires goods from other industries to build and often to operate.

| Building | Purpose | Key Unlock |
|---|---|---|
| Market Stall | Basic exchange of surplus goods | Trade efficiency |
| Market Square | Regional trade hub | Import / export routes |
| Grand Bazaar | City-scale trade | Exotic goods access |
| Well | Water access; sanitation radius | Prerequisite for Bathhouse |
| Bathhouse | Sanitation; disease prevention | Required for Craftsmen tier |
| Shrine | Happiness radius | Temple unlock |
| Temple | Major happiness + blessings | Cathedral unlock; Tier 3 civic |
| Cathedral | City-scale faith | Nobles tier unlock |
| School | Specialist profession unlock | Library, Academy access |
| Library | Research points/turn | Academy unlock; needs Paper |
| Academy / University | Advanced research | Top-tier upgrades and buildings |
| Town Hall | Governance; laws/edicts | Required for Tier 4 buildings |
| Watchtower | Map vision; threat warning | Barracks unlock |
| Barracks | Military training | Defense chain |
| Tavern | Happiness; travelers bringing trade | Inn upgrade |
| Trading Post | External trade routes | Exotic resource import |
| Harbor / Dock | Sea trade | Massive resource access expansion |

---

## Building / Village Tiers

Tier progression governs *which buildings can be constructed*, not just which resources exist. A Town Hall, for instance, cannot be built until Tier 3 prerequisites are met. This prevents players from skipping directly to advanced buildings without building a supporting base.

### Tier 0 — Survival (current state)

*Goal: Stay alive, accumulate enough wood to begin building.*

- Player chops trees manually
- Log Storage appears when inventory is full
- Houses appear as wood accumulates in storage
- Villagers are unlocked via XP cards
- All resource flow is wood only

---

### Tier 1 — Settlement (Resource Depot unlock)

*Trigger: ~4 houses built.*
*Goal: Diversify beyond wood; establish food supply.*

New buildings: Resource Depot, Hunting Lodge, Farmstead, Granary, Quarry, Lumber Mill, Charcoal Kiln
New resources: Stone, Grain, Meat/Fish, Charcoal, Planks
New terrain: Rock outcroppings, Clay deposits, Fertile plains, Water

The Resource Depot is the physical hub of this tier — it stores non-wood resources and acts as the distribution point for the early food and stone chains. Its presence on the map signals to the player that the game is expanding beyond the wood economy.

---

### Tier 2 — Village

*Trigger: Food chain running; ~50 wood stored in Granary equivalent; first stone building built.*
*Goal: Achieve food self-sufficiency; begin processing chains.*

New buildings: Windmill, Bakery, Animal Farm, Carpenter's Workshop, Brickyard, Shrine, Market Stall, Well, School
New resources: Flour, Bread, Bricks, Furniture, Livestock products
Population tier unlocked: Craftsmen

Craftsmen-tier villagers are more productive in specialized roles but require Bread (not raw grain) and access to a Bathhouse. This is the first point where population quality, not just quantity, matters.

---

### Tier 3 — Town

*Trigger: Craftsmen population tier reached; Smithy producing tools.*
*Goal: Establish metal production; unlock civic infrastructure.*

New buildings: Mine, Smelter, Forge, Smithy, Brewery, Dairy, Spinning Workshop, Loom, Bathhouse, Temple, Library, Town Hall, Market Square, Watchtower
New resources: Iron bars, Tools, Cloth, Clothing, Ale, Paper, Research points
Population tier unlocked: Burghers

Iron Tools are a global efficiency multiplier — the single most impactful unlock in the game. All harvesting, farming, mining, and construction speeds increase. The cost is the entire metal chain, which demands coal or large charcoal reserves.

Burghers require Ale, Clothing, and Market access. They unlock the Trade and Commerce buildings.

---

### Tier 4 — Proto-City

*Trigger: Town Hall built; trade routes established; Burghers population present.*
*Goal: Achieve regional trade, advanced research, and military capability.*

New buildings: Blast Furnace, Armorer, Weaponsmith, Academy, Cathedral, Grand Bazaar, Barracks, Trading Post, Harbor/Dock
New resources: Steel, Armor, Weapons, Advanced research output
Population tier unlocked: Scholars, Nobles

This tier is the endgame expansion area. Scholars require Paper, Library access, and comfort. Nobles require luxury goods (fine cloth, wine, art) and Cathedral access. Both provide powerful labor multipliers (research acceleration; military command).

---

## Population Types and Dynamics

Population tiers are inspired by Anno's citizen tier system. Higher-tier citizens provide more skilled labor but require a richer supply chain to maintain their tier.

| Tier | Name | Unlock Condition | Labor Provided | Needs |
|---|---|---|---|---|
| 0 | Settlers | Basic housing (current) | Manual labor (chop, carry, build) | Food (any), Shelter |
| 1 | Craftsmen | Carpentry + Smithy operating | Skilled production (smithing, baking, brewing) | Bread, Clothing, Bathhouse access |
| 2 | Burghers | Town Hall + Market Square | Commerce, trade, bookkeeping | Ale, Furniture, diverse food, Market |
| 3 | Scholars | Library + School | Research, medicine, planning | Paper, Comfort, Civic buildings |
| 4 | Nobles | Cathedral + Grand Bazaar | Administration, military command | Luxury goods, Fine cloth, Cathedral |

### Tier Mechanics

- A villager's tier is determined by whether their housing tier and all needs are met
- Unmet needs cause tier regression (a Craftsman without Bread reverts to Settler productivity)
- Higher-tier citizens produce more output per unit time in their specialized role, but contribute nothing outside their specialization
- This creates an ongoing tension: broad generalists vs. narrow specialists

### Foreman System (planned)

The Foreman is a promoted Craftsmen-tier villager who:
- Continues their task autonomously without player presence
- Applies a productivity bonus to nearby villagers doing the same task
- Can be assigned a default task and radius of influence
- Is the key mechanism allowing the player to "hand off" an industry and focus on building the next tier

---

## Scaling Challenges and Design Tensions

The game is designed around a series of inflection points where the current system becomes the bottleneck for the next system.

### Challenge 1 — The Wood Squeeze

*When it hits:* Mid Tier 1, when Charcoal Kiln, Brickyard, and Paper Mill all come online simultaneously.

Wood goes from "one use (building)" to "four competing uses (building, charcoal, furniture, paper)." Players who did not invest in a Forester's Lodge hit a sharp shortage. The resolution is sustainable forestry — managing harvest rotation so cleared areas regrow before the frontier moves too far.

### Challenge 2 — Food Multiplier

*When it hits:* Tier 1 → 2 transition, as population grows past 4–5 villagers.

Each new villager creates food demand. Food production requires cleared farmland (more tree cutting), farming labor (fewer cutters), processing buildings (more wood/stone to build), and fuel (charcoal). Solving food requires coordinating 4–5 systems simultaneously. Players who rushed population face this crisis hard; players who built Granary and Farmstead early have a buffer.

### Challenge 3 — The Fuel Bottleneck

*When it hits:* Tier 2 → 3, as Bakery, Brickyard, and Smelter all go online.

Charcoal from wood is the early universal fuel. When Smelting begins, charcoal demand spikes sharply. Coal mines are the long-term solution but require iron picks (which require smelting). The bootstrapping tension is real and intentional: start smelting on charcoal, use first tools to open a coal mine, transition smelting to coal, free up charcoal for other uses.

### Challenge 4 — Tool Investment Moment

*When it hits:* First time a Forge comes online.

Iron tools provide a ~50% speed boost to all production. But building the full metal chain (mine + smelter + forge + smithy) requires diverting wood, stone, charcoal, and labor from current production. Everything slows down during construction. The payoff is large but delayed — it's the most significant "spend now to earn later" bet in the game.

### Challenge 5 — Population Quality Wall

*When it hits:* Craftsmen tier unlock.

A player with 8 Settler-tier villagers suddenly faces the need to provide Bread (not grain), Clothing, and Bathhouse access for any to function as Craftsmen. All three require supply chains not yet running. This wall is a good design: it forces the player to build the Bakery, Tailor, and Bathhouse before population quality improves. Attempting to skip directly to Craftsmen without infrastructure causes regression.

### Challenge 6 — Transport Bottleneck

*When it hits:* Late Tier 2 / early Tier 3, as production sites and consumption sites diverge geographically.

Mines are where ore is, not where the Smithy is. Farmsteads are on open land, not adjacent to the Bakery. The road system (already implemented) becomes increasingly important. Higher-quality roads mean faster hauling villagers. Carts and pack animals (future upgrade cards?) allow bulk transport. A market in the middle of the production zone can smooth imbalances.

### Key Design Tensions

| Tension | Option A | Option B |
|---|---|---|
| Wood vs. Stone buildings | Cheap, fast wood | Expensive, durable stone |
| Charcoal vs. Coal fuel | Renewable, immediate | Efficient, needs mine first |
| Grain allocation | Bread (food) | Ale (happiness) |
| Population breadth vs. depth | Many generalist Settlers | Few specialist Craftsmen |
| Forester's Lodge vs. frontier expansion | Sustainable yield | Short-term gain, long-term shortage |
| Charcoal kiln vs. coal mine | No infrastructure needed | Higher efficiency after ramp-up |
| Build food chain now vs. defer | Steady state earlier | More wood for construction short-term |

None of these have a universally correct answer — the right choice depends on the map layout, current resources, and which crisis is most imminent.

---

## Upgrade Card Dynamics

The XP / card system extends through all tiers. Cards are the primary lever for player agency that is *not* about building placement — they answer "how does the player shape their village's specialization?"

### Card Categories by Tier

**Tier 0 — Survival (current)**
- Harvest speed (+10% per card, stackable)
- Carry capacity (+80 max wood per card)
- Deposit speed (+10%)
- Move speed (+10%)
- Build speed (+10%)
- Spawn Villager (offered when unoccupied house exists)

**Tier 1 — Settlement**
- Forester: trees replant 2× faster within cleared zones
- Stone tools: quarrying and stone-cutting speed +20%
- Pack animal: villagers carry +3 units of any resource
- Granary efficiency: food storage capacity +50%
- Charcoal yield: Charcoal Kiln produces +1 per batch

**Tier 2 — Village**
- Iron tools: all production speeds +30% (one-time; feels like a major milestone)
- Bread surplus: Bakery produces +1 extra loaf per batch
- Ale quality: Brewery produces stronger ale; happiness radius increases
- Expert farmer: Farmstead yield +25%
- Mill automation: Windmill operates without assigned villager

**Tier 3 — Town**
- Steel tools: all production speeds +50% (replaces iron tools tier)
- Guild organization: villagers in the same profession share a productivity bonus
- Trade routes: market exchange rates improve; less loss on surplus sales
- Promote Foreman: one Craftsmen-tier villager becomes a Foreman (key unlock)
- Coal mastery: Coal Mine output +25%; fuel substitution easier

**Tier 4 — Proto-City**
- Engineering: construction costs reduced 15%
- Grand Library: research rate doubled
- Naval access: Harbor trade volume +50%
- Military order: barracks trains units 40% faster
- Master craftsman: one specialist villager produces at legendary quality (output sells for bonus)

### Card Offer Gating

Cards are filtered by what is currently relevant:
- "Spawn Villager" only appears when an unoccupied house exists (current behavior)
- "Charcoal yield" only appears when a Charcoal Kiln is built
- "Iron tools" only appears once a Smithy is operational
- Foreman card only appears when a Craftsmen-tier villager exists

This prevents the card pool from being polluted with irrelevant offers and ensures each card feels timely.

### Stacking and Specialization

Most production cards are stackable (current behavior). A player who takes Harvest Speed three times is committing to a wood-focused playstyle. This creates soft "builds" — a player running 3× Harvest + 2× Charcoal yield is playing a very different game than one running 2× Pack Animal + 2× Trade routes. The roguelike element is that the card pool is random within the filtered set, so players must adapt their strategy to what they are offered.

---

## Open Questions

- What exactly does the Resource Depot do mechanically? (Generic multi-resource storage? Village center with effect radius? Trade node?)
- Population tier regression: gradual (productivity falls) or hard (villager refuses to work)?
- Should road quality directly affect hauling speed, or provide a villager-count multiplier in a region?
- Foreman radius of influence: fixed tiles, or scales with upgrades?
- Is there a population cap, or does the game scale indefinitely?
- What triggers the Tier 3 → 4 transition? (A specific building? A population count? A resource milestone?)
