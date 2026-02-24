package game

import "time"

// State holds serializable game state (truth data).
// Derived runtime structures (e.g. StorageManager) live on Game.
type State struct {
	Player              *Player
	World               *World
	FoundationDeposited map[Point]int
	// PendingOfferIDs stores each queued offer as a slice of upgrade IDs (strings),
	// keeping State serializable without embedding interface values.
	PendingOfferIDs [][]string
}

// AddOffer enqueues a card offer by its upgrade IDs.
func (s *State) AddOffer(ids []string) {
	if len(ids) == 0 {
		return
	}
	s.PendingOfferIDs = append(s.PendingOfferIDs, ids)
}

// HasPendingOffer reports whether there is at least one offer waiting.
func (s *State) HasPendingOffer() bool {
	return len(s.PendingOfferIDs) > 0
}

// CurrentOffer resolves the front offer's IDs to UpgradeDef values.
// Returns nil when there is no pending offer or no IDs resolve.
func (s *State) CurrentOffer() []UpgradeDef {
	if len(s.PendingOfferIDs) == 0 {
		return nil
	}
	ids := s.PendingOfferIDs[0]
	result := make([]UpgradeDef, 0, len(ids))
	for _, id := range ids {
		if u, ok := upgradeRegistry[id]; ok {
			result = append(result, u)
		}
	}
	return result
}

// SelectCard applies the card at idx from the front offer and pops it from the queue.
func (s *State) SelectCard(idx int) {
	offer := s.CurrentOffer()
	if len(offer) == 0 {
		return
	}
	if idx >= 0 && idx < len(offer) {
		offer[idx].Apply(s.Player)
	}
	s.PendingOfferIDs = s.PendingOfferIDs[1:]
}

// Harvest harvests adjacent trees without moving the player.
// Spawns a foundation when the spawn condition is met.
func (s *State) Harvest(env *Env) {
	s.Player.HarvestAdjacent(s.World)
	s.maybeSpawnFoundation(env)
}

// TickAdjacentStructures calls OnPlayerInteraction once per structure instance
// that the player is cardinally adjacent to, then commits any pending cooldowns.
// Cooldowns are committed after all interactions so that multiple adjacent
// structures of the same type all fire within the same tick.
func (s *State) TickAdjacentStructures(env *Env, now time.Time) {
	seen := make(map[Point]bool)
	for _, d := range [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
		p := Point{s.Player.X + d[0], s.Player.Y + d[1]}
		entry, ok := s.World.StructureIndex[p]
		if !ok || seen[entry.Origin] {
			continue
		}
		seen[entry.Origin] = true
		entry.Def.OnPlayerInteraction(env, entry.Origin, now)
	}
	s.Player.commitCooldowns()
}

// newState creates an initial game state with defaults.
func newState() *State {
	world := GenerateWorld(100, 100, DefaultSeed)
	player := NewPlayer(world.Width/2, world.Height/2)

	return &State{
		Player:              player,
		World:               world,
		FoundationDeposited: make(map[Point]int),
	}
}
