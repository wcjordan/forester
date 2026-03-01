package structures

import (
	"time"

	"forester/game"
	"forester/game/core"
	"forester/game/geom"
)

const logStorageBuildCost = 20
const logStorageCapacity = 500

// FoundationLogStorage and LogStorage are the StructureType values for the two
// stages of this structure. Defined here so external packages can reference them
// without importing all of game/ or editing a central enum.
const (
	FoundationLogStorage game.StructureType = "foundation_log_storage"
	LogStorage           game.StructureType = "log_storage"
)

func init() {
	game.RegisterStructure(logStorageDef{})
	game.RegisterVillagerDepositType(LogStorage)

	// Order 100: spawn the first log storage foundation when the player's inventory is full.
	game.RegisterStoryBeat(100, "initial_log_storage",
		func(env *game.Env) bool {
			p := env.State.Player
			return p.Inventory[game.Wood] >= p.MaxCarry
		},
		func(env *game.Env) bool {
			return game.SpawnFoundationByType(env, FoundationLogStorage)
		},
	)

	// Order 200: queue the carry upgrade offer when the first log storage is completed.
	game.RegisterStoryBeat(200, "first_log_storage_built",
		func(env *game.Env) bool {
			return env.State.World.HasStructureOfType(LogStorage)
		},
		func(env *game.Env) bool {
			env.State.AddOffer([]string{"carry_capacity"})
			return true
		},
	)
}

// logStorageDef implements game.StructureDef for the Log Storage structure.
type logStorageDef struct{}

// FoundationType returns the foundation tile type for Log Storage.
func (logStorageDef) FoundationType() game.StructureType { return FoundationLogStorage }

// BuiltType returns the built tile type for Log Storage.
func (logStorageDef) BuiltType() game.StructureType { return LogStorage }

// Footprint returns the 4×4 dimensions of a Log Storage.
func (logStorageDef) Footprint() (w, h int) { return 4, 4 }

// BuildCost returns the number of wood required to complete a Log Storage foundation.
func (logStorageDef) BuildCost() int { return logStorageBuildCost }

// StorageResource returns the resource type stored by a Log Storage.
func (logStorageDef) StorageResource() game.ResourceType { return game.Wood }

// StorageCapacity returns the capacity of a single Log Storage instance.
func (logStorageDef) StorageCapacity() int { return logStorageCapacity }

// ShouldSpawn returns false: the initial log storage is triggered by the story beat system.
func (logStorageDef) ShouldSpawn(_ core.StructureEnv) bool { return false }

// OnBuilt registers a new storage instance when a Log Storage is completed.
func (logStorageDef) OnBuilt(coreEnv core.StructureEnv, origin geom.Point) {
	env := coreEnv.(*game.Env)
	env.Stores.Register(origin, game.Wood, logStorageCapacity)
}

// OnPlayerInteraction handles adjacent-player interaction for both foundation and built states.
// When adjacent to a foundation, deposits one wood toward the build cost each cooldown tick.
// When adjacent to a built storage, deposits one wood into the storage instance.
func (d logStorageDef) OnPlayerInteraction(coreEnv core.StructureEnv, origin geom.Point, now time.Time) {
	env := coreEnv.(*game.Env)
	p := env.State.Player
	tile := env.State.World.TileAt(origin.X, origin.Y)
	if tile != nil && tile.Structure == FoundationLogStorage {
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
		p.QueueCooldown(game.Deposit, now.Add(p.DepositInterval))
		game.AwardXP(env, game.XPPerWoodDeposited*deposited)
	}
}
