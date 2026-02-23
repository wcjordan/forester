package game

import "time"

// State holds serializable game state (truth data).
// Derived runtime structures (e.g. StorageManager) live on Game.
type State struct {
	Player              *Player
	World               *World
	FoundationDeposited map[Point]int
}

// Harvest harvests adjacent trees without moving the player.
// Spawns a foundation when the spawn condition is met.
func (s *State) Harvest(env *Env) {
	s.Player.HarvestAdjacent(s.World)
	s.maybeSpawnFoundation(env)
}

// TickAdjacentStructures calls OnPlayerInteraction once per structure instance
// that the player is cardinally adjacent to, then commits any pending cooldowns.
// Cooldowns are committed after all interactions so that multiple adjacent
// structures of the same type all fire within the same tick.
func (s *State) TickAdjacentStructures(env *Env, now time.Time) {
	seen := make(map[Point]bool)
	for _, d := range [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
		p := Point{s.Player.X + d[0], s.Player.Y + d[1]}
		entry, ok := s.World.StructureIndex[p]
		if !ok || seen[entry.Origin] {
			continue
		}
		seen[entry.Origin] = true
		entry.Def.OnPlayerInteraction(env, entry.Origin, now)
	}
	s.Player.commitCooldowns()
}

// newState creates an initial game state with defaults.
func newState() *State {
	world := GenerateWorld(100, 100, DefaultSeed)
	player := NewPlayer(world.Width/2, world.Height/2)

	return &State{
		Player:              player,
		World:               world,
		FoundationDeposited: make(map[Point]int),
	}
}
