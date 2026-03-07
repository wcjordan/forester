# Village Progression Design

## High-Level Approach

The core progression philosophy is **emergent complexity through resource chains**. Each new tier of the game introduces a resource or building that:

1. Creates new demand for existing resources (wood is always needed — it just competes with more things as the game grows)
2. Unlocks a new category of production that was previously out of reach
3. Raises the floor for what "sustainable" means at that scale

### Population Needs vs. Production Scale

Resource chains are the engine of progression. New buildings unlock new resources, and new resources create new demand — both for the chains that produce them and for the chains that feed them. **The dominant gameplay loop is scaling supply to meet the volume demands of a growing village.**

Villager consumption is real: higher-tier citizens physically walk to buildings to consume goods (Bread from a Bakery, Ale from a Brewery), and sustained consumption plus happiness is what makes them eligible for tier promotion. But the primary pressure is not "make sure each villager has bread" — it is that a village of 20 Craftsmen suddenly demands 20× the bread output of a village of 2, requiring a proportional scale-up of the entire grain → windmill → bakery chain. The tier system gates progression; the production chain is where the game lives.

### Progression Tension: Order of Magnitude Demand

Each tier transition is designed to feel like a step-change in complexity. When a player stabilizes wood production for ~3 villagers, the introduction of food demand creates a sudden increase in the number of production systems they must manage. When metal arrives, it creates another step-change: a single Smelter needs coal, ore, a mine, haulers, a forge, and a Smithy before it produces a single tool.

The goal is that the player is always chasing sustainability at the current tier while glimpsing the *shape* of the next tier challenge ahead.

### Structural Unlock Pattern

Buildings unlock in chains — a building rarely appears in isolation. The pattern is:

```
New terrain type → Harvesting building → Processing building → Refined good
                                                                    ↓
                                               Consumed by villagers (gates tier eligibility)
                                               AND feeds the next production chain
```

Example for the metal chain:
```
Ore Vein (terrain) → Mine → Smelter → Forge → Smithy
                                             ↓
                                     Iron Tools (global efficiency boost)
```

### Resource Depot as Progression Gate

The Resource Depot is the milestone that gates moving from Tier 0 (Survival) to Tier 1 (Settlement). It provides large storage for all resource types beyond wood, and its location becomes the new anchor point for subsequent structure spawning — shifting the "village center" away from the original log storage as the settlement grows. Future civic milestones (Temple, Market Square) may shift the center further still.

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
| Carpenter's Workshop | Planks | Furniture (housing quality upgrade) |
| Paper Mill | Logs | Paper (required for research) |
| Forester's Lodge | — | Plants new trees; sustains harvest radius |

*Key tension*: Wood serves four competing uses (building, charcoal, paper, furniture). Players who don't invest in a Forester's Lodge eventually exhaust the local forest and face a hard shortage.

---

### Stone & Earth

The first non-wood material chain. Unlocks durable buildings and the Brickyard, which produces the construction material required for Tier 3+ civic structures.

| Terrain / Building | Inputs | Outputs |
|---|---|---|
| Rock Outcroppings (terrain) | — | Source for quarrying |
| Quarry | Built on/adjacent to rock outcroppings | Stone blocks |
| Clay Deposits (terrain) | — | Source for clay |
| Clay Pit | Clay deposits | Raw clay |
| Lime Kiln | Stone + Charcoal | Lime |
| Brickyard | Clay + Charcoal | Bricks (durable construction material) |
| Stonecutter | Stone + Lime | Dressed stone (premium construction) |

*Key tension*: Stone and bricks both require charcoal, creating a direct competition with the metal chain during scale-up.

---

### Food & Agriculture

Early in the game, villagers with a food need can visit forests to forage for berries, slowly satisfying their hunger while in the woods. Forests have a finite food pool that depletes during foraging and regenerates slowly — enough to sustain a small settlement but unable to scale with population growth. A Hunting Lodge boosts foraging yield within its radius. This natural pressure pushes players toward proper agriculture.

| Building | Inputs | Outputs |
|---|---|---|
| Hunting Lodge | — | Boosts foraging yield in radius; adds game meat |
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

*Key tension*: Grain is consumed by both Bakery (food) and Brewery (happiness). Early game, all grain goes to bread. Later, happiness becomes a meaningful factor and players must grow grain supply rather than divert it.

---

### Metal

The highest-leverage industry. Iron tools accelerate all other production systems significantly, but the infrastructure to produce them is the most involved chain in the game. Intentionally feels like a significant capital investment before any return.

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

Non-production buildings that provide happiness, research, population tier eligibility, and trade access. Each civic building requires goods from other industries to build and often to operate.

| Building | Purpose | Key Unlock |
|---|---|---|
| Market Stall | Basic exchange of surplus goods | Trade efficiency |
| Market Square | Regional trade hub | Import / export routes |
| Grand Bazaar | City-scale trade | Exotic goods access |
| Well | Water access; sanitation radius | Prerequisite for Bathhouse |
| Bathhouse | Sanitation; happiness bonus | Craftsmen tier eligibility |
| Shrine | Happiness radius | Temple unlock |
| Temple | Major happiness + blessings | Cathedral unlock; Tier 3 civic |
| Cathedral | City-scale faith | Nobles tier eligibility |
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
- Villagers can forage in forests to meet early food needs
- All tracked resource flow is wood only

---

### Tier 1 — Settlement

*Trigger: Resource Depot built (~4 houses).*
*Goal: Diversify beyond wood; establish food supply.*

New buildings: Hunting Lodge, Farmstead, Granary, Quarry, Lumber Mill, Charcoal Kiln, Forester's Lodge
New resources: Stone, Grain, Meat/Fish, Charcoal, Planks
New terrain: Rock outcroppings, Clay deposits, Fertile plains, Water

The Resource Depot provides large multi-resource storage and shifts the village center anchor point — subsequent structure spawning orients around it rather than the original log storage. Its construction marks the end of the wood-only phase and opens the food and stone chains.

---

### Tier 2 — Village

*Trigger: Food chain self-sustaining; Brickyard producing bricks.*
*Goal: Achieve food self-sufficiency; begin processing chains; unlock first population tier.*

New buildings: Windmill, Bakery, Animal Farm, Carpenter's Workshop, Brickyard, Well, Bathhouse, Shrine, Market Stall, School
New resources: Flour, Bread, Bricks, Furniture, Livestock products
Population tier eligible: Craftsmen

Craftsmen-tier villagers are more productive in specialized roles but require sustained access to Bread and a Bathhouse — and must maintain happiness — to reach and hold their tier. This is the first point where population quality, not just quantity, matters.

---

### Tier 3 — Town

*Trigger: Craftsmen population present; Market Stall established.*
*Goal: Establish metal production; unlock civic infrastructure.*

New buildings: Mine, Smelter, Forge, Smithy, Brewery, Dairy, Spinning Workshop, Loom, Temple, Library, Town Hall, Market Square, Watchtower
New resources: Iron bars, Tools, Cloth, Clothing, Ale, Paper, Research points
Population tier eligible: Burghers

Iron Tools are the single most impactful unlock in this tier. All harvesting, farming, mining, and construction speeds increase. The cost is the entire metal chain, which demands coal or large charcoal reserves.

Burghers require Ale, Clothing, and Market access in addition to sustained happiness to reach their tier.

---

### Tier 4 — Proto-City

*Trigger: Town Hall built; trade routes established; Burghers population present.*
*Goal: Achieve regional trade, advanced research, and military capability.*

New buildings: Blast Furnace, Armorer, Weaponsmith, Academy, Cathedral, Grand Bazaar, Barracks, Trading Post, Harbor/Dock
New resources: Steel, Armor, Weapons, Advanced research output
Population tier eligible: Scholars, Nobles

Scholars require Paper, Library access, and sustained happiness and comfort. Nobles require luxury goods (fine cloth, wine, art) and Cathedral access. Both provide powerful labor multipliers once their tier is reached.

---

## Population Types and Dynamics

Population progression is consumption-driven and card-gated. Villagers physically walk to buildings to consume goods — a villager must travel to a Bakery to eat bread, not just have it exist in storage. Each villager independently tracks their own consumption history and happiness level. When a villager has sustained sufficient consumption of tier-appropriate goods and maintained happiness long enough, a **Promote Villager** card becomes available in the XP offer pool (similar to Spawn Villager). The player selects it to promote that specific villager.

The cost of promotion is the infrastructure challenge of sustaining the villager's ongoing needs at the new tier — not a one-time resource payment. Higher-tier citizens bring commensurately higher productivity in their specialized roles, but their maintenance requirements are steep and ongoing.

| Tier | Name | Eligibility Criteria | Labor Provided | Ongoing Needs |
|---|---|---|---|---|
| 0 | Settlers | Basic housing | Manual labor (chop, carry, build) | Food (any, including foraged), Shelter |
| 1 | Craftsmen | Bathhouse access + sustained Bread consumption + happiness | Skilled production (smithing, baking, brewing) | Bread, Clothing, Bathhouse, Happiness |
| 2 | Burghers | Market access + sustained Ale + Clothing + happiness | Commerce, trade, bookkeeping | Ale, Furniture, diverse food, Market, Happiness |
| 3 | Scholars | Library + School access + sustained Paper + happiness | Research, medicine, planning | Paper, Comfort, Civic buildings, Happiness |
| 4 | Nobles | Cathedral + Grand Bazaar access + luxury goods + happiness | Administration, military command | Luxury goods, Fine cloth, Cathedral, Happiness |

### Tier Mechanics

- Tier promotion is permanent — once a villager reaches a tier, they do not regress
- **Unmet needs cause a graduated productivity penalty** — a Craftsman without Bread does not revert to Settler, but their productivity drops significantly (floor around 50% or lower) until needs are restored
- Happiness is a small but real productivity modifier at all tiers, and a requirement for tier eligibility. It is provided by civic buildings (Shrine, Tavern, Temple) and goods (Ale)
- Higher-tier citizens are more productive in their specialized role but do not contribute to tasks outside it — this creates a meaningful tradeoff between breadth (many generalist Settlers) and depth (fewer, more capable specialists)

### Player Following and Task Weighting

Villagers in the player's vicinity naturally weight their tasks toward the player's recent activities. If the player has been chopping trees, nearby idle villagers are more likely to chop trees. This makes the player's movement and actions a form of soft direction without explicit assignment.

### Foreman System

The Foreman is a permanent promotion for a villager who has demonstrated sustained work in an area. When a villager qualifies (sufficient time on task, village prerequisites met), a **Promote Foreman** card appears in the XP offer pool with high priority.

Once promoted, the Foreman:
- Anchors to the industry's primary building (e.g., Log Storage for wood, Granary for food, Forge for metal)
- Sustains that industry autonomously within a radius of that building, even without the player present
- Enables villagers working within the radius to maintain their task weighting without player proximity

This is the mechanism that allows the player to "hand off" a mature industry and focus on developing the next one. A common rhythm: player establishes a new resource area, works it until it is running, a Foreman emerges and takes over, player moves to the next challenge.

The Foreman's radius of influence scales with specific upgrades targeted at that industry or building type.

---

## Scaling Challenges and Design Tensions

The game is designed around a series of inflection points where the current system becomes the bottleneck for the next.

### Challenge 1 — The Wood Squeeze

*When it hits:* Tier 1, as Charcoal Kiln comes online; deepens into Tier 2 as Brickyard and Paper Mill add further demand.

Wood goes from one use (building) to three or four competing uses (building, charcoal, paper, furniture). The squeeze is gradual — charcoal arrives first, then bricks, then paper — giving the player time to respond, but only if they notice the trend early. Players who did not invest in a Forester's Lodge eventually face a hard shortage. The resolution is sustainable forestry: managing harvest rotation so cleared areas regrow before the frontier moves too far.

### Challenge 2 — Food Does Not Scale

*When it hits:* Mid Tier 1, as population grows past 4–5 villagers.

Forest foraging sustains a tiny settlement but cannot scale. Each new villager draws from a forest food pool that regenerates slowly and depletes faster as the village grows. The pressure to build a real food chain (farmstead → granary → bakery) mounts naturally and unavoidably. This challenge does not require a hard gate — the world enforces it.

Once the food chain is running, the challenge shifts to volume: 20 Craftsmen need dramatically more bread than 5 Settlers. The grain → windmill → bakery chain must scale to match population, creating recurring investment pressure on the food chain throughout the game.

### Challenge 3 — The Fuel Bottleneck

*When it hits:* Tier 2 → 3, as Bakery, Brickyard, and Smelter all go online.

Charcoal from wood is the early universal fuel. When Smelting begins, charcoal demand spikes sharply. Coal mines are the long-term solution but require iron picks (which require smelting). The bootstrapping tension is real and intentional: start smelting on charcoal, use first tools to open a coal mine, transition smelting to coal, free up charcoal for other uses.

### Challenge 4 — Tool Investment Moment

*When it hits:* First time a Forge comes online.

Iron tools provide a significant speed boost to all production. But building the full metal chain (mine + smelter + forge + smithy) requires diverting wood, stone, charcoal, and labor from current production. Everything slows down during construction. The payoff is large but delayed — it's the most significant "spend now to earn later" bet in the game.

### Challenge 5 — Population Quality Pressure

*When it hits:* First Craftsmen promotions.

Promoting a villager to Craftsmen requires sustained Bread consumption and Bathhouse access — neither of which runs automatically. The player must build and staff the supply chain, then maintain it at scale. As more villagers reach Craftsmen tier, the ongoing demand for Bread, Clothing, and Ale grows proportionally. Letting those chains fall behind causes widespread productivity penalties rather than tier regression, but 50% productivity across a large workforce is a serious setback.

### Challenge 6 — Transport Bottleneck

*When it hits:* Late Tier 2 / early Tier 3, as production sites and consumption sites diverge geographically.

Mines are where ore is, not where the Smithy is. Farmsteads are on open land, not adjacent to the Bakery. Road quality directly affects hauling speed — villagers move faster on roads, so the quality of the road network between production and consumption sites becomes a real constraint. This is already partly addressed by the road formation system; the challenge deepens as the map scales.

### Key Design Tensions

| Tension | Option A | Option B |
|---|---|---|
| Wood vs. Stone buildings | Cheap, fast wood | Expensive, durable stone |
| Charcoal vs. Coal fuel | Renewable, immediate | Efficient, needs mine first |
| Grain allocation | Bread (food) | Ale (happiness / Burgher need) |
| Population breadth vs. depth | Many generalist Settlers | Fewer, more capable specialists |
| Forester's Lodge vs. frontier expansion | Sustainable yield | Short-term gain, long-term shortage |
| Promote now vs. sustain later | Tier benefits sooner | Ongoing needs may outpace supply |
| Build food chain now vs. defer | Steady state earlier | More wood for construction short-term |

None of these have a universally correct answer — the right choice depends on map layout, current resources, and which crisis is most imminent.

---

## Upgrade Card Dynamics

The XP / card system extends through all tiers. Cards are the primary lever for player agency that operates outside of direct player actions — travel, resource collection, and working alongside villagers. They answer "how does the player shape their village's specialization?" without placing buildings or assigning tasks directly.

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
- Bread surplus: Bakery produces +1 extra loaf per batch
- Ale quality: Brewery produces stronger ale; happiness radius increases
- Expert farmer: Farmstead yield +25%
- Mill automation: Windmill operates without assigned villager
- Promote Villager (offered when a villager meets tier eligibility criteria)

**Tier 3 — Town**
- Iron tools: all production speeds +40% (one-time; major milestone unlock)
- Guild organization: villagers in the same profession share a productivity bonus
- Trade routes: market exchange rates improve; less loss on surplus sales
- Promote Foreman (offered with high priority when a villager qualifies)
- Coal mastery: Coal Mine output +25%; fuel substitution easier

**Tier 4 — Proto-City**
- Steel tools: all production speeds +60% (replaces iron tools effect)
- Engineering: construction costs reduced 15%
- Grand Library: research rate doubled
- Naval access: Harbor trade volume +50%
- Master craftsman: one specialist villager produces at exceptional quality

### Card Offer Gating

Cards are filtered by what is currently relevant:
- "Spawn Villager" only appears when an unoccupied house exists (current behavior)
- "Charcoal yield" only appears when a Charcoal Kiln is built
- "Iron tools" only appears once a Smithy is operational
- "Promote Villager" only appears when a specific villager meets their tier's eligibility criteria
- "Promote Foreman" only appears when a villager qualifies, and is weighted to appear with high probability

This prevents the card pool from being polluted with irrelevant offers and ensures each card feels timely.

### Stacking and Specialization

Most production cards are stackable (current behavior). A player who takes Harvest Speed three times is committing to a wood-focused playstyle. This creates soft "builds" — a player running 3× Harvest + 2× Charcoal yield is playing a very different game than one running 2× Pack Animal + 2× Trade routes. The roguelike element is that the card pool is random within the filtered set, so players must adapt their strategy to what they are offered.

Production-specific cards (e.g., Coal mastery, Bread surplus) are intentionally narrow — they reward deep investment in a single chain rather than broad generalism.

---

## Open Questions

- What triggers the Tier 3 → 4 transition? Town Hall construction and/or trade milestones are strong candidates; defer to playtesting.
- Foreman eligibility criteria: how many ticks on-task, and what village prerequisites?
- Precise happiness values: what productivity bonus does happiness provide, and what is the exact penalty floor for unmet needs?
- Road quality tiers: should hauling speed scale continuously with WalkCount, or at fixed thresholds (trodden / road / paved)?
