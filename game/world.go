package game

import (
	"time"

	"forester/game/geom"
)

// point is an internal alias for geom.Point, used as a map key for spatial indexes.
type point = geom.Point

// World represents the game map as a 2D grid of tiles.
type World struct {
	Width              int
	Height             int
	Tiles              [][]Tile
	structureIndex     map[point]structureEntry
	StructureTypeIndex map[StructureType]map[geom.Point]struct{}
	// structureInstanceIndex maps each StructureType to the set of origin Points
	// for all instances of that type. Maintained by PlaceFoundation/PlaceBuilt
	// so that CountStructureInstances is O(1).
	structureInstanceIndex map[StructureType]map[point]struct{}
	// NoGrowTiles is the set of tiles suppressed from tree regrowth because
	// they are within noGrowRadius of the spawn point or any structure.
	// Populated by NewWorld (spawn zone) and PlaceFoundation/PlaceBuilt (structures).
	NoGrowTiles    map[geom.Point]struct{}
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
		structureIndex:         make(map[point]structureEntry),
		StructureTypeIndex:     make(map[StructureType]map[point]struct{}),
		structureInstanceIndex: make(map[StructureType]map[point]struct{}),
		NoGrowTiles:            make(map[point]struct{}),
	}
	w.markNoGrowZoneRect(width/2, height/2, 1, 1)
	return w
}

// InBounds returns true if the given coordinates are within the world.
func (w *World) InBounds(x, y int) bool {
	return x >= 0 && x < w.Width && y >= 0 && y < w.Height
}

// IsBlocked returns true if the tile at (x, y) cannot be traversed.
// Returns true for out-of-bounds coordinates and tiles with a structure.
func (w *World) IsBlocked(x, y int) bool {
	t := w.TileAt(x, y)
	return t == nil || t.Structure != NoStructure
}

// MoveCost returns the movement cost to enter the tile at (x, y).
// Derived from MoveCooldownFor so pathfinding cost stays in sync with movement speed.
// Always >= 1.0 so the A* Manhattan heuristic in geom.FindPath remains admissible.
func (w *World) MoveCost(x, y int) float64 {
	t := w.TileAt(x, y)
	if t == nil {
		return 1
	}
	return float64(MoveCooldownFor(t)) / float64(defaultMoveCooldown)
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
					w.NoGrowTiles[point{X: tx, Y: ty}] = struct{}{}
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
	origin := point{X: x, Y: y}

	// Stamp tiles and maintain StructureTypeIndex.
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			pt := point{X: x + dx, Y: y + dy}
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
					w.StructureTypeIndex[stype] = make(map[point]struct{})
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

	// Record per-tile and per-instance entries for the new type, or remove
	// stale structureIndex entries when clearing.
	if stype != NoStructure {
		if w.structureInstanceIndex[stype] == nil {
			w.structureInstanceIndex[stype] = make(map[point]struct{})
		}
		w.structureInstanceIndex[stype][origin] = struct{}{}
		for dy := 0; dy < height; dy++ {
			for dx := 0; dx < width; dx++ {
				pt := point{X: x + dx, Y: y + dy}
				if w.TileAt(pt.X, pt.Y) == nil {
					continue
				}
				w.structureIndex[pt] = structureEntry{Def: def, Origin: origin}
			}
		}
	} else {
		for dy := 0; dy < height; dy++ {
			for dx := 0; dx < width; dx++ {
				pt := point{X: x + dx, Y: y + dy}
				if w.TileAt(pt.X, pt.Y) != nil {
					delete(w.structureIndex, pt)
				}
			}
		}
	}
}

// CountStructureInstances returns the number of distinct instances of stype.
// O(1) — maintained by PlaceFoundation and PlaceBuilt.
func (w *World) CountStructureInstances(stype StructureType) int {
	return len(w.structureInstanceIndex[stype])
}

// isHarvestable returns true if the tile at (x, y) can be harvested for wood.
// Returns true for Forest tiles w/ TreeSize > 0
func (w *World) isHarvestable(x, y int) bool {
	t := w.TileAt(x, y)
	return t != nil && t.Terrain == Forest && t.TreeSize > 0
}
