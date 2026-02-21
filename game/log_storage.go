package game

// LogStorageBuildTicks is the number of ticks (at 100ms each) to complete a Log Storage build (~3s).
const LogStorageBuildTicks = 30

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

// OnAdjacentTick deposits one wood into the storage when the player is adjacent.
func (logStorageDef) OnAdjacentTick(s *State) {
	if s.Player.Wood == 0 {
		return
	}
	s.Player.Wood--
	s.LogStorageDeposited++
}
