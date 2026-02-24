package game

import "time"

// LogStorageBuildCost is the number of wood required to complete a Log Storage foundation.
const LogStorageBuildCost = 20

// LogStorageCapacity is the maximum number of wood a single Log Storage can hold.
const LogStorageCapacity = 500

// DepositTickInterval is how often the player auto-deposits one wood when adjacent to a storage structure.
const DepositTickInterval = 100 * time.Millisecond

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

// ShouldSpawn returns true when the player's inventory is full.
func (logStorageDef) ShouldSpawn(env *Env) bool {
	return env.State.Player.Wood >= MaxWood
}

// OnBuilt registers a new storage instance when a Log Storage is completed.
func (logStorageDef) OnBuilt(env *Env, origin Point) {
	env.Stores.Register(origin, Wood, LogStorageCapacity)
	env.State.AddOffer(CardOffer{carryCapacityUpgrade{}})
}

// OnPlayerInteraction handles adjacent-player interaction for both foundation and built states.
// When adjacent to a foundation, deposits one wood toward the build cost each cooldown tick.
// When adjacent to a built storage, deposits one wood into the storage instance.
func (d logStorageDef) OnPlayerInteraction(env *Env, origin Point, now time.Time) {
	if !env.State.Player.CooldownExpired(Deposit, now) {
		return
	}
	if env.State.Player.Wood == 0 {
		return
	}

	tile := env.State.World.TileAt(origin.X, origin.Y)
	if tile != nil && tile.Structure == FoundationLogStorage {
		env.State.FoundationDeposited[origin]++
		env.State.Player.Wood--
		env.State.Player.QueueCooldown(Deposit, now.Add(DepositTickInterval))
		if env.State.FoundationDeposited[origin] >= d.BuildCost() {
			w, h := d.Footprint()
			env.State.World.SetStructure(origin.X, origin.Y, w, h, LogStorage)
			env.State.World.IndexStructure(origin.X, origin.Y, w, h, d)
			delete(env.State.FoundationDeposited, origin)
			d.OnBuilt(env, origin)
		}
		return
	}

	deposited := env.Stores.DepositAt(origin, 1)
	env.State.Player.Wood -= deposited
	if deposited > 0 {
		env.State.Player.QueueCooldown(Deposit, now.Add(DepositTickInterval))
	}
}
