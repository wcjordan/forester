package game

// TerrainType represents the base terrain of a tile.
type TerrainType int

// Terrain types.
const (
	Grassland TerrainType = iota
	Forest
)

// StructureType represents a structure placed on top of terrain.
type StructureType int

// Structure types.
const (
	NoStructure          StructureType = iota
	FoundationLogStorage               // Log Storage foundation (blocks movement, accepts resource deposits)
	LogStorage                         // built Log Storage (blocks movement)
	FoundationHouse                    // House foundation (blocks movement, accepts resource deposits)
	House                              // built House (blocks movement)
)

// Tile represents a single cell in the world grid.
type Tile struct {
	Terrain   TerrainType
	TreeSize  int // Forest tiles only: wood remaining (1–10); 0 = cut tree (visual stump)
	WalkCount int
	Structure StructureType
}
