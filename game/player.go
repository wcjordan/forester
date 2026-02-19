package game

// Player represents the player character.
type Player struct {
	X, Y               int
	FacingDX, FacingDY int
	Wood               int
}

// NewPlayer creates a player at the given position, facing north.
func NewPlayer(x, y int) *Player {
	return &Player{X: x, Y: y, FacingDX: 0, FacingDY: -1}
}

// MovePlayer moves the player by (dx, dy), clamped to world bounds.
// Updates the player's facing direction whenever a non-zero direction is given.
func (p *Player) MovePlayer(dx, dy int, w *World) {
	if dx != 0 || dy != 0 {
		p.FacingDX = dx
		p.FacingDY = dy
	}
	nx := p.X + dx
	ny := p.Y + dy
	if w.InBounds(nx, ny) {
		p.X = nx
		p.Y = ny
	}
}

// harvestPerStep is how much wood is taken from each adjacent Forest tile per turn.
const harvestPerStep = 1

// HarvestAdjacent harvests wood from the Forest tile in the player's facing direction.
// The tile loses harvestPerStep wood; when TreeSize reaches 0 it becomes a Stump.
// The harvested wood is added to the player's inventory.
func (p *Player) HarvestAdjacent(w *World) {
	tile := w.TileAt(p.X+p.FacingDX, p.Y+p.FacingDY)
	if tile == nil || tile.Terrain != Forest {
		return
	}
	harvest := min(harvestPerStep, tile.TreeSize)
	tile.TreeSize -= harvest
	p.Wood += harvest
	if tile.TreeSize == 0 {
		tile.Terrain = Stump
	}
}
