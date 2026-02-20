package game

import "time"

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

// DefaultMoveCooldown is the minimum time between moves on standard terrain.
const DefaultMoveCooldown = 150 * time.Millisecond

// MoveCooldowns defines the minimum time between moves per terrain type.
// Terrain types not present use DefaultMoveCooldown.
var MoveCooldowns = map[TerrainType]time.Duration{
	Grassland: 150 * time.Millisecond,
	Stump:     150 * time.Millisecond,
	Forest:    300 * time.Millisecond,
}

// MaxWood is the maximum amount of wood the player can carry.
const MaxWood = 20

// harvestPerStep is how much wood is taken from each adjacent Forest tile per turn.
const harvestPerStep = 1

// HarvestTickInterval is how often the player automatically harvests without moving.
const HarvestTickInterval = 100 * time.Millisecond

// HarvestAdjacent harvests wood from the three Forest tiles in front of the player:
// straight ahead and the two forward diagonals.
// Each tile loses harvestPerStep wood; when TreeSize reaches 0 it becomes a Stump.
// The harvested wood is added to the player's inventory.
func (p *Player) HarvestAdjacent(w *World) {
	if p.Wood >= MaxWood {
		return
	}
	dx, dy := p.FacingDX, p.FacingDY
	// Three tiles in the forward arc: straight, diagonal-left, diagonal-right.
	targets := [3][2]int{
		{p.X + dx, p.Y + dy},
		{p.X + dx - dy, p.Y + dy + dx},
		{p.X + dx + dy, p.Y + dy - dx},
	}
	for _, coord := range targets {
		tile := w.TileAt(coord[0], coord[1])
		if tile == nil || tile.Terrain != Forest {
			continue
		}
		canTake := min(harvestPerStep, MaxWood-p.Wood)
		harvest := min(canTake, tile.TreeSize)
		tile.TreeSize -= harvest
		p.Wood += harvest
		if tile.TreeSize == 0 {
			tile.Terrain = Stump
		}
	}
}
