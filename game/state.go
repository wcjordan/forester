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

// FoundationProgress returns the build progress (0.0–1.0) of the first active foundation,
// and whether any foundation is in progress. Uses structureIndex to look up BuildCost.
func (s *State) FoundationProgress() (float64, bool) {
	for origin, deposited := range s.FoundationDeposited {
		entry, ok := s.World.structureIndex[origin]
		if !ok {
			continue
		}
		cost := entry.Def.BuildCost()
		if cost == 0 {
			continue
		}
		return float64(deposited) / float64(cost), true
	}
	return 0, false
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
