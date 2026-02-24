package game

import "time"

// CooldownType identifies a named per-player interaction cooldown.
type CooldownType int

const (
	// Deposit is the cooldown type for depositing resources into a storage structure.
	Deposit CooldownType = iota
	// Move is the cooldown type for player movement.
	Move
)

// Player represents the player character.
type Player struct {
	X, Y               int
	FacingDX, FacingDY int
	Wood               int
	MaxCarry           int
	Cooldowns          map[CooldownType]time.Time
	pendingCooldowns   map[CooldownType]time.Time
}

// NewPlayer creates a player at the given position, facing north.
func NewPlayer(x, y int) *Player {
	return &Player{
		X: x, Y: y, FacingDX: 0, FacingDY: -1,
		MaxCarry:         MaxCarryingCapacity,
		Cooldowns:        make(map[CooldownType]time.Time),
		pendingCooldowns: make(map[CooldownType]time.Time),
	}
}

// CooldownExpired reports whether the given cooldown type has expired (or was never set).
// Returns true when now is at or after the stored expiry time.
func (p *Player) CooldownExpired(ct CooldownType, now time.Time) bool {
	return !now.Before(p.Cooldowns[ct])
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

// Move attempts to move the player by (dx, dy).
// The move is skipped if the move cooldown has not elapsed since the last move attempt.
// Cooldown duration is based on the terrain of the tile the player currently stands on.
// Movement is blocked by world bounds and any tile that contains a structure.
// Updates the player's facing direction when the cooldown is satisfied.
func (p *Player) Move(dx, dy int, w *World, now time.Time) {
	tile := w.TileAt(p.X, p.Y)
	cooldown := DefaultMoveCooldown
	if tile != nil {
		cooldown = MoveCooldownFor(tile)
	}
	if !p.CooldownExpired(Move, now) {
		return
	}
	p.SetCooldown(Move, now.Add(cooldown))
	if dx != 0 || dy != 0 {
		p.FacingDX = dx
		p.FacingDY = dy
	}
	nx, ny := p.X+dx, p.Y+dy
	if !w.InBounds(nx, ny) {
		return
	}
	destTile := w.TileAt(nx, ny)
	if destTile != nil && destTile.Structure != NoStructure {
		return
	}
	p.X = nx
	p.Y = ny
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

// MaxCarryingCapacity is the maximum amount of wood the player can carry.
const MaxCarryingCapacity = 20

// harvestPerStep is how much wood is taken from each adjacent Forest tile per turn.
const harvestPerStep = 1

// HarvestTickInterval is how often the player automatically harvests without moving.
const HarvestTickInterval = 100 * time.Millisecond

// HarvestAdjacent harvests wood from the tile under the player and the three Forest tiles
// in front of the player: straight ahead and the two forward diagonals.
// Each tile loses harvestPerStep wood; when TreeSize reaches 0 it stays Forest (cut tree).
// The harvested wood is added to the player's inventory.
func (p *Player) HarvestAdjacent(w *World) {
	if p.Wood >= p.MaxCarry {
		return
	}
	dx, dy := p.FacingDX, p.FacingDY
	// Four tiles: under the player, straight ahead, diagonal-left, diagonal-right.
	targets := [4][2]int{
		{p.X, p.Y},
		{p.X + dx, p.Y + dy},
		{p.X + dx - dy, p.Y + dy + dx},
		{p.X + dx + dy, p.Y + dy - dx},
	}
	for _, coord := range targets {
		tile := w.TileAt(coord[0], coord[1])
		if tile == nil || tile.Terrain != Forest {
			continue
		}
		canTake := min(harvestPerStep, p.MaxCarry-p.Wood)
		harvest := min(canTake, tile.TreeSize)
		tile.TreeSize -= harvest
		p.Wood += harvest
	}
}
