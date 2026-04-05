package game

import (
	"math"
	"time"
)

// CooldownType identifies a named per-player interaction cooldown.
type CooldownType int

const (
	// Deposit is the cooldown type for depositing resources into a built storage structure.
	Deposit CooldownType = iota
	// Move is the cooldown type for player movement.
	Move
	// Build is the cooldown type for depositing resources into a foundation (building it up).
	Build
	// Harvest is the cooldown type for auto-harvesting adjacent trees.
	Harvest
)

// Player represents the player character.
type Player struct {
	X, Y               int     // current tile (always int(math.Floor(PosX/PosY)))
	PosX, PosY         float64 // continuous position in tile coordinates
	FacingDX, FacingDY int
	Inventory          map[ResourceType]int
	MaxCarry           int
	// BuildInterval controls how often the player can deposit one wood into a foundation.
	BuildInterval time.Duration
	// DepositInterval controls how often the player can auto-deposit one wood into built storage.
	DepositInterval time.Duration
	// HarvestInterval controls how often the player auto-harvests adjacent trees.
	HarvestInterval time.Duration
	// MoveSpeedMultiplier scales all movement cooldowns. Starts at 1.0; values below 1.0 are faster.
	MoveSpeedMultiplier float64
	Cooldowns           map[CooldownType]time.Time
	pendingCooldowns    map[CooldownType]time.Time
	// LastHarvestAt is the last time the player successfully harvested wood (harvest > 0).
	// Used by the render layer to trigger the slash animation.
	LastHarvestAt time.Time
	// LastThrustAt is the last time the player made a build deposit or resource deposit.
	// Used by the render layer to trigger the thrust animation.
	LastThrustAt time.Time
}

// NewPlayer creates a player at the given position, facing north.
func NewPlayer(x, y int) *Player {
	return &Player{
		X: x, Y: y, PosX: float64(x), PosY: float64(y), FacingDX: 0, FacingDY: -1,
		Inventory:           make(map[ResourceType]int),
		MaxCarry:            InitialCarryingCapacity,
		BuildInterval:       DepositTickInterval,
		DepositInterval:     DepositTickInterval,
		HarvestInterval:     harvestTickInterval,
		MoveSpeedMultiplier: 1.0,
		Cooldowns:           make(map[CooldownType]time.Time),
		pendingCooldowns:    make(map[CooldownType]time.Time),
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
	baseCooldown := defaultMoveCooldown
	if tile != nil {
		baseCooldown = MoveCooldownFor(tile)
	}
	cooldown := time.Duration(float64(baseCooldown) * p.MoveSpeedMultiplier)
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
	p.PosX = float64(p.X)
	p.PosY = float64(p.Y)
	if isRoadEligible(destTile) {
		destTile.WalkCount++
	}
}

// MoveSmooth moves the player continuously in direction (dx, dy) over duration dt.
// dx and dy are expected to be in {-1, 0, 1}. The player's PosX/PosY are updated
// to the new continuous position; X and Y are synced to int(math.Floor(PosX/PosY)).
// Collision is checked at tile boundaries: the player stops just before a blocked
// tile. WalkCount is incremented when entering a road-eligible tile.
// No-op when dx == 0 && dy == 0.
func (p *Player) MoveSmooth(dx, dy float64, w *World, dt time.Duration) {
	if dx == 0 && dy == 0 {
		return
	}
	if dx != 0 {
		p.FacingDX = int(math.Copysign(1, dx))
		p.FacingDY = 0
	}
	if dy != 0 {
		p.FacingDY = int(math.Copysign(1, dy))
		p.FacingDX = 0
	}

	curTile := w.TileAt(p.X, p.Y)
	baseCooldown := defaultMoveCooldown
	if curTile != nil {
		baseCooldown = MoveCooldownFor(curTile)
	}
	cooldown := time.Duration(float64(baseCooldown) * p.MoveSpeedMultiplier)
	speed := float64(time.Second) / float64(cooldown) // tiles/sec

	if dx != 0 {
		p.PosX = p.advancePosX(p.PosX+dx*speed*dt.Seconds(), dx, p.Y, w)
		p.X = int(math.Floor(p.PosX))
	}
	if dy != 0 {
		p.PosY = p.advancePosY(p.PosY+dy*speed*dt.Seconds(), dy, p.X, w)
		p.Y = int(math.Floor(p.PosY))
	}
}

// advancePosX returns the allowed new X position after attempting to move to newX.
// Checks only the immediately adjacent tile in the direction of movement (dx > 0 or dx < 0).
// If that tile is blocked or out of bounds, the position is clamped to the near side
// of that tile boundary. WalkCount is incremented when entering a road-eligible tile.
func (p *Player) advancePosX(newX, dx float64, tileY int, w *World) float64 {
	oldTileX := int(math.Floor(p.PosX))
	if dx > 0 {
		boundary := float64(oldTileX + 1)
		if newX < boundary {
			return newX // still within current tile
		}
		dest := w.TileAt(oldTileX+1, tileY)
		if !w.InBounds(oldTileX+1, tileY) || (dest != nil && dest.Structure != NoStructure) {
			return math.Nextafter(boundary, float64(oldTileX))
		}
		if dest != nil && isRoadEligible(dest) {
			dest.WalkCount++
		}
		return newX
	}
	// dx < 0
	boundary := float64(oldTileX)
	if newX >= boundary {
		return newX // still within current tile
	}
	dest := w.TileAt(oldTileX-1, tileY)
	if !w.InBounds(oldTileX-1, tileY) || (dest != nil && dest.Structure != NoStructure) {
		return boundary
	}
	if dest != nil && isRoadEligible(dest) {
		dest.WalkCount++
	}
	return newX
}

// advancePosY returns the allowed new Y position after attempting to move to newY.
// Symmetric with advancePosX.
func (p *Player) advancePosY(newY, dy float64, tileX int, w *World) float64 {
	oldTileY := int(math.Floor(p.PosY))
	if dy > 0 {
		boundary := float64(oldTileY + 1)
		if newY < boundary {
			return newY
		}
		dest := w.TileAt(tileX, oldTileY+1)
		if !w.InBounds(tileX, oldTileY+1) || (dest != nil && dest.Structure != NoStructure) {
			return math.Nextafter(boundary, float64(oldTileY))
		}
		if dest != nil && isRoadEligible(dest) {
			dest.WalkCount++
		}
		return newY
	}
	// dy < 0
	boundary := float64(oldTileY)
	if newY >= boundary {
		return newY
	}
	dest := w.TileAt(tileX, oldTileY-1)
	if !w.InBounds(tileX, oldTileY-1) || (dest != nil && dest.Structure != NoStructure) {
		return boundary
	}
	if dest != nil && isRoadEligible(dest) {
		dest.WalkCount++
	}
	return newY
}

// defaultMoveCooldown is the base time between moves on standard terrain (Grassland).
const defaultMoveCooldown = 150 * time.Millisecond

// troddenMoveCooldown is the time between moves on a trodden path tile.
const troddenMoveCooldown = 120 * time.Millisecond

// roadMoveCooldown is the time between moves on a road tile. This is also the
// minimum possible move cooldown, so World.MoveCost normalizes by this value to
// ensure all terrain costs are >= 1.0 (required for A* admissibility).
const roadMoveCooldown = 90 * time.Millisecond

// moveCooldowns defines the base time between moves per terrain type.
// Terrain types not present fall through to defaultMoveCooldown.
// Road-eligible tiles may override these values based on WalkCount; see MoveCooldownFor.
var moveCooldowns = map[TerrainType]time.Duration{
	Grassland: defaultMoveCooldown,
	Forest:    300 * time.Millisecond,
}

// MoveCooldownFor returns the move cooldown for the given tile.
// Road-eligible tiles use a shorter cooldown based on their traffic level.
// Forest with TreeSize=0 (cut tree) uses defaultMoveCooldown.
func MoveCooldownFor(tile *Tile) time.Duration {
	if tile.Terrain == Forest && tile.TreeSize == 0 {
		return defaultMoveCooldown
	}
	switch RoadLevelFor(tile) {
	case 2:
		return roadMoveCooldown
	case 1:
		return troddenMoveCooldown
	}
	if d, ok := moveCooldowns[tile.Terrain]; ok {
		return d
	}
	return defaultMoveCooldown
}

// InitialCarryingCapacity is the carrying capacity a new player starts with.
const InitialCarryingCapacity = 20

// DepositTickInterval is how often the player auto-deposits one wood when adjacent to a storage structure.
const DepositTickInterval = 100 * time.Millisecond

// GameTickInterval is the base cadence of the game loop (how often game.Tick is called).
const GameTickInterval = 100 * time.Millisecond

// harvestTickInterval is how often the player auto-harvests adjacent trees.
const harvestTickInterval = 100 * time.Millisecond

// PlayerSaveData holds the persistent fields of Player.
// Runtime-only fields (Cooldowns, pendingCooldowns, LastHarvestAt, LastThrustAt) are excluded.
type PlayerSaveData struct {
	X, Y                int
	PosX, PosY          float64
	FacingDX, FacingDY  int
	Inventory           map[ResourceType]int
	MaxCarry            int
	BuildInterval       time.Duration
	DepositInterval     time.Duration
	HarvestInterval     time.Duration
	MoveSpeedMultiplier float64
}

// SaveData returns a snapshot of the player's persistent state.
func (p *Player) SaveData() PlayerSaveData {
	return PlayerSaveData{
		X:                   p.X,
		Y:                   p.Y,
		PosX:                p.PosX,
		PosY:                p.PosY,
		FacingDX:            p.FacingDX,
		FacingDY:            p.FacingDY,
		Inventory:           copyMap(p.Inventory),
		MaxCarry:            p.MaxCarry,
		BuildInterval:       p.BuildInterval,
		DepositInterval:     p.DepositInterval,
		HarvestInterval:     p.HarvestInterval,
		MoveSpeedMultiplier: p.MoveSpeedMultiplier,
	}
}

// LoadFrom restores the player's persistent state from data.
// Runtime-only fields (Cooldowns, pendingCooldowns) are reset to empty maps.
func (p *Player) LoadFrom(data PlayerSaveData) {
	p.X = data.X
	p.Y = data.Y
	// PosX/PosY may be zero for saves written before continuous movement was added;
	// fall back to the integer tile position in that case.
	if data.PosX == 0 && data.PosY == 0 {
		p.PosX = float64(data.X)
		p.PosY = float64(data.Y)
	} else {
		p.PosX = data.PosX
		p.PosY = data.PosY
	}
	p.FacingDX = data.FacingDX
	p.FacingDY = data.FacingDY
	p.Inventory = copyMap(data.Inventory)
	p.MaxCarry = data.MaxCarry
	p.BuildInterval = data.BuildInterval
	p.DepositInterval = data.DepositInterval
	p.HarvestInterval = data.HarvestInterval
	p.MoveSpeedMultiplier = data.MoveSpeedMultiplier
	p.Cooldowns = make(map[CooldownType]time.Time)
	p.pendingCooldowns = make(map[CooldownType]time.Time)
}
