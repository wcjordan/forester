package structures

import (
	"time"

	"forester/game"
	"forester/game/geom"
)

const logStorageBuildCost = 20
const logStorageCapacity = 500

func init() { game.RegisterStructure(logStorageDef{}) }

// logStorageDef implements game.StructureDef for the Log Storage structure.
type logStorageDef struct{}

// FoundationType returns the foundation tile type for Log Storage.
func (logStorageDef) FoundationType() game.StructureType { return game.FoundationLogStorage }

// BuiltType returns the built tile type for Log Storage.
func (logStorageDef) BuiltType() game.StructureType { return game.LogStorage }

// Footprint returns the 4×4 dimensions of a Log Storage.
func (logStorageDef) Footprint() (w, h int) { return 4, 4 }

// BuildCost returns the number of wood required to complete a Log Storage foundation.
func (logStorageDef) BuildCost() int { return logStorageBuildCost }

// StorageResource returns the resource type stored by a Log Storage.
func (logStorageDef) StorageResource() game.ResourceType { return game.Wood }

// StorageCapacity returns the capacity of a single Log Storage instance.
func (logStorageDef) StorageCapacity() int { return logStorageCapacity }

// ShouldSpawn returns false: the initial log storage is triggered by the story beat system.
func (logStorageDef) ShouldSpawn(_ *game.Env) bool { return false }

// OnBuilt registers a new storage instance when a Log Storage is completed.
func (logStorageDef) OnBuilt(env *game.Env, origin geom.Point) {
	env.Stores.Register(origin, game.Wood, logStorageCapacity)
}

// OnPlayerInteraction handles adjacent-player interaction for both foundation and built states.
// When adjacent to a foundation, deposits one wood toward the build cost each cooldown tick.
// When adjacent to a built storage, deposits one wood into the storage instance.
func (d logStorageDef) OnPlayerInteraction(env *game.Env, origin geom.Point, now time.Time) {
	p := env.State.Player
	tile := env.State.World.TileAt(origin.X, origin.Y)
	if tile != nil && tile.Structure == game.FoundationLogStorage {
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
	}
}
