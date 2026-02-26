package game

import "sort"

// SpawnAnchoredPlacer is an optional interface for StructureDef implementations
// that want to be placed as close as possible to the world spawn point rather
// than on the path from the player toward the center.
type SpawnAnchoredPlacer interface {
	UseSpawnAnchoredPlacement() bool
}

// HasStructureOfType returns true if any tile in the world has the given structure type.
func (s *State) HasStructureOfType(stype StructureType) bool {
	return len(s.World.StructureTypeIndex[stype]) > 0
}

// findStructureDefByFoundationType returns the first registered StructureDef whose
// FoundationType matches the given type, or nil if none is found.
func findStructureDefByFoundationType(ft StructureType) StructureDef {
	for _, def := range structures {
		if def.FoundationType() == ft {
			return def
		}
	}
	return nil
}

// spawnFoundationAt finds a valid location for def and places its foundation tile.
// Placement is near the world spawn point if def implements SpawnAnchoredPlacer,
// otherwise near the player. Does nothing if no valid location is found.
func (s *State) spawnFoundationAt(def StructureDef) {
	w, h := def.Footprint()
	var cx, cy int
	if sa, ok := def.(SpawnAnchoredPlacer); ok && sa.UseSpawnAnchoredPlacement() {
		cx, cy = s.findValidLocationNearSpawn(w, h)
	} else {
		cx, cy = s.findValidLocationNearPlayer(w, h)
	}
	if cx >= 0 {
		s.World.SetStructure(cx, cy, w, h, def.FoundationType())
		s.World.IndexStructure(cx, cy, w, h, def)
	}
}

// maybeSpawnFoundation checks each registered structure definition and places a foundation
// when its ShouldSpawn world condition is met. Each ShouldSpawn implementation is responsible
// for its own idempotency (e.g. checking that no foundation is already pending).
func (s *State) maybeSpawnFoundation(env *Env) {
	for _, def := range structures {
		if !def.ShouldSpawn(env) {
			continue
		}
		s.spawnFoundationAt(def)
	}
}

// findValidLocationNearPlayer walks from the player position toward the world center,
// returning the top-left corner of the first valid area of the given dimensions.
// Returns (-1, -1) if no valid location is found.
func (s *State) findValidLocationNearPlayer(w, h int) (x, y int) {
	px, py := s.Player.X, s.Player.Y
	spawnX := s.World.Width / 2
	spawnY := s.World.Height / 2

	dx := spawnX - px
	dy := spawnY - py
	steps := abs(dx)
	if abs(dy) > steps {
		steps = abs(dy)
	}
	if steps == 0 {
		steps = 1
	}

	for i := 0; i <= steps; i++ {
		tx := px + dx*i/steps
		ty := py + dy*i/steps
		if s.isValidArea(tx, ty, w, h) {
			return tx, ty
		}
	}
	return -1, -1
}

// findValidLocationNearSpawn searches outward from the world spawn point in
// expanding Chebyshev rings, returning the top-left corner of the closest valid
// w×h area by Euclidean distance from footprint center to spawn.
// Stops as soon as the first valid location is found. Returns (-1, -1) if none found.
func (s *State) findValidLocationNearSpawn(w, h int) (x, y int) {
	spawnX := s.World.Width / 2
	spawnY := s.World.Height / 2
	// anchorX/anchorY is the top-left that would center the footprint on spawn.
	anchorX := spawnX - w/2
	anchorY := spawnY - h/2

	footprintDist2 := func(px, py int) float64 {
		cx := float64(px) + float64(w)/2 - float64(spawnX)
		cy := float64(py) + float64(h)/2 - float64(spawnY)
		return cx*cx + cy*cy
	}

	type pos struct{ x, y int }
	maxR := s.World.Width + s.World.Height

	for r := 0; r <= maxR; r++ {
		// Collect the perimeter of the Chebyshev ring at distance r from anchor.
		var ring []pos
		if r == 0 {
			ring = []pos{{anchorX, anchorY}}
		} else {
			for dx := -r; dx <= r; dx++ {
				ring = append(ring, pos{anchorX + dx, anchorY - r})
				ring = append(ring, pos{anchorX + dx, anchorY + r})
			}
			for dy := -r + 1; dy <= r-1; dy++ {
				ring = append(ring, pos{anchorX - r, anchorY + dy})
				ring = append(ring, pos{anchorX + r, anchorY + dy})
			}
		}

		// Sort this ring by Euclidean distance so we check the closest positions first.
		sort.Slice(ring, func(i, j int) bool {
			di := footprintDist2(ring[i].x, ring[i].y)
			dj := footprintDist2(ring[j].x, ring[j].y)
			if di != dj {
				return di < dj
			}
			if ring[i].y != ring[j].y {
				return ring[i].y < ring[j].y
			}
			return ring[i].x < ring[j].x
		})

		for _, p := range ring {
			if s.isValidArea(p.x, p.y, w, h) {
				return p.x, p.y
			}
		}
	}
	return -1, -1
}

// isValidArea returns true if the w×h area with top-left at (x, y) satisfies
// all placement constraints:
//   - Every footprint tile is in-bounds, grassland, and has no structure.
//   - No footprint tile overlaps the player's current position.
//   - The full 1-tile Chebyshev border around the footprint contains no structures
//     (ensures at least a 1-tile gap from every existing structure in all 8 directions).
func (s *State) isValidArea(x, y, w, h int) bool {
	px, py := s.Player.X, s.Player.Y

	// Check footprint tiles.
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			tx, ty := x+dx, y+dy
			// Player overlap check.
			if tx == px && ty == py {
				return false
			}
			tile := s.World.TileAt(tx, ty)
			if tile == nil || tile.Terrain != Grassland || tile.Structure != NoStructure {
				return false
			}
		}
	}

	// Check 1-tile Chebyshev border for existing structures.
	for by := y - 1; by <= y+h; by++ {
		for bx := x - 1; bx <= x+w; bx++ {
			// Skip the footprint itself (already checked above).
			if bx >= x && bx < x+w && by >= y && by < y+h {
				continue
			}
			tile := s.World.TileAt(bx, by)
			if tile != nil && tile.Structure != NoStructure {
				return false
			}
		}
	}

	return true
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
