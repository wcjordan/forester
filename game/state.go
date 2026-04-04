package game

import "forester/game/geom"

// State holds serializable game state (truth data).
// Derived runtime structures (e.g. StorageManager, VillagerManager) live on Game.
type State struct {
	Player              *Player
	World               *World
	FoundationDeposited map[geom.Point]int
	// HouseOccupancy maps each built house origin to whether it has a villager assigned.
	HouseOccupancy map[geom.Point]bool
	// XP is the total experience the player has earned.
	XP int
	// XPMilestoneIdx is the index of the next XP milestone to award a card offer for.
	XPMilestoneIdx int
	// pendingOfferIDs stores each queued offer as a slice of upgrade IDs (strings),
	// keeping State serializable without embedding interface values.
	pendingOfferIDs [][]string
	// completedBeats records which story beats have already fired (keyed by beat ID).
	completedBeats map[string]bool
}

// AddOffer enqueues a card offer by its upgrade IDs.
func (s *State) AddOffer(ids []string) {
	if len(ids) == 0 {
		return
	}
	s.pendingOfferIDs = append(s.pendingOfferIDs, ids)
}

// FoundationProgress returns the build progress (0.0–1.0) of the first active
// foundation, and whether any foundation is in progress.
func (s *State) FoundationProgress() (progress float64, ok bool) {
	all := s.AllFoundationsProgress()
	if len(all) == 0 {
		return 0, false
	}
	return all[0].Progress, true
}

// FoundationInfo holds rendering data for one active foundation.
type FoundationInfo struct {
	Origin   geom.Point
	Width    int
	Height   int
	Progress float64 // 0.0–1.0
}

// AllFoundationsProgress returns FoundationInfo for every active foundation.
func (s *State) AllFoundationsProgress() []FoundationInfo {
	var result []FoundationInfo
	for origin, deposited := range s.FoundationDeposited {
		entry, ok := s.World.structureIndex[origin]
		if !ok {
			continue
		}
		cost := entry.Def.BuildCost()
		if cost == 0 {
			continue
		}
		w, h := entry.Def.Footprint()
		result = append(result, FoundationInfo{
			Origin:   origin,
			Width:    w,
			Height:   h,
			Progress: float64(deposited) / float64(cost),
		})
	}
	return result
}

// FoundationProgressAt returns the build progress (0.0–1.0) for the foundation
// that owns tile pt, or (0, false) if pt is not part of an active foundation.
func (s *State) FoundationProgressAt(pt geom.Point) (progress float64, ok bool) {
	entry, exists := s.World.structureIndex[pt]
	if !exists {
		return 0, false
	}
	deposited, active := s.FoundationDeposited[entry.Origin]
	if !active {
		return 0, false
	}
	cost := entry.Def.BuildCost()
	if cost == 0 {
		return 0, false
	}
	return float64(deposited) / float64(cost), true
}

// newState creates an initial game state with defaults.
func newState() *State {
	world := GenerateWorld(100, 100, defaultSeed)
	player := NewPlayer(world.Width/2, world.Height/2)

	return &State{
		Player:              player,
		World:               world,
		FoundationDeposited: make(map[geom.Point]int),
		HouseOccupancy:      make(map[geom.Point]bool),
		completedBeats:      make(map[string]bool),
	}
}
