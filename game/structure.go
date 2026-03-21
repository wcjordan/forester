package game

import (
	"fmt"
	"time"

	"forester/game/geom"
)

// StructureDef describes the static shape of a structure type — its tile types,
// footprint, and build cost. It has no dependency on *Env so that external
// packages (e.g. game/internal/gametest) can implement it without creating an
// import cycle.
type StructureDef interface {
	FoundationType() StructureType
	BuiltType() StructureType
	Footprint() (w, h int)
	BuildCost() int
}

// StructureCallbacks holds the runtime behaviors for a structure type that
// require access to *Env. All fields are optional; nil functions are no-ops.
// Register callbacks alongside a StructureDef via RegisterStructure.
type StructureCallbacks struct {
	// ShouldSpawn returns true when world conditions call for placing a new
	// foundation. The generic spawn loop handles the "already placed" guard.
	ShouldSpawn func(env *Env) bool
	// OnPlayerInteraction is called each tick the player is adjacent to an
	// instance. origin is the top-left corner; now is the current clock time.
	OnPlayerInteraction func(env *Env, origin geom.Point, now time.Time)
	// OnBuilt is called once when a foundation is completed.
	// origin is the top-left corner of the instance that was just finished.
	OnBuilt func(env *Env, origin geom.Point)
}

// registeredDef stores a StructureDef together with its runtime callbacks.
type registeredDef struct {
	Def StructureDef
	CB  StructureCallbacks
}

// structureEntry pairs a StructureDef with the origin (top-left corner) of the
// specific instance it belongs to.  Used as values in World.structureIndex.
type structureEntry struct {
	Def    StructureDef
	Origin point
}

// FinalizeFoundation converts a completed foundation into its built structure
// type, updates the world index, clears the deposit record, and calls OnBuilt.
// Call this once the deposited amount has reached or exceeded BuildCost.
func FinalizeFoundation(env *Env, def StructureDef, origin geom.Point) {
	env.State.World.PlaceBuilt(origin.X, origin.Y, def)
	delete(env.State.FoundationDeposited, origin)
	cb := lookupCallbacks(def.BuiltType())
	if cb.OnBuilt != nil {
		cb.OnBuilt(env, origin)
	}
}

// structures maps every known StructureType to its registeredDef.
// Each def is stored under two keys: its FoundationType and its BuiltType.
// Use IterateStructures to visit each def exactly once.
var structures = map[StructureType]registeredDef{}

// RegisterStructure adds a StructureDef and its optional callbacks to the
// global registry. Call this from an init() function in an external package
// (e.g. game/structures). Panics on nil def or if either key is already taken.
func RegisterStructure(def StructureDef, cb StructureCallbacks) {
	if def == nil {
		panic("RegisterStructure: def is nil")
	}
	if _, exists := structures[def.FoundationType()]; exists {
		panic(fmt.Sprintf("RegisterStructure: FoundationType %q already registered", def.FoundationType()))
	}
	if _, exists := structures[def.BuiltType()]; exists {
		panic(fmt.Sprintf("RegisterStructure: BuiltType %q already registered", def.BuiltType()))
	}
	entry := registeredDef{Def: def, CB: cb}
	structures[def.FoundationType()] = entry
	structures[def.BuiltType()] = entry
}

// IterateStructures calls fn once for each registered StructureDef and its
// callbacks. Only the FoundationType entry is visited to avoid double-calling.
// Iteration order is undefined.
func IterateStructures(fn func(StructureDef, StructureCallbacks)) {
	for stype, reg := range structures {
		if stype == reg.Def.FoundationType() {
			fn(reg.Def, reg.CB)
		}
	}
}

// lookupStructureDef returns the StructureDef for the given StructureType.
// Returns false if the type is not registered.
func lookupStructureDef(stype StructureType) (StructureDef, bool) {
	if reg, ok := structures[stype]; ok {
		return reg.Def, true
	}
	return nil, false
}

// lookupCallbacks returns the StructureCallbacks for the given StructureType.
// Returns a zero-value StructureCallbacks (all nil fields) if not found.
func lookupCallbacks(st StructureType) StructureCallbacks {
	if reg, ok := structures[st]; ok {
		return reg.CB
	}
	return StructureCallbacks{}
}
