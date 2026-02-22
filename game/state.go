package game

import "time"

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
