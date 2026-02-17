package game

// State holds all mutable game state.
type State struct {
	Player *Player
	World  *World
}

// newState creates an initial game state with defaults.
func newState() *State {
	world := NewWorld(100, 100)
	player := NewPlayer(world.Width/2, world.Height/2)

	return &State{
		Player: player,
		World:  world,
	}
}
