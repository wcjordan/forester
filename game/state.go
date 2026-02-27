package game

// State holds serializable game state (truth data).
// Derived runtime structures (e.g. StorageManager, VillagerManager) live on Game.
type State struct {
	Player              *Player
	World               *World
	FoundationDeposited map[Point]int
	// PendingOfferIDs stores each queued offer as a slice of upgrade IDs (strings),
	// keeping State serializable without embedding interface values.
	PendingOfferIDs [][]string
	// CompletedBeats records which story beats have already fired (keyed by beat ID).
	CompletedBeats map[string]bool
}

// AddOffer enqueues a card offer by its upgrade IDs.
func (s *State) AddOffer(ids []string) {
	if len(ids) == 0 {
		return
	}
	s.PendingOfferIDs = append(s.PendingOfferIDs, ids)
}

// FoundationProgress returns the build progress (0.0–1.0) of the first active foundation,
// and whether any foundation is in progress. Uses StructureIndex to look up BuildCost.
func (s *State) FoundationProgress() (float64, bool) {
	for origin, deposited := range s.FoundationDeposited {
		entry, ok := s.World.StructureIndex[origin]
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
	world := GenerateWorld(100, 100, DefaultSeed)
	player := NewPlayer(world.Width/2, world.Height/2)

	return &State{
		Player:              player,
		World:               world,
		FoundationDeposited: make(map[Point]int),
		CompletedBeats:      make(map[string]bool),
	}
}
