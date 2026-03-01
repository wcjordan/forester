package game

// TerrainType represents the base terrain of a tile.
type TerrainType int

// Terrain types.
const (
	Grassland TerrainType = iota
	Forest
)

// StructureType represents a structure placed on top of terrain.
// It is a string so that external packages (e.g. game/structures) can define
// new types without editing this file.
type StructureType string

// NoStructure is the zero value for StructureType, representing an empty tile.
const NoStructure StructureType = ""

// Tile represents a single cell in the world grid.
type Tile struct {
	Terrain   TerrainType
	TreeSize  int // Forest tiles only: wood remaining (1–10); 0 = cut tree (visual stump)
	WalkCount int
	Structure StructureType
}
