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

// harvestPerStep is how much wood is taken from each adjacent Forest tile per turn.
const harvestPerStep = 1

// HarvestAdjacent harvests wood from all cardinal-adjacent Forest tiles.
// Each Forest neighbour loses harvestPerStep wood; when TreeSize reaches 0
// the tile becomes a Stump. The harvested wood is added to the player's inventory.
func (p *Player) HarvestAdjacent(w *World) {
	dirs := [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	for _, d := range dirs {
		tile := w.TileAt(p.X+d[0], p.Y+d[1])
		if tile == nil || tile.Terrain != Forest {
			continue
		}
		harvest := min(harvestPerStep, tile.TreeSize)
		tile.TreeSize -= harvest
		p.Wood += harvest
		if tile.TreeSize == 0 {
			tile.Terrain = Stump
		}
	}
}
