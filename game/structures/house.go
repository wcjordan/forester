package structures

import (
	"time"

	"forester/game"
	"forester/game/geom"
)

const houseBuildCost = 50

// FoundationHouse and House are the StructureType values for the two stages of
// this structure. Defined here so external packages can reference them without
// importing all of game/ or editing a central enum.
const (
	FoundationHouse game.StructureType = "foundation_house"
	House           game.StructureType = "house"
)

func init() {
	game.RegisterStructure(houseDef{})
	game.RegisterVillagerDeliveryType(FoundationHouse)

	// Order 300: spawn the first house foundation once enough wood has been deposited.
	// NOTE: 50 matches houseBuildCost so the player has enough wood on hand after
	// depositing to immediately build the house. If houseBuildCost changes, update
	// this threshold to match.
	game.RegisterStoryBeat(300, "initial_house",
		func(env *game.Env) bool {
			return env.Stores.Total(game.Wood) >= houseBuildCost
		},
		func(env *game.Env) bool {
			return game.SpawnFoundationByType(env, FoundationHouse)
		},
	)

	// Order 400: queue the build/deposit speed upgrade offer when the first house is completed.
	game.RegisterStoryBeat(400, "first_house_built",
		func(env *game.Env) bool {
			return env.State.World.HasStructureOfType(House)
		},
		func(env *game.Env) bool {
			env.State.AddOffer([]string{"build_speed", "deposit_speed"})
			return true
		},
	)
}

// houseDef implements game.StructureDef for the House structure.
type houseDef struct{}

// FoundationType returns the foundation tile type for House.
func (houseDef) FoundationType() game.StructureType { return FoundationHouse }

// BuiltType returns the built tile type for House.
func (houseDef) BuiltType() game.StructureType { return House }

// Footprint returns the 2×2 dimensions of a House.
func (houseDef) Footprint() (w, h int) { return 2, 2 }

// BuildCost returns the number of wood required to complete a House foundation.
func (houseDef) BuildCost() int { return houseBuildCost }

// ShouldSpawn is the world condition for spawning additional houses.
// Returns true when at least one house has been built and no house foundation is pending.
// The first house is handled by the story beat system; this drives all subsequent spawns.
func (houseDef) ShouldSpawn(env *game.Env) bool {
	built := len(env.State.World.StructureTypeIndex[House])
	pending := len(env.State.World.StructureTypeIndex[FoundationHouse])
	return built >= 1 && pending == 0
}

// UseSpawnAnchoredPlacement signals that the house foundation should be placed
// as close as possible to the world spawn point rather than near the player.
func (houseDef) UseSpawnAnchoredPlacement() bool { return true }

// OnBuilt is called when a House is completed.
// It marks the house as unoccupied; villager spawning is handled via XP milestone cards.
func (d houseDef) OnBuilt(env *game.Env, origin geom.Point) {
	env.State.HouseOccupancy[origin] = false
}

// OnPlayerInteraction handles adjacent-player interaction.
// When adjacent to a foundation, deposits one wood toward the build cost each cooldown tick.
// When adjacent to a built house, nothing happens (no storage).
func (d houseDef) OnPlayerInteraction(env *game.Env, origin geom.Point, now time.Time) {
	tile := env.State.World.TileAt(origin.X, origin.Y)
	if tile == nil || tile.Structure != FoundationHouse {
		return
	}
	p := env.State.Player
	if !p.CooldownExpired(game.Build, now) {
		return
	}
	if p.Inventory[game.Wood] == 0 {
		return
	}
	env.State.FoundationDeposited[origin]++
	p.Inventory[game.Wood]--
	p.QueueCooldown(game.Build, now.Add(p.BuildInterval))
	if env.State.FoundationDeposited[origin] >= d.BuildCost() {
		game.AwardXP(env, game.XPBuildCompletePlayer)
		game.FinalizeFoundation(env, d, origin)
	}
}
