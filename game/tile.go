package game

// TerrainType represents the base terrain of a tile.
type TerrainType int

// Terrain types.
const (
	Grassland TerrainType = iota
	Forest
	Stump // fully harvested tree
)

// StructureType represents a structure placed on top of terrain.
type StructureType int

// Structure types.
const (
	NoStructure     StructureType = iota
	GhostLogStorage               // planned Log Storage footprint (walkable)
	LogStorage                    // built Log Storage (blocks movement)
)

// Tile represents a single cell in the world grid.
type Tile struct {
	Terrain   TerrainType
	TreeSize  int // Forest tiles only: wood remaining (1–10); 0 after harvest → Stump
	WalkCount int
	Structure StructureType
}
