package game

import (
	"time"

	"forester/game/geom"
)

// SaveData collects a full snapshot of the game state for persistence.
func (g *Game) SaveData() SaveGameData {
	p := g.State.Player
	playerData := PlayerSaveData{
		X:                   p.X,
		Y:                   p.Y,
		FacingDX:            p.FacingDX,
		FacingDY:            p.FacingDY,
		Inventory:           copyIntMap(p.Inventory),
		MaxCarry:            p.MaxCarry,
		BuildInterval:       p.BuildInterval,
		DepositInterval:     p.DepositInterval,
		HarvestInterval:     p.HarvestInterval,
		MoveSpeedMultiplier: p.MoveSpeedMultiplier,
	}

	w := g.State.World
	tiles := make([][]TileSaveData, w.Height)
	for y := 0; y < w.Height; y++ {
		tiles[y] = make([]TileSaveData, w.Width)
		for x := 0; x < w.Width; x++ {
			t := w.Tiles[y][x]
			tiles[y][x] = TileSaveData{
				Terrain:   t.Terrain,
				TreeSize:  t.TreeSize,
				WalkCount: t.WalkCount,
			}
		}
	}
	var structs []StructureSaveDatum
	for stype, origins := range w.structureInstanceIndex {
		for origin := range origins {
			structs = append(structs, StructureSaveDatum{Origin: origin, Type: stype})
		}
	}
	worldData := WorldSaveData{
		Width:      w.Width,
		Height:     w.Height,
		Tiles:      tiles,
		Structures: structs,
	}

	var villagers []VillagerSaveDatum
	for _, v := range g.Villagers.Villagers {
		villagers = append(villagers, VillagerSaveDatum{
			X: v.X, Y: v.Y,
			TargetX: v.TargetX, TargetY: v.TargetY,
			Wood: v.Wood,
			Task: v.Task,
		})
	}

	s := g.State
	return SaveGameData{
		Player:              playerData,
		World:               worldData,
		Storage:             g.Stores.SaveData(),
		Villagers:           villagers,
		FoundationDeposited: copyIntMap(s.FoundationDeposited),
		HouseOccupancy:      copyBoolMap(s.HouseOccupancy),
		XP:                  s.XP,
		XPMilestoneIdx:      s.XPMilestoneIdx,
		PendingOfferIDs:     copyStringSliceSlice(s.pendingOfferIDs),
		CompletedBeats:      copyBoolMap(s.completedBeats),
	}
}

func copyIntMap[K comparable](m map[K]int) map[K]int {
	out := make(map[K]int, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func copyBoolMap[K comparable](m map[K]bool) map[K]bool {
	out := make(map[K]bool, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func copyStringSliceSlice(s [][]string) [][]string {
	if s == nil {
		return nil
	}
	out := make([][]string, len(s))
	for i, inner := range s {
		cp := make([]string, len(inner))
		copy(cp, inner)
		out[i] = cp
	}
	return out
}

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
