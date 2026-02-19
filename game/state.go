package game

// State holds all mutable game state.
type State struct {
	Player *Player
	World  *World
}

// Move moves the player and then harvests adjacent trees.
func (s *State) Move(dx, dy int) {
	s.Player.MovePlayer(dx, dy, s.World)
	s.Player.HarvestAdjacent(s.World)
}

// Harvest harvests adjacent trees without moving the player.
func (s *State) Harvest() {
	s.Player.HarvestAdjacent(s.World)
}

// Regrow advances tree regrowth across the world.
func (s *State) Regrow() {
	s.World.Regrow()
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
