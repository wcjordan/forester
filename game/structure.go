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

// FinalizeFoundation converts a completed foundation into its built structure type,
// updates the world index, clears the deposit record, and calls OnBuilt.
// Call this once the deposited amount has reached or exceeded BuildCost.
func FinalizeFoundation(env *Env, def StructureDef, origin Point) {
	w, h := def.Footprint()
	env.State.World.SetStructure(origin.X, origin.Y, w, h, def.BuiltType())
	env.State.World.IndexStructure(origin.X, origin.Y, w, h, def)
	delete(env.State.FoundationDeposited, origin)
	def.OnBuilt(env, origin)
}

// structures maps every known StructureType to its StructureDef.
// Each def is stored under two keys: its FoundationType and its BuiltType.
// Use IterateStructures to visit each def exactly once.
var structures = map[StructureType]StructureDef{}

// RegisterStructure adds a StructureDef to the global registry.
// Call this from an init() function in an external package (e.g. game/structures).
// Panics on nil or if either FoundationType or BuiltType is already registered.
func RegisterStructure(d StructureDef) {
	if d == nil {
		panic("RegisterStructure: def is nil")
	}
	if _, exists := structures[d.FoundationType()]; exists {
		panic("RegisterStructure: FoundationType already registered")
	}
	if _, exists := structures[d.BuiltType()]; exists {
		panic("RegisterStructure: BuiltType already registered")
	}
	structures[d.FoundationType()] = d
	structures[d.BuiltType()] = d
}

// IterateStructures calls fn once for each registered StructureDef.
// Because each def is stored under both its FoundationType and BuiltType keys,
// only the FoundationType entry is visited to avoid double-calling.
func IterateStructures(fn func(StructureDef)) {
	for stype, def := range structures {
		if stype == def.FoundationType() {
			fn(def)
		}
	}
}
