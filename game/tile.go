package game

import "forester/game/core"

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
type StructureType = core.StructureType

// NoStructure is the zero value for StructureType, representing an empty tile.
const NoStructure = core.NoStructure

// WalkCountTrodden is the WalkCount threshold at which a Grassland tile becomes a trodden path.
const WalkCountTrodden = 20

// WalkCountRoad is the WalkCount threshold at which a trodden tile becomes a road.
const WalkCountRoad = 100

// Tile represents a single cell in the world grid.
type Tile struct {
	Terrain   TerrainType
	TreeSize  int // Forest tiles only: wood remaining (1–10); 0 = cut tree (visual stump)
	WalkCount int
	Structure StructureType
}

// isRoadEligible reports whether walk traffic on tile should be counted toward road formation.
// Only Grassland is eligible; Forest and future terrain types must opt in explicitly.
func isRoadEligible(tile *Tile) bool {
	return tile.Terrain == Grassland
}

// RoadLevelFor returns the road level for a tile: 0 = none, 1 = trodden, 2 = road.
// Returns 0 for ineligible terrain regardless of WalkCount.
func RoadLevelFor(tile *Tile) int {
	if !isRoadEligible(tile) {
		return 0
	}
	switch {
	case tile.WalkCount >= WalkCountRoad:
		return 2
	case tile.WalkCount >= WalkCountTrodden:
		return 1
	default:
		return 0
	}
}
