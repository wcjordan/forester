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
	// Build is the cooldown type for depositing resources into a foundation (building it up).
	Build
	// Harvest is the cooldown type for auto-harvesting adjacent trees.
	Harvest
)

// Player represents the player character.
type Player struct {
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
	// MoveSpeedMultiplier is the player's movement speed multiplier.
	// 1.0 = default; higher = faster. Only this multiplier is persisted, not the
	// base speed, so future changes to DefaultMoveSpeed affect all loaded saves.
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

// TileX returns the tile column the player currently occupies.
func (p *Player) TileX() int { return int(math.Floor(p.PosX)) }

// TileY returns the tile row the player currently occupies.
func (p *Player) TileY() int { return int(math.Floor(p.PosY)) }

// SetTilePos snaps the player to the given tile position. Use for test setup and
// save/load; during gameplay use MoveSmooth.
func (p *Player) SetTilePos(x, y int) {
	p.PosX = float64(x)
	p.PosY = float64(y)
}

// NewPlayer creates a player at the given position, facing north.
func NewPlayer(x, y int) *Player {
	return &Player{
		PosX: float64(x), PosY: float64(y), FacingDX: 0, FacingDY: -1,
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

	curTile := w.TileAt(p.TileX(), p.TileY())
	speed := DefaultMoveSpeed * p.MoveSpeedMultiplier * TerrainSpeedFor(curTile)

	if dx != 0 {
		p.PosX = advancePos1D(p.PosX, p.PosX+dx*speed*dt.Seconds(), dx, p.TileY(), true, w)
	}
	if dy != 0 {
		p.PosY = advancePos1D(p.PosY, p.PosY+dy*speed*dt.Seconds(), dy, p.TileX(), false, w)
	}
}

// advancePos1D returns the allowed new position along one axis after attempting to
// move from oldPos to newPos in direction dir (+1 or -1). All cells between
// oldPos and newPos are checked in order; the player stops just before the first
// blocked or out-of-bounds cell. WalkCount is incremented for each road-eligible
// cell entered. fixed is the coordinate on the perpendicular axis.
// isX controls the tile lookup order: true → TileAt(moving, fixed), false → TileAt(fixed, moving).
func advancePos1D(oldPos, newPos, dir float64, fixed int, isX bool, w *World) float64 {
	tileAt := func(moving int) *Tile {
		if isX {
			return w.TileAt(moving, fixed)
		}
		return w.TileAt(fixed, moving)
	}
	inBounds := func(moving int) bool {
		if isX {
			return w.InBounds(moving, fixed)
		}
		return w.InBounds(fixed, moving)
	}

	oldCell := int(math.Floor(oldPos))
	newCell := int(math.Floor(newPos))

	if dir > 0 {
		if newPos < float64(oldCell+1) {
			return newPos // still within current tile
		}
		for cell := oldCell + 1; cell <= newCell; cell++ {
			if !inBounds(cell) {
				return math.Nextafter(float64(cell), float64(cell-1))
			}
			dest := tileAt(cell)
			if dest != nil && dest.Structure != NoStructure {
				return math.Nextafter(float64(cell), float64(cell-1))
			}
			if dest != nil && isRoadEligible(dest) {
				dest.WalkCount++
			}
		}
		return newPos
	}
	// dir < 0
	if newPos >= float64(oldCell) {
		return newPos // still within current tile
	}
	for cell := oldCell - 1; cell >= newCell; cell-- {
		if !inBounds(cell) {
			return float64(cell + 1)
		}
		dest := tileAt(cell)
		if dest != nil && dest.Structure != NoStructure {
			return float64(cell + 1)
		}
		if dest != nil && isRoadEligible(dest) {
			dest.WalkCount++
		}
	}
	return newPos
}

// DefaultMoveSpeed is the player's base movement speed in tiles/sec on default terrain.
// Exported so renderers and tests can compute traversal durations.
const DefaultMoveSpeed = float64(time.Second) / float64(150*time.Millisecond) // ≈6.667 tiles/sec

// Terrain speed factors relative to Grassland (1.0).
const (
	forestSpeedFactor  = 0.5  // half speed through dense forest
	troddenSpeedFactor = 1.25 // 1.25× on trodden path
	roadSpeedFactor    = 1.65 // 1.65× on road
)

// TerrainSpeedFor returns the speed multiplier for the given tile relative to Grassland (1.0).
// Forest = 0.5, trodden path = 1.25, road = 1.65.
// nil tile returns 1.0.
func TerrainSpeedFor(tile *Tile) float64 {
	if tile == nil {
		return 1.0
	}
	if tile.Terrain == Forest && tile.TreeSize == 0 {
		return 1.0
	}
	switch RoadLevelFor(tile) {
	case 2:
		return roadSpeedFactor
	case 1:
		return troddenSpeedFactor
	}
	if tile.Terrain == Forest {
		return forestSpeedFactor
	}
	return 1.0
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
	// X, Y are kept for backward-compat reading of saves written before PosX/PosY
	// were introduced. New saves do not write these fields; LoadFrom falls back to
	// them only when PosX and PosY are both zero.
	X, Y               int
	PosX, PosY         float64
	FacingDX, FacingDY int
	Inventory          map[ResourceType]int
	MaxCarry           int
	BuildInterval      time.Duration
	DepositInterval    time.Duration
	HarvestInterval    time.Duration
	// MoveSpeedMultiplier is the canonical speed field (direct: 1.0=default, >1.0=faster).
	// Saves from origin/main wrote inverted values (lower=faster); LoadFrom converts
	// values < 1.0 to the direct semantics.
	MoveSpeedMultiplier float64
}

// SaveData returns a snapshot of the player's persistent state.
func (p *Player) SaveData() PlayerSaveData {
	return PlayerSaveData{
		// X/Y intentionally omitted — PosX/PosY are the canonical saved position.
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
	mult := data.MoveSpeedMultiplier
	if mult == 0 {
		mult = 1.0
	}
	// Values < 1.0 are from origin/main saves where MoveSpeedMultiplier was inverted
	// (lower = faster; e.g. 0.9 = 10% faster). Convert to direct semantics.
	if mult < 1.0 {
		p.MoveSpeedMultiplier = 1.0 / mult
	} else {
		p.MoveSpeedMultiplier = mult
	}
	p.Cooldowns = make(map[CooldownType]time.Time)
	p.pendingCooldowns = make(map[CooldownType]time.Time)
}
