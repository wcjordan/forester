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
	// for all instances of that type. Maintained by PlaceFoundation/PlaceBuilt
	// so that CountStructureInstances is O(1).
	structureInstanceIndex map[StructureType]map[Point]struct{}
	// NoGrowTiles is the set of tiles suppressed from tree regrowth because
	// they are within noGrowRadius of the spawn point or any structure.
	// Populated by NewWorld (spawn zone) and PlaceFoundation/PlaceBuilt (structures).
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
// Called by NewWorld (spawn zone) and addStructure (structure footprint).
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

// PlaceFoundation places def as a foundation at (x, y), deriving the tile
// type from def.FoundationType() and the footprint from def.Footprint().
func (w *World) PlaceFoundation(x, y int, def StructureDef) {
	fw, fh := def.Footprint()
	w.addStructure(x, y, fw, fh, def.FoundationType(), def)
}

// PlaceBuilt places def as a completed structure at (x, y), deriving the tile
// type from def.BuiltType() and the footprint from def.Footprint().
func (w *World) PlaceBuilt(x, y int, def StructureDef) {
	fw, fh := def.Footprint()
	w.addStructure(x, y, fw, fh, def.BuiltType(), def)
}

// clearStructure removes the structure placed at (x, y), using def.Footprint()
// to determine which tiles to clear. Only used in tests.
func (w *World) clearStructure(x, y int, def StructureDef) {
	fw, fh := def.Footprint()
	w.addStructure(x, y, fw, fh, NoStructure, nil)
}

// addStructure is the underlying implementation used by PlaceFoundation,
// PlaceBuilt, and clearStructure. It stamps stype onto the tile grid, maintains
// all three structure indexes, and expands the NoGrowTiles zone.
// Callers outside this file should use PlaceFoundation or PlaceBuilt instead.
// Pass stype=NoStructure and def=nil only when clearing (via clearStructure).
func (w *World) addStructure(x, y, width, height int, stype StructureType, def StructureDef) {
	origin := Point{x, y}

	// Stamp tiles and maintain StructureTypeIndex.
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			pt := Point{x + dx, y + dy}
			tile := w.TileAt(pt.X, pt.Y)
			if tile == nil {
				continue
			}
			if old := tile.Structure; old != NoStructure {
				inner := w.StructureTypeIndex[old]
				delete(inner, pt)
				if len(inner) == 0 {
					delete(w.StructureTypeIndex, old)
				}
			}
			tile.Structure = stype
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

	// Remove old instance index entry for this origin (handles type transitions
	// and clearing).
	for st, origins := range w.structureInstanceIndex {
		if _, ok := origins[origin]; ok {
			delete(origins, origin)
			if len(origins) == 0 {
				delete(w.structureInstanceIndex, st)
			}
			break
		}
	}

	// Record per-tile and per-instance entries for the new type.
	if stype != NoStructure {
		if w.structureInstanceIndex[stype] == nil {
			w.structureInstanceIndex[stype] = make(map[Point]struct{})
		}
		w.structureInstanceIndex[stype][origin] = struct{}{}
		for dy := 0; dy < height; dy++ {
			for dx := 0; dx < width; dx++ {
				w.structureIndex[Point{x + dx, y + dy}] = structureEntry{Def: def, Origin: origin}
			}
		}
	}
}

// CountStructureInstances returns the number of distinct instances of stype.
// O(1) — maintained by PlaceFoundation and PlaceBuilt.
func (w *World) CountStructureInstances(stype StructureType) int {
	return len(w.structureInstanceIndex[stype])
}
