package game

// GameState holds all mutable game state.
type GameState struct {
	Player *Player
	World  *World
}

// newGameState creates an initial game state with defaults.
func newGameState() *GameState {
	world := NewWorld(100, 100)
	player := NewPlayer(world.Width/2, world.Height/2)

	return &GameState{
		Player: player,
		World:  world,
	}
}
