package game

import "time"

// LogStorageBuildCost is the number of wood required to complete a Log Storage foundation.
const LogStorageBuildCost = 20

// LogStorageCapacity is the maximum number of wood a single Log Storage can hold.
const LogStorageCapacity = 500

func init() { structures = append(structures, logStorageDef{}) }

// logStorageDef implements StructureDef for the Log Storage structure.
type logStorageDef struct{}

// FoundationType returns the foundation tile type for Log Storage.
func (logStorageDef) FoundationType() StructureType { return FoundationLogStorage }

// BuiltType returns the built tile type for Log Storage.
func (logStorageDef) BuiltType() StructureType { return LogStorage }

// Footprint returns the 4×4 dimensions of a Log Storage.
func (logStorageDef) Footprint() (w, h int) { return 4, 4 }

// BuildCost returns the number of wood required to complete a Log Storage foundation.
func (logStorageDef) BuildCost() int { return LogStorageBuildCost }

// StorageResource returns the resource type stored by a Log Storage.
func (logStorageDef) StorageResource() ResourceType { return Wood }

// StorageCapacity returns the capacity of a single Log Storage instance.
func (logStorageDef) StorageCapacity() int { return LogStorageCapacity }

// ShouldSpawn returns true when the player's inventory has enough wood.
func (logStorageDef) ShouldSpawn(env *Env) bool {
	return env.State.Player.Wood >= LogStorageBuildCost
}

// OnBuilt registers a new storage instance when a Log Storage is completed.
func (logStorageDef) OnBuilt(env *Env, origin Point) {
	env.Stores.Register(origin, Wood, LogStorageCapacity)
	env.State.AddOffer([]string{"carry_capacity"})
}

// OnPlayerInteraction handles adjacent-player interaction for both foundation and built states.
// When adjacent to a foundation, deposits one wood toward the build cost each cooldown tick.
// When adjacent to a built storage, deposits one wood into the storage instance.
func (d logStorageDef) OnPlayerInteraction(env *Env, origin Point, now time.Time) {
	p := env.State.Player
	tile := env.State.World.TileAt(origin.X, origin.Y)
	if tile != nil && tile.Structure == FoundationLogStorage {
		if !p.CooldownExpired(Build, now) {
			return
		}
		if p.Wood == 0 {
			return
		}
		env.State.FoundationDeposited[origin]++
		p.Wood--
		p.QueueCooldown(Build, now.Add(p.BuildInterval))
		if env.State.FoundationDeposited[origin] >= d.BuildCost() {
			w, h := d.Footprint()
			env.State.World.SetStructure(origin.X, origin.Y, w, h, LogStorage)
			env.State.World.IndexStructure(origin.X, origin.Y, w, h, d)
			delete(env.State.FoundationDeposited, origin)
			d.OnBuilt(env, origin)
		}
		return
	}

	if !p.CooldownExpired(Deposit, now) {
		return
	}
	if p.Wood == 0 {
		return
	}
	deposited := env.Stores.DepositAt(origin, 1)
	p.Wood -= deposited
	if deposited > 0 {
		p.QueueCooldown(Deposit, now.Add(p.DepositInterval))
	}
}
