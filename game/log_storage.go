package game

import "time"

// LogStorageBuildCost is the number of wood required to complete a Log Storage foundation.
const LogStorageBuildCost = 20

// LogStorageCapacity is the maximum number of wood a single Log Storage can hold.
const LogStorageCapacity = 100

// DepositTickInterval is how often the player auto-deposits one wood when adjacent to a storage structure.
const DepositTickInterval = 500 * time.Millisecond

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
func (logStorageDef) ShouldSpawn(s *State) bool {
	return s.Player.Wood >= MaxWood
}

// OnBuilt registers a new storage instance when a Log Storage is completed.
func (logStorageDef) OnBuilt(s *State, origin Point) {
	inst := s.getStorage(Wood).AddInstance(Wood, LogStorageCapacity)
	if s.StorageByOrigin == nil {
		s.StorageByOrigin = make(map[Point]*StorageInstance)
	}
	s.StorageByOrigin[origin] = inst
}

// OnPlayerInteraction handles adjacent-player interaction for both foundation and built states.
// When adjacent to a foundation, deposits one wood toward the build cost each cooldown tick.
// When adjacent to a built storage, deposits one wood into the storage instance.
func (d logStorageDef) OnPlayerInteraction(s *State, origin Point, now time.Time) {
	if !s.Player.CooldownExpired(Deposit, now) {
		return
	}
	if s.Player.Wood == 0 {
		return
	}

	tile := s.World.TileAt(origin.X, origin.Y)
	if tile != nil && tile.Structure == FoundationLogStorage {
		s.FoundationDeposited[origin]++
		s.Player.Wood--
		s.Player.QueueCooldown(Deposit, now.Add(DepositTickInterval))
		if s.FoundationDeposited[origin] >= d.BuildCost() {
			w, h := d.Footprint()
			s.World.SetStructure(origin.X, origin.Y, w, h, LogStorage)
			s.World.IndexStructure(origin.X, origin.Y, w, h, d)
			delete(s.FoundationDeposited, origin)
			d.OnBuilt(s, origin)
		}
		return
	}

	inst := s.StorageByOrigin[origin]
	if inst == nil {
		return
	}
	deposited := inst.Deposit(1)
	s.Player.Wood -= deposited
	if deposited > 0 {
		s.Player.QueueCooldown(Deposit, now.Add(DepositTickInterval))
	}
}
