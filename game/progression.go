package game

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
		cx, cy := s.findValidLocation(w, h)
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

// isValidArea returns true if the w×h area with top-left at (x, y) is entirely
// in-bounds, grassland terrain, and has no structure.
func (s *State) isValidArea(x, y, w, h int) bool {
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			tile := s.World.TileAt(x+dx, y+dy)
			if tile == nil || tile.Terrain != Grassland || tile.Structure != NoStructure {
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
