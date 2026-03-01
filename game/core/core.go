package core

// StructureType is a string identifier for a structure placed on a tile.
type StructureType string

// NoStructure is the zero value — an empty tile.
const NoStructure StructureType = ""

// StructureEnv is passed to StructureDef callbacks. The concrete type is
// *game.Env; implementations that need it should type-assert.
// Defined here so game/internal/gametest can implement StructureDef
// without importing the game package (which would create a cycle).
type StructureEnv interface{}
