package game

import "time"

// LogStorageBuildTicks is the number of ticks (at 100ms each) to complete a Log Storage build (~3s).
const LogStorageBuildTicks = 30

// LogStorageCapacity is the maximum number of wood a single Log Storage can hold.
const LogStorageCapacity = 100

// DepositTickInterval is how often the player auto-deposits one wood when adjacent to a storage structure.
const DepositTickInterval = 500 * time.Millisecond

func init() { structures = append(structures, logStorageDef{}) }

// logStorageDef implements StructureDef for the Log Storage structure.
type logStorageDef struct{}

// GhostType returns the planned/ghost tile type for Log Storage.
func (logStorageDef) GhostType() StructureType { return GhostLogStorage }

// BuiltType returns the built tile type for Log Storage.
func (logStorageDef) BuiltType() StructureType { return LogStorage }

// Footprint returns the 4×4 dimensions of a Log Storage.
func (logStorageDef) Footprint() (w, h int) { return 4, 4 }

// BuildTicks returns how many ticks it takes to build a Log Storage.
func (logStorageDef) BuildTicks() int { return LogStorageBuildTicks }

// ShouldSpawn returns true once 10 wood has been cut.
func (logStorageDef) ShouldSpawn(s *State) bool {
	return s.TotalWoodCut >= 10
}

// OnBuilt registers a new storage instance when a Log Storage is completed.
func (logStorageDef) OnBuilt(s *State, origin Point) {
	inst := s.getStorage(Wood).AddInstance(Wood, LogStorageCapacity)
	if s.StorageByOrigin == nil {
		s.StorageByOrigin = make(map[Point]*StorageInstance)
	}
	s.StorageByOrigin[origin] = inst
}

// OnPlayerInteraction deposits one wood into the specific adjacent storage instance
// when the Deposit cooldown has expired.
func (logStorageDef) OnPlayerInteraction(s *State, origin Point, now time.Time) {
	if !s.Player.CooldownExpired(Deposit, now) {
		return
	}
	if s.Player.Wood == 0 {
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
