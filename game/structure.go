package game

// StructureDef describes the behavior of one structure type.
// Each structure is registered in the structures slice (see log_storage.go etc.).
type StructureDef interface {
	GhostType() StructureType
	BuiltType() StructureType
	Footprint() (w, h int)
	BuildTicks() int
	// ShouldSpawn returns true when domain conditions are met (e.g. enough wood cut).
	// The generic spawn loop handles the "already placed" guard separately.
	ShouldSpawn(s *State) bool
	// OnAdjacentTick is called each tick the player is adjacent to the built structure.
	OnAdjacentTick(s *State)
}

// structures is the registry of all known structure definitions.
// Each definition registers itself via init() in its own file.
var structures []StructureDef

// BuildOperation tracks an in-progress structure build.
type BuildOperation struct {
	X, Y          int
	Width, Height int
	Target        StructureType
	ProgressTicks int
	TotalTicks    int
}

// Progress returns build completion as a fraction in [0, 1].
func (b *BuildOperation) Progress() float64 {
	return float64(b.ProgressTicks) / float64(b.TotalTicks)
}

// Done returns true when the build is complete.
func (b *BuildOperation) Done() bool {
	return b.ProgressTicks >= b.TotalTicks
}
