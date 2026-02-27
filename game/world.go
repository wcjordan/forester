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
	// StructureInstanceIndex maps each StructureType to the set of origin Points
	// for all instances of that type. Maintained by IndexStructure so that
	// CountStructureInstances is O(1).
	StructureInstanceIndex map[StructureType]map[Point]struct{}
	// NoGrowTiles is the set of tiles suppressed from tree regrowth because
	// they are within noGrowRadius of the spawn point or any structure.
	// Populated by NewWorld (spawn zone) and SetStructure (structure zones).
	NoGrowTiles    map[Point]struct{}
	regrowCooldown time.Time
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

	w := &World{
		Width:                  width,
		Height:                 height,
		Tiles:                  tiles,
		StructureIndex:         make(map[Point]StructureEntry),
		StructureTypeIndex:     make(map[StructureType]map[Point]struct{}),
		StructureInstanceIndex: make(map[StructureType]map[Point]struct{}),
		NoGrowTiles:            make(map[Point]struct{}),
	}
	w.markNoGrowZoneRect(width/2, height/2, 1, 1)
	return w
}

// InBounds returns true if the given coordinates are within the world.
func (w *World) InBounds(x, y int) bool {
	return x >= 0 && x < w.Width && y >= 0 && y < w.Height
}

// noGrowRadius is the Euclidean radius around the spawn point and any structure
// within which Forest tiles are suppressed from regrowing.
const noGrowRadius = 8

// markNoGrowZoneRect adds all in-bounds tiles within noGrowRadius of the
// rectangle (fx, fy)–(fx+fw-1, fy+fh-1) to w.NoGrowTiles. Distance is
// measured from each candidate tile to the nearest point inside the rectangle.
// Called by NewWorld (spawn zone) and SetStructure (structure footprint).
func (w *World) markNoGrowZoneRect(fx, fy, fw, fh int) {
	for ty := fy - noGrowRadius; ty <= fy+fh-1+noGrowRadius; ty++ {
		for tx := fx - noGrowRadius; tx <= fx+fw-1+noGrowRadius; tx++ {
			// Nearest point in footprint rectangle to (tx, ty).
			nx := min(max(tx, fx), fx+fw-1)
			ny := min(max(ty, fy), fy+fh-1)
			dx, dy := tx-nx, ty-ny
			if dx*dx+dy*dy <= noGrowRadius*noGrowRadius {
				if w.InBounds(tx, ty) {
					w.NoGrowTiles[Point{tx, ty}] = struct{}{}
				}
			}
		}
	}
}

// Regrow advances tree regrowth probabilistically if the regrowth cooldown has elapsed.
// Each eligible Forest tile (including TreeSize=0 cut trees) has a 1/RegrowthOdds chance to grow,
// unless it is in the precomputed NoGrowTiles set (within noGrowRadius of the spawn point or any structure tile).
func (w *World) Regrow(rng *rand.Rand, now time.Time) {
	if !now.After(w.regrowCooldown) {
		return
	}
	w.regrowCooldown = now.Add(RegrowthCooldown)
	for y := range w.Tiles {
		for x := range w.Tiles[y] {
			tile := &w.Tiles[y][x]
			if tile.Terrain != Forest || tile.TreeSize >= maxTreeSize {
				continue
			}
			if _, blocked := w.NoGrowTiles[Point{x, y}]; blocked {
				// Cut trees (TreeSize=0) in no-grow zones convert to Grassland
				// so the cleared area stays open for village growth.
				if tile.TreeSize == 0 {
					tile.Terrain = Grassland
				}
				continue
			}
			if rng.Intn(RegrowthOdds) == 0 {
				tile.TreeSize++
			}
		}
	}
}

// RegrowElapsed reports whether the regrowth cooldown has elapsed.
func (w *World) RegrowElapsed(now time.Time) bool {
	return now.After(w.regrowCooldown)
}

// MarkRegrowCooldown sets the next regrowth cooldown deadline.
func (w *World) MarkRegrowCooldown(d time.Duration, now time.Time) {
	w.regrowCooldown = now.Add(d)
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
// It also maintains StructureTypeIndex and expands the NoGrowTiles zone once
// for the entire footprint.
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
			// Add new entry to type index.
			if stype != NoStructure {
				if w.StructureTypeIndex[stype] == nil {
					w.StructureTypeIndex[stype] = make(map[Point]struct{})
				}
				w.StructureTypeIndex[stype][pt] = struct{}{}
			}
		}
	}
	// Expand the no-grow zone once for the entire footprint.
	if stype != NoStructure {
		w.markNoGrowZoneRect(x, y, width, height)
	}
}

// CountStructureInstances returns the number of distinct instances of stype.
// O(1) — maintained by IndexStructure.
func (w *World) CountStructureInstances(stype StructureType) int {
	return len(w.StructureInstanceIndex[stype])
}

// IndexStructure records every tile in the w×h footprint at (x, y) in the
// StructureIndex, all sharing the same Origin so multi-tile instances can be
// deduplicated by callers. Also maintains StructureInstanceIndex.
func (w *World) IndexStructure(x, y, width, height int, def StructureDef) {
	origin := Point{x, y}

	// If this origin was previously indexed under a different type, remove it.
	for stype, origins := range w.StructureInstanceIndex {
		if _, ok := origins[origin]; ok {
			delete(origins, origin)
			if len(origins) == 0 {
				delete(w.StructureInstanceIndex, stype)
			}
			break
		}
	}

	// Add origin to the new type bucket (determined from the tile, which
	// SetStructure has already updated before IndexStructure is called).
	tile := w.TileAt(x, y)
	if tile != nil && tile.Structure != NoStructure {
		stype := tile.Structure
		if w.StructureInstanceIndex[stype] == nil {
			w.StructureInstanceIndex[stype] = make(map[Point]struct{})
		}
		w.StructureInstanceIndex[stype][origin] = struct{}{}
	}

	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			w.StructureIndex[Point{x + dx, y + dy}] = StructureEntry{Def: def, Origin: origin}
		}
	}
}
