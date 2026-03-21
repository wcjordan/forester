package game

import (
	"time"

	"forester/game/geom"
)

// SaveGameData is the top-level serializable snapshot of the full game state.
type SaveGameData struct {
	Player              PlayerSaveData
	World               WorldSaveData
	Storage             StorageState
	Villagers           []VillagerSaveDatum
	FoundationDeposited map[geom.Point]int
	HouseOccupancy      map[geom.Point]bool
	XP                  int
	XPMilestoneIdx      int
	PendingOfferIDs     [][]string
	CompletedBeats      map[string]bool
}

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

// WorldSaveData holds the persistent world state.
// The Structure field of each tile is excluded — it is rebuilt on load by
// replaying Structures via PlaceFoundation / PlaceBuilt.
type WorldSaveData struct {
	Width, Height int
	Tiles         [][]TileSaveData     // terrain data only
	Structures    []StructureSaveDatum // one entry per structure instance
}

// TileSaveData holds the persistent fields of a Tile.
// Structure is excluded (derived from WorldSaveData.Structures on load).
type TileSaveData struct {
	Terrain   TerrainType
	TreeSize  int
	WalkCount int
}

// StructureSaveDatum records the origin and exact StructureType of one structure instance.
// Type may be a foundation type or a built type.
type StructureSaveDatum struct {
	Origin geom.Point
	Type   StructureType
}

// VillagerSaveDatum holds the persistent fields of a Villager.
// Runtime-only fields (moveCooldown, path, pathFailures) are excluded.
type VillagerSaveDatum struct {
	X, Y    int
	TargetX int
	TargetY int
	Wood    int
	Task    VillagerTask
}
