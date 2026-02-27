package game

import "time"

// Point is a 2D coordinate used as a map key for spatial indexes.
type Point struct{ X, Y int }

// World represents the game map as a 2D grid of tiles.
type World struct {
	Width              int
	Height             int
	Tiles              [][]Tile
	structureIndex     map[Point]structureEntry
	StructureTypeIndex map[StructureType]map[Point]struct{}
	// structureInstanceIndex maps each StructureType to the set of origin Points
	// for all instances of that type. Maintained by IndexStructure so that
	// CountStructureInstances is O(1).
	structureInstanceIndex map[StructureType]map[Point]struct{}
	// NoGrowTiles is the set of tiles suppressed from tree regrowth because
	// they are within noGrowRadius of the spawn point or any structure.
	// Populated by NewWorld (spawn zone) and SetStructure (structure zones).
	NoGrowTiles    map[Point]struct{}
	regrowCooldown time.Time
}

// HasStructureOfType returns true if any tile in the world has the given structure type.
func (w *World) HasStructureOfType(stype StructureType) bool {
	return len(w.StructureTypeIndex[stype]) > 0
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
		structureIndex:         make(map[Point]structureEntry),
		StructureTypeIndex:     make(map[StructureType]map[Point]struct{}),
		structureInstanceIndex: make(map[StructureType]map[Point]struct{}),
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

// AddStructure stamps stype onto the tile grid and records the def in the
// structure index. It is the combined form of SetStructure + IndexStructure
// and should be used whenever both need to be called together.
func (w *World) AddStructure(x, y, width, height int, stype StructureType, def StructureDef) {
	w.SetStructure(x, y, width, height, stype)
	w.IndexStructure(x, y, width, height, def)
}

// CountStructureInstances returns the number of distinct instances of stype.
// O(1) — maintained by IndexStructure.
func (w *World) CountStructureInstances(stype StructureType) int {
	return len(w.structureInstanceIndex[stype])
}

// IndexStructure records every tile in the w×h footprint at (x, y) in the
// structureIndex, all sharing the same Origin so multi-tile instances can be
// deduplicated by callers. Also maintains structureInstanceIndex.
func (w *World) IndexStructure(x, y, width, height int, def StructureDef) {
	origin := Point{x, y}

	// If this origin was previously indexed under a different type, remove it.
	for stype, origins := range w.structureInstanceIndex {
		if _, ok := origins[origin]; ok {
			delete(origins, origin)
			if len(origins) == 0 {
				delete(w.structureInstanceIndex, stype)
			}
			break
		}
	}

	// Add origin to the new type bucket (determined from the tile, which
	// SetStructure has already updated before IndexStructure is called).
	tile := w.TileAt(x, y)
	if tile != nil && tile.Structure != NoStructure {
		stype := tile.Structure
		if w.structureInstanceIndex[stype] == nil {
			w.structureInstanceIndex[stype] = make(map[Point]struct{})
		}
		w.structureInstanceIndex[stype][origin] = struct{}{}
	}

	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			w.structureIndex[Point{x + dx, y + dy}] = structureEntry{Def: def, Origin: origin}
		}
	}
}
