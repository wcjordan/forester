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

// maybeSpawnFoundation checks each registered structure definition and places a foundation
// when its spawn condition is met and no foundation or built instance already exists.
func (s *State) maybeSpawnFoundation(env *Env) {
	for _, def := range structures {
		if !def.ShouldSpawn(env) {
			continue
		}
		if s.HasStructureOfType(def.FoundationType()) || s.HasStructureOfType(def.BuiltType()) {
			continue
		}
		w, h := def.Footprint()
		var cx, cy int
		if sa, ok := def.(SpawnAnchoredPlacer); ok && sa.UseSpawnAnchoredPlacement() {
			cx, cy = s.findValidLocationNearSpawn(w, h)
		} else {
			cx, cy = s.findValidLocation(w, h)
		}
		if cx >= 0 {
			s.World.SetStructure(cx, cy, w, h, def.FoundationType())
			s.World.IndexStructure(cx, cy, w, h, def)
		}
	}
}

// findValidLocation walks from the player position toward the world center,
// returning the top-left corner of the first valid area of the given dimensions.
// Returns (-1, -1) if no valid location is found.
func (s *State) findValidLocation(w, h int) (x, y int) {
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

// findValidLocationNearSpawn returns the top-left corner of the w×h area that
// is closest (by Euclidean distance from footprint center to spawn) to the world
// spawn point, satisfying all placement constraints. Returns (-1, -1) if none found.
func (s *State) findValidLocationNearSpawn(w, h int) (x, y int) {
	spawnX := s.World.Width / 2
	spawnY := s.World.Height / 2

	type candidate struct {
		x, y  int
		dist2 float64
	}

	var candidates []candidate
	for cy := 0; cy+h <= s.World.Height; cy++ {
		for cx := 0; cx+w <= s.World.Width; cx++ {
			centerX := float64(cx) + float64(w)/2
			centerY := float64(cy) + float64(h)/2
			dx := centerX - float64(spawnX)
			dy := centerY - float64(spawnY)
			candidates = append(candidates, candidate{cx, cy, dx*dx + dy*dy})
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].dist2 != candidates[j].dist2 {
			return candidates[i].dist2 < candidates[j].dist2
		}
		if candidates[i].y != candidates[j].y {
			return candidates[i].y < candidates[j].y
		}
		return candidates[i].x < candidates[j].x
	})

	for _, c := range candidates {
		if s.isValidArea(c.x, c.y, w, h) {
			return c.x, c.y
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
