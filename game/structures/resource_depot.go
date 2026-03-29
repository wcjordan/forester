package structures

import (
	"time"

	"forester/game"
	"forester/game/geom"
)

const resourceDepotBuildCost = 800
const resourceDepotCapacity = 2000
const resourceDepotTriggerHouses = 4

// FoundationResourceDepot and ResourceDepot are the StructureType values for the
// two stages of this structure.
const (
	FoundationResourceDepot game.StructureType = "foundation_resource_depot"
	ResourceDepot           game.StructureType = "resource_depot"
)

func init() {
	game.RegisterStructure(resourceDepotDef{}, game.StructureCallbacks{
		ShouldSpawn: func(_ *game.Env) bool { return false },
		OnBuilt: func(env *game.Env, origin geom.Point) {
			env.Stores.Register(origin, game.Wood, resourceDepotCapacity)
		},
		OnPlayerInteraction: resourceDepotOnPlayerInteraction,
	})
	game.RegisterVillagerDepositType(ResourceDepot)
	game.RegisterVillagerDeliveryType(FoundationResourceDepot)
	game.RegisterVillageCenterType(ResourceDepot)

	// Order 500: spawn the depot foundation once 4 houses are built and no depot exists.
	game.RegisterStoryBeat(500, "initial_resource_depot",
		func(env *game.Env) bool {
			houses := env.State.World.CountStructureInstances(House)
			hasDepot := env.State.World.HasStructureOfType(ResourceDepot) ||
				env.State.World.HasStructureOfType(FoundationResourceDepot)
			return houses >= resourceDepotTriggerHouses && !hasDepot
		},
		func(env *game.Env) bool {
			return game.SpawnFoundationByType(env, FoundationResourceDepot)
		},
	)

	// Order 600: queue the large carry capacity upgrade when the depot is first built.
	game.RegisterStoryBeat(600, "first_resource_depot_built",
		func(env *game.Env) bool {
			return env.State.World.HasStructureOfType(ResourceDepot)
		},
		func(env *game.Env) bool {
			env.State.AddOffer([]string{"large_carry_capacity"})
			return true
		},
	)
}

// resourceDepotDef implements game.StructureDef for the Resource Depot structure.
type resourceDepotDef struct{}

// FoundationType returns the foundation tile type for the Resource Depot.
func (resourceDepotDef) FoundationType() game.StructureType { return FoundationResourceDepot }

// BuiltType returns the built tile type for the Resource Depot.
func (resourceDepotDef) BuiltType() game.StructureType { return ResourceDepot }

// Footprint returns the 5×4 dimensions of the Resource Depot.
func (resourceDepotDef) Footprint() (w, h int) { return 5, 4 }

// BuildCost returns the number of wood required to complete a Resource Depot foundation.
func (resourceDepotDef) BuildCost() int { return resourceDepotBuildCost }

// UseSpawnAnchoredPlacement signals that the depot foundation should be placed
// near the world spawn point rather than near the player.
func (resourceDepotDef) UseSpawnAnchoredPlacement() bool { return true }

// StorageResource returns the resource type stored by the Resource Depot.
func (resourceDepotDef) StorageResource() game.ResourceType { return game.Wood }

// StorageCapacity returns the wood capacity of a single Resource Depot instance.
func (resourceDepotDef) StorageCapacity() int { return resourceDepotCapacity }

// resourceDepotOnPlayerInteraction handles adjacent-player interaction for both
// foundation and built states.
// When adjacent to a foundation, deposits one wood toward the build cost each cooldown tick.
// When adjacent to a built depot, deposits one wood into storage.
func resourceDepotOnPlayerInteraction(env *game.Env, origin geom.Point, now time.Time) {
	p := env.State.Player
	tile := env.State.World.TileAt(origin.X, origin.Y)
	if tile != nil && tile.Structure == FoundationResourceDepot {
		if !p.CooldownExpired(game.Build, now) {
			return
		}
		if p.Inventory[game.Wood] == 0 {
			return
		}
		env.State.FoundationDeposited[origin]++
		p.Inventory[game.Wood]--
		p.LastThrustAt = now
		p.QueueCooldown(game.Build, now.Add(p.BuildInterval))
		if env.State.FoundationDeposited[origin] >= resourceDepotBuildCost {
			game.AwardXP(env, game.XPBuildCompletePlayer)
			game.FinalizeFoundation(env, resourceDepotDef{}, origin)
		}
		return
	}

	if !p.CooldownExpired(game.Deposit, now) {
		return
	}
	if p.Inventory[game.Wood] == 0 {
		return
	}
	deposited := env.Stores.DepositAt(origin, 1)
	p.Inventory[game.Wood] -= deposited
	if deposited > 0 {
		p.LastThrustAt = now
		p.QueueCooldown(game.Deposit, now.Add(p.DepositInterval))
		game.AwardXP(env, game.XPPerWoodDeposited*deposited)
	}
}
