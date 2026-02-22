package game

import (
	"math"
	"time"
)

// State holds all mutable game state.
type State struct {
	Player          *Player
	World           *World
	TotalWoodCut    int
	Building        *BuildOperation
	Storage         map[ResourceType]*ResourceStorage
	StorageByOrigin map[Point]*StorageInstance
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
			origin := Point{s.Building.X, s.Building.Y}
			s.World.IndexStructure(s.Building.X, s.Building.Y, s.Building.Width, s.Building.Height, def)
			def.OnBuilt(s, origin)
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

// TickAdjacentStructures calls OnPlayerInteraction once per structure instance
// that the player is cardinally adjacent to, then commits any pending cooldowns.
// Cooldowns are committed after all interactions so that multiple adjacent
// structures of the same type all fire within the same tick.
func (s *State) TickAdjacentStructures(now time.Time) {
	seen := make(map[Point]bool)
	for _, d := range [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
		p := Point{s.Player.X + d[0], s.Player.Y + d[1]}
		entry, ok := s.World.StructureIndex[p]
		if !ok || seen[entry.Origin] {
			continue
		}
		seen[entry.Origin] = true
		entry.Def.OnPlayerInteraction(s, entry.Origin, now)
	}
	s.Player.commitCooldowns()
}

// newState creates an initial game state with defaults.
func newState() *State {
	world := GenerateWorld(100, 100, DefaultSeed)
	player := NewPlayer(world.Width/2, world.Height/2)

	return &State{
		Player:          player,
		World:           world,
		Storage:         make(map[ResourceType]*ResourceStorage),
		StorageByOrigin: make(map[Point]*StorageInstance),
	}
}
