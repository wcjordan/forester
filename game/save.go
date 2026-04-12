package game

import "forester/game/geom"

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
	ZoomLevel           float64
}

// SaveData collects a full snapshot of the game state for persistence.
func (g *Game) SaveData() SaveGameData {
	s := g.State
	return SaveGameData{
		Player:              s.Player.SaveData(),
		World:               s.World.SaveData(),
		Storage:             g.Stores.SaveData(),
		Villagers:           g.Villagers.SaveData(),
		FoundationDeposited: copyMap(s.FoundationDeposited),
		HouseOccupancy:      copyMap(s.HouseOccupancy),
		XP:                  s.XP,
		XPMilestoneIdx:      s.XPMilestoneIdx,
		PendingOfferIDs:     copyStringSliceSlice(s.pendingOfferIDs),
		CompletedBeats:      copyMap(s.completedBeats),
		ZoomLevel:           g.ZoomLevel,
	}
}

// LoadFrom restores the game from a SaveGameData snapshot without disk I/O.
// Useful for testing: load a pre-built fixture into a game that was created
// with a known clock and RNG.
func (g *Game) LoadFrom(data SaveGameData) error {
	return g.loadSaveData(data)
}

// loadSaveData restores the game from a SaveGameData snapshot.
func (g *Game) loadSaveData(data SaveGameData) error {
	player := &Player{}
	player.LoadFrom(data.Player)

	world := &World{}
	if err := world.LoadFrom(data.World); err != nil {
		return err
	}

	stores := NewStorageManager()
	stores.LoadFrom(data.Storage, world)

	vm := NewVillagerManager()
	vm.LoadFrom(data.Villagers)

	g.State = &State{
		Player:              player,
		World:               world,
		FoundationDeposited: copyMap(data.FoundationDeposited),
		HouseOccupancy:      copyMap(data.HouseOccupancy),
		XP:                  data.XP,
		XPMilestoneIdx:      data.XPMilestoneIdx,
		pendingOfferIDs:     copyStringSliceSlice(data.PendingOfferIDs),
		completedBeats:      copyMap(data.CompletedBeats),
	}
	g.Stores = stores
	g.Villagers = vm
	g.ZoomLevel = data.ZoomLevel
	return nil
}

func copyMap[K comparable, V any](m map[K]V) map[K]V {
	out := make(map[K]V, len(m))
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
