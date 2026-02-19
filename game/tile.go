package game

// TerrainType represents the base terrain of a tile.
type TerrainType int

// Terrain types.
const (
	Grassland TerrainType = iota
	Forest
	Stump // fully harvested tree
)

// Tile represents a single cell in the world grid.
type Tile struct {
	Terrain   TerrainType
	TreeSize  int // Forest tiles only: wood remaining (1–10); 0 after harvest → Stump
	WalkCount int
}
