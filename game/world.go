package game

import (
	"math/rand"
	"time"
)

// RegrowthCooldown is how often the regrowth tick fires.
const RegrowthCooldown = 500 * time.Millisecond

// RegrowthOdds is the 1-in-N chance each eligible Forest tile grows per regrowth tick.
const RegrowthOdds = 40

// maxTreeSize is the maximum TreeSize a Forest tile can grow to.
const maxTreeSize = 10

// Point is a 2D coordinate used as a map key for spatial indexes.
type Point struct{ X, Y int }

// World represents the game map as a 2D grid of tiles.
type World struct {
	Width              int
	Height             int
	Tiles              [][]Tile
	StructureIndex     map[Point]StructureEntry
	StructureTypeIndex map[StructureType]map[Point]struct{}
	// NoGrowTiles is the set of tiles suppressed from tree regrowth because
	// they are within noGrowRadius of a structure. Updated by SetStructure.
	NoGrowTiles map[Point]struct{}
}

// NewWorld creates a world with the given dimensions, filled with grassland.
func NewWorld(width, height int) *World {
	tiles := make([][]Tile, height)
	for y := range tiles {
		tiles[y] = make([]Tile, width)
		for x := range tiles[y] {
			tiles[y][x] = Tile{Terrain: Grassland}
		}
	}

	return &World{
		Width:              width,
		Height:             height,
		Tiles:              tiles,
		StructureIndex:     make(map[Point]StructureEntry),
		StructureTypeIndex: make(map[StructureType]map[Point]struct{}),
		NoGrowTiles:        make(map[Point]struct{}),
	}
}

// InBounds returns true if the given coordinates are within the world.
func (w *World) InBounds(x, y int) bool {
	return x >= 0 && x < w.Width && y >= 0 && y < w.Height
}

// noGrowRadius is the Euclidean radius around the spawn point and any structure
// within which Forest tiles are suppressed from regrowing.
const noGrowRadius = 8

// markNoGrowZone adds all in-bounds tiles within noGrowRadius of (cx, cy) to
// w.NoGrowTiles. Called by SetStructure whenever a structure tile is placed.
func (w *World) markNoGrowZone(cx, cy int) {
	for dy := -noGrowRadius; dy <= noGrowRadius; dy++ {
		for dx := -noGrowRadius; dx <= noGrowRadius; dx++ {
			if dx*dx+dy*dy <= noGrowRadius*noGrowRadius {
				if w.InBounds(cx+dx, cy+dy) {
					w.NoGrowTiles[Point{cx + dx, cy + dy}] = struct{}{}
				}
			}
		}
	}
}

// Regrow advances tree regrowth probabilistically.
// Each eligible Forest tile (including TreeSize=0 cut trees) has a 1/RegrowthOdds chance to grow,
// unless it is within Euclidean distance noGrowRadius of the spawn point or any structure tile.
// Structure proximity is checked via the precomputed NoGrowTiles set.
func (w *World) Regrow(rng *rand.Rand) {
	cx, cy := w.Width/2, w.Height/2
	for y := range w.Tiles {
		for x := range w.Tiles[y] {
			tile := &w.Tiles[y][x]
			if tile.Terrain != Forest || tile.TreeSize >= maxTreeSize {
				continue
			}
			// Suppress growth near the spawn point.
			dx, dy := x-cx, y-cy
			if dx*dx+dy*dy <= noGrowRadius*noGrowRadius {
				continue
			}
			// Suppress growth near any structure tile (precomputed set).
			if _, blocked := w.NoGrowTiles[Point{x, y}]; blocked {
				continue
			}
			if rng.Intn(RegrowthOdds) == 0 {
				tile.TreeSize++
			}
		}
	}
}

// TileAt returns a pointer to the tile at the given coordinates.
// Returns nil if out of bounds.
func (w *World) TileAt(x, y int) *Tile {
	if !w.InBounds(x, y) {
		return nil
	}
	return &w.Tiles[y][x]
}

// SetStructure stamps a rectangle of tiles (x, y) to (x+width-1, y+height-1)
// with the given structure type. Out-of-bounds tiles are skipped.
// It also maintains StructureTypeIndex.
func (w *World) SetStructure(x, y, width, height int, stype StructureType) {
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			pt := Point{x + dx, y + dy}
			tile := w.TileAt(pt.X, pt.Y)
			if tile == nil {
				continue
			}
			// Remove old entry from type index.
			if old := tile.Structure; old != NoStructure {
				inner := w.StructureTypeIndex[old]
				delete(inner, pt)
				if len(inner) == 0 {
					delete(w.StructureTypeIndex, old)
				}
			}
			tile.Structure = stype
			// Add new entry to type index and expand the no-grow zone.
			if stype != NoStructure {
				if w.StructureTypeIndex[stype] == nil {
					w.StructureTypeIndex[stype] = make(map[Point]struct{})
				}
				w.StructureTypeIndex[stype][pt] = struct{}{}
				w.markNoGrowZone(pt.X, pt.Y)
			}
		}
	}
}

// IndexStructure records every tile in the w×h footprint at (x, y) in the
// StructureIndex, all sharing the same Origin so multi-tile instances can be
// deduplicated by callers.
func (w *World) IndexStructure(x, y, width, height int, def StructureDef) {
	origin := Point{x, y}
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			w.StructureIndex[Point{x + dx, y + dy}] = StructureEntry{Def: def, Origin: origin}
		}
	}
}
