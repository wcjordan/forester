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

// MovePlayer moves the player by (dx, dy), clamped to world bounds.
func (p *Player) MovePlayer(dx, dy int, w *World) {
	nx := p.X + dx
	ny := p.Y + dy
	if w.InBounds(nx, ny) {
		p.X = nx
		p.Y = ny
	}
}
