package game

// Player represents the player character.
type Player struct {
	X, Y int
	Wood int
}

// NewPlayer creates a player at the given position.
func NewPlayer(x, y int) *Player {
	return &Player{X: x, Y: y}
}
