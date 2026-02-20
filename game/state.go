package game

// State holds all mutable game state.
type State struct {
	Player       *Player
	World        *World
	TotalWoodCut int
	Building     *BuildOperation
}

// Move moves the player and checks for ghost contact.
func (s *State) Move(dx, dy int) {
	s.Player.MovePlayer(dx, dy, s.World)
	s.checkGhostContact()
}

// checkGhostContact starts a build operation when the player steps onto a ghost tile.
func (s *State) checkGhostContact() {
	if s.Building != nil {
		return
	}
	tile := s.World.TileAt(s.Player.X, s.Player.Y)
	if tile == nil || tile.Structure != GhostLogStorage {
		return
	}
	gx, gy, ok := s.GhostOrigin()
	if !ok {
		return
	}
	s.Building = &BuildOperation{
		X: gx, Y: gy,
		Width: 4, Height: 4,
		Target:     LogStorage,
		TotalTicks: LogStorageBuildTicks,
	}
	// Nudge player to the nearest tile outside the 4×4 footprint.
	s.nudgePlayerOutside(gx, gy, 4, 4)
}

// nudgePlayerOutside moves the player to the closest in-bounds tile outside the rectangle.
func (s *State) nudgePlayerOutside(rx, ry, rw, rh int) {
	type candidate struct {
		x, y int
		dist int
	}
	best := candidate{dist: 1<<31 - 1}
	px, py := s.Player.X, s.Player.Y

	// Check one-tile border around the footprint.
	for dy := -1; dy <= rh; dy++ {
		for dx := -1; dx <= rw; dx++ {
			// Only consider tiles on the perimeter of the extended border.
			if dx >= 0 && dx < rw && dy >= 0 && dy < rh {
				continue // inside footprint
			}
			cx, cy := rx+dx, ry+dy
			if !s.World.InBounds(cx, cy) {
				continue
			}
			t := s.World.TileAt(cx, cy)
			if t == nil || t.Structure == LogStorage {
				continue
			}
			d := (cx-px)*(cx-px) + (cy-py)*(cy-py)
			if d < best.dist {
				best = candidate{cx, cy, d}
			}
		}
	}
	if best.dist < 1<<31-1 {
		s.Player.X = best.x
		s.Player.Y = best.y
	}
}

// AdvanceBuild increments the in-progress build and completes it when done.
func (s *State) AdvanceBuild() {
	if s.Building == nil {
		return
	}
	s.Building.ProgressTicks++
	if s.Building.Done() {
		s.World.SetStructure(s.Building.X, s.Building.Y, s.Building.Width, s.Building.Height, LogStorage)
		s.Building = nil
	}
}

// GhostOrigin returns the top-left corner of the current ghost footprint.
// ok is false if no ghost exists.
func (s *State) GhostOrigin() (x, y int, ok bool) {
	for row := range s.World.Tiles {
		for col := range s.World.Tiles[row] {
			if s.World.Tiles[row][col].Structure == GhostLogStorage {
				return col, row, true
			}
		}
	}
	return 0, 0, false
}

// Harvest harvests adjacent trees without moving the player.
// Tracks total wood cut and spawns a ghost structure when the threshold is reached.
func (s *State) Harvest() {
	before := s.Player.Wood
	s.Player.HarvestAdjacent(s.World)
	s.TotalWoodCut += s.Player.Wood - before
	s.maybeSpawnGhost()
}

// Regrow advances tree regrowth across the world.
func (s *State) Regrow() {
	s.World.Regrow()
}

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

// maybeSpawnGhost places a GhostLogStorage if the wood-cut threshold is met
// and no ghost or built Log Storage already exists.
// It searches from the player's position toward the world center for the
// first 4×4 all-grassland area.
func (s *State) maybeSpawnGhost() {
	const woodThreshold = 10
	if s.TotalWoodCut < woodThreshold {
		return
	}
	if s.HasStructureOfType(GhostLogStorage) || s.HasStructureOfType(LogStorage) {
		return
	}
	cx, cy := s.findGhostLocation()
	if cx >= 0 {
		s.World.SetStructure(cx, cy, 4, 4, GhostLogStorage)
	}
}

// findGhostLocation walks from the player position toward the world center,
// returning the top-left corner of the first valid 4×4 all-grassland area.
// Returns (-1, -1) if no valid location is found.
func (s *State) findGhostLocation() (x, y int) {
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
		if s.isValid4x4(tx, ty) {
			return tx, ty
		}
	}
	return -1, -1
}

// isValid4x4 returns true if the 4×4 area with top-left at (x, y) is entirely
// in-bounds, grassland terrain, and has no structure.
func (s *State) isValid4x4(x, y int) bool {
	for dy := 0; dy < 4; dy++ {
		for dx := 0; dx < 4; dx++ {
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

// newState creates an initial game state with defaults.
func newState() *State {
	world := GenerateWorld(100, 100, DefaultSeed)
	player := NewPlayer(world.Width/2, world.Height/2)

	return &State{
		Player: player,
		World:  world,
	}
}
