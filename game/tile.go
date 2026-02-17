package game

// TerrainType represents the base terrain of a tile.
type TerrainType int

// Terrain types.
const (
	Grassland TerrainType = iota
	Forest
)

// Tile represents a single cell in the world grid.
type Tile struct {
	Terrain   TerrainType
	WalkCount int
}
