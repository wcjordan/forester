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

// Structure types.
const (
	NoStructure          StructureType = ""                       // empty tile
	FoundationLogStorage StructureType = "foundation_log_storage" // Log Storage foundation (blocks movement, accepts resource deposits)
	LogStorage           StructureType = "log_storage"            // built Log Storage (blocks movement)
	FoundationHouse      StructureType = "foundation_house"       // House foundation (blocks movement, accepts resource deposits)
	House                StructureType = "house"                  // built House (blocks movement)
)

// Tile represents a single cell in the world grid.
type Tile struct {
	Terrain   TerrainType
	TreeSize  int // Forest tiles only: wood remaining (1–10); 0 = cut tree (visual stump)
	WalkCount int
	Structure StructureType
}
