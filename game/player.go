package game

import "time"

// CooldownType identifies a named per-player interaction cooldown.
type CooldownType int

const (
	// Deposit is the cooldown type for depositing resources into a storage structure.
	Deposit CooldownType = iota
)

// Player represents the player character.
type Player struct {
	X, Y               int
	FacingDX, FacingDY int
	Wood               int
	Cooldowns          map[CooldownType]time.Time
	pendingCooldowns   map[CooldownType]time.Time
}

// NewPlayer creates a player at the given position, facing north.
func NewPlayer(x, y int) *Player {
	return &Player{
		X: x, Y: y, FacingDX: 0, FacingDY: -1,
		Cooldowns:        make(map[CooldownType]time.Time),
		pendingCooldowns: make(map[CooldownType]time.Time),
	}
}

// CooldownExpired reports whether the given cooldown type has expired (or was never set).
func (p *Player) CooldownExpired(ct CooldownType, now time.Time) bool {
	return now.After(p.Cooldowns[ct])
}

// SetCooldown immediately sets the given cooldown type. Use for direct state
// manipulation (e.g. tests). Structure interactions should use QueueCooldown instead.
func (p *Player) SetCooldown(ct CooldownType, until time.Time) {
	p.Cooldowns[ct] = until
}

// QueueCooldown schedules a cooldown to be applied after the current interaction
// tick completes. Call this from OnPlayerInteraction so that all adjacent
// structure interactions within the same tick see the pre-tick cooldown state.
func (p *Player) QueueCooldown(ct CooldownType, until time.Time) {
	p.pendingCooldowns[ct] = until
}

// commitCooldowns applies all pending cooldowns and clears the pending set.
// Called by TickAdjacentStructures after all interactions have been processed.
func (p *Player) commitCooldowns() {
	for ct, until := range p.pendingCooldowns {
		p.Cooldowns[ct] = until
	}
	clear(p.pendingCooldowns)
}

// MovePlayer moves the player by (dx, dy), clamped to world bounds.
// Both LogStorage and FoundationLogStorage block movement.
// Updates the player's facing direction whenever a non-zero direction is given.
func (p *Player) MovePlayer(dx, dy int, w *World) {
	if dx != 0 || dy != 0 {
		p.FacingDX = dx
		p.FacingDY = dy
	}
	nx := p.X + dx
	ny := p.Y + dy
	if w.InBounds(nx, ny) {
		tile := w.TileAt(nx, ny)
		if tile == nil || (tile.Structure != LogStorage && tile.Structure != FoundationLogStorage) {
			p.X = nx
			p.Y = ny
		}
	}
}

// DefaultMoveCooldown is the minimum time between moves on standard terrain.
const DefaultMoveCooldown = 150 * time.Millisecond

// MoveCooldowns defines the minimum time between moves per terrain type.
// Terrain types not present use DefaultMoveCooldown.
var MoveCooldowns = map[TerrainType]time.Duration{
	Grassland: 150 * time.Millisecond,
	Forest:    300 * time.Millisecond,
}

// MoveCooldownFor returns the move cooldown for the given tile.
// Forest with TreeSize=0 (cut tree) uses DefaultMoveCooldown.
func MoveCooldownFor(tile *Tile) time.Duration {
	if tile.Terrain == Forest && tile.TreeSize == 0 {
		return DefaultMoveCooldown
	}
	if d, ok := MoveCooldowns[tile.Terrain]; ok {
		return d
	}
	return DefaultMoveCooldown
}

// MaxWood is the maximum amount of wood the player can carry.
const MaxWood = 20

// harvestPerStep is how much wood is taken from each adjacent Forest tile per turn.
const harvestPerStep = 1

// HarvestTickInterval is how often the player automatically harvests without moving.
const HarvestTickInterval = 100 * time.Millisecond

// HarvestAdjacent harvests wood from the three Forest tiles in front of the player:
// straight ahead and the two forward diagonals.
// Each tile loses harvestPerStep wood; when TreeSize reaches 0 it stays Forest (cut tree).
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
	}
}
