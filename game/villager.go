package game

// VillagerMaxCarry is the maximum wood a villager can carry at once.
const VillagerMaxCarry = 5

// Villager is an autonomous agent that collects and delivers wood.
// Task behavior is added in the task tick layer; this struct holds only
// the fields that are always present.
type Villager struct {
	X, Y int
}

// SpawnVillager appends a new villager at (x, y) to the state.
func (s *State) SpawnVillager(x, y int) {
	s.Villagers = append(s.Villagers, &Villager{X: x, Y: y})
}
