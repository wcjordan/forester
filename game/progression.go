package game

// HasStructureOfType returns true if any tile in the world has the given structure type.
func (s *State) HasStructureOfType(stype StructureType) bool {
	for y := range s.World.Tiles {
		for x := range s.World.Tiles[y] {
			if s.World.Tiles[y][x].Structure == stype {
				return true
			}
		}
	}
	return false
}

// ghostOriginFor returns the top-left corner of the current ghost footprint for the given type.
// ok is false if no ghost of that type exists.
func (s *State) ghostOriginFor(st StructureType) (x, y int, ok bool) {
	for row := range s.World.Tiles {
		for col := range s.World.Tiles[row] {
			if s.World.Tiles[row][col].Structure == st {
				return col, row, true
			}
		}
	}
	return 0, 0, false
}

// maybeSpawnGhosts checks each registered structure definition and places a ghost
// when its spawn condition is met and no ghost or built instance already exists.
func (s *State) maybeSpawnGhosts() {
	for _, def := range structures {
		if !def.ShouldSpawn(s) {
			continue
		}
		if s.HasStructureOfType(def.GhostType()) || s.HasStructureOfType(def.BuiltType()) {
			continue
		}
		w, h := def.Footprint()
		cx, cy := s.findValidLocation(w, h)
		if cx >= 0 {
			s.World.SetStructure(cx, cy, w, h, def.GhostType())
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
