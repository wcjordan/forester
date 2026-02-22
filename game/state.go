package game

import (
	"math"
)

// State holds all mutable game state.
type State struct {
	Player       *Player
	World        *World
	TotalWoodCut int
	Building     *BuildOperation
	Storage      map[ResourceType]*ResourceStorage
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
	if tile == nil {
		return
	}
	def := findDefForGhost(tile.Structure)
	if def == nil {
		return
	}
	gx, gy, ok := s.ghostOriginFor(def.GhostType())
	if !ok {
		return
	}
	w, h := def.Footprint()
	s.Building = &BuildOperation{
		X: gx, Y: gy,
		Width: w, Height: h,
		Target:     def.BuiltType(),
		TotalTicks: def.BuildTicks(),
	}
	s.nudgePlayerOutside(gx, gy, w, h)
}

// nudgePlayerOutside moves the player to the closest in-bounds tile outside the rectangle.
func (s *State) nudgePlayerOutside(rx, ry, rw, rh int) {
	type candidate struct {
		x, y int
		dist int
	}
	best := candidate{dist: math.MaxInt32}
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
	if best.dist < math.MaxInt32 {
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
		s.World.SetStructure(s.Building.X, s.Building.Y, s.Building.Width, s.Building.Height, s.Building.Target)
		if def := findDefForBuilt(s.Building.Target); def != nil {
			s.World.IndexStructure(s.Building.X, s.Building.Y, s.Building.Width, s.Building.Height, def)
			def.OnBuilt(s)
		}
		s.Building = nil
	}
}

// findDefForBuilt returns the StructureDef whose BuiltType matches st, or nil.
func findDefForBuilt(st StructureType) StructureDef {
	for _, def := range structures {
		if def.BuiltType() == st {
			return def
		}
	}
	return nil
}

// getStorage returns (creating if needed) the ResourceStorage for the given type.
func (s *State) getStorage(r ResourceType) *ResourceStorage {
	if s.Storage == nil {
		s.Storage = make(map[ResourceType]*ResourceStorage)
	}
	if s.Storage[r] == nil {
		s.Storage[r] = &ResourceStorage{}
	}
	return s.Storage[r]
}

// TotalStored returns the total stored amount for a resource type.
func (s *State) TotalStored(r ResourceType) int {
	if s.Storage[r] == nil {
		return 0
	}
	return s.Storage[r].Total()
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

// findDefForGhost returns the StructureDef whose GhostType matches st, or nil.
func findDefForGhost(st StructureType) StructureDef {
	if st == NoStructure {
		return nil
	}
	for _, def := range structures {
		if def.GhostType() == st {
			return def
		}
	}
	return nil
}

// Harvest harvests adjacent trees without moving the player.
// Tracks total wood cut and spawns a ghost structure when the threshold is reached.
func (s *State) Harvest() {
	before := s.Player.Wood
	s.Player.HarvestAdjacent(s.World)
	s.TotalWoodCut += s.Player.Wood - before
	s.maybeSpawnGhosts()
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

// TickAdjacentStructures calls OnPlayerInteraction once per structure instance
// that the player is cardinally adjacent to.
func (s *State) TickAdjacentStructures() {
	seen := make(map[Point]bool)
	for _, d := range [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
		p := Point{s.Player.X + d[0], s.Player.Y + d[1]}
		entry, ok := s.World.StructureIndex[p]
		if !ok || seen[entry.Origin] {
			continue
		}
		seen[entry.Origin] = true
		entry.Def.OnPlayerInteraction(s, entry.Origin)
	}
}

// newState creates an initial game state with defaults.
func newState() *State {
	world := GenerateWorld(100, 100, DefaultSeed)
	player := NewPlayer(world.Width/2, world.Height/2)

	return &State{
		Player:  player,
		World:   world,
		Storage: make(map[ResourceType]*ResourceStorage),
	}
}
