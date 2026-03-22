package game

import "time"

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
	X, Y               int
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
		X: x, Y: y, FacingDX: 0, FacingDY: -1,
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
	if isRoadEligible(destTile) {
		destTile.WalkCount++
	}
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
