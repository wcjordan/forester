package game

import "time"

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
	// OnPlayerInteraction is called each tick the player is adjacent to this structure instance.
	// origin is the top-left corner of the specific instance being interacted with.
	// now is the current clock time; implementations use it to check and set cooldowns.
	OnPlayerInteraction(s *State, origin Point, now time.Time)
	// OnBuilt is called once when the structure is completed.
	// origin is the top-left corner of the specific instance that was just built.
	OnBuilt(s *State, origin Point)
}

// StructureEntry pairs a StructureDef with the origin (top-left corner) of the
// specific instance it belongs to.  Used as values in World.StructureIndex.
type StructureEntry struct {
	Def    StructureDef
	Origin Point
}

// structures is the registry of all known structure definitions.
// Each definition registers itself via init() in its own file.
var structures []StructureDef

// findDefForStructureType returns the StructureDef whose BuiltType matches st, or nil.
func findDefForStructureType(st StructureType) StructureDef {
	for _, def := range structures {
		if def.BuiltType() == st {
			return def
		}
	}
	return nil
}

// findDefForGhostStructureType returns the StructureDef whose GhostType matches st, or nil.
func findDefForGhostStructureType(st StructureType) StructureDef {
	if st == NoStructure {
		return nil
	}
	for _, def := range structures {
		if def.GhostType() == st {
			return def
		}
	}
	return nil
}

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
