package game

import "time"

// StructureDef describes the behavior of one structure type.
// Each structure is registered in the structures slice (see log_storage.go etc.).
type StructureDef interface {
	FoundationType() StructureType
	BuiltType() StructureType
	Footprint() (w, h int)
	BuildCost() int
	// ShouldSpawn returns true when domain conditions are met (e.g. enough wood cut).
	// The generic spawn loop handles the "already placed" guard separately.
	ShouldSpawn(env *Env) bool
	// OnPlayerInteraction is called each tick the player is adjacent to this structure instance.
	// origin is the top-left corner of the specific instance being interacted with.
	// now is the current clock time; implementations use it to check and set cooldowns.
	OnPlayerInteraction(env *Env, origin Point, now time.Time)
	// OnBuilt is called once when the structure is completed.
	// origin is the top-left corner of the specific instance that was just built.
	OnBuilt(env *Env, origin Point)
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

// RegisterStructure adds a StructureDef to the global registry.
// Call this from an init() function in an external package (e.g. game/structures).
// Panics on nil or duplicate registration (same FoundationType+BuiltType pair).
func RegisterStructure(d StructureDef) {
	if d == nil {
		panic("RegisterStructure: def is nil")
	}
	for _, existing := range structures {
		if existing.FoundationType() == d.FoundationType() && existing.BuiltType() == d.BuiltType() {
			panic("RegisterStructure: duplicate registration")
		}
	}
	structures = append(structures, d)
}
