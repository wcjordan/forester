package upgrades

import (
	"forester/game"
	"forester/game/geom"
)

func init() { game.RegisterUpgrade(spawnVillagerUpgrade{}) }

// spawnVillagerUpgrade spawns a villager at a random unoccupied house when applied.
// It only appears in card offers when at least one unoccupied house exists.
type spawnVillagerUpgrade struct{}

// ID returns the unique identifier for this upgrade.
func (spawnVillagerUpgrade) ID() string { return "spawn_villager" }

// Name returns the display name for this upgrade.
func (spawnVillagerUpgrade) Name() string { return "Recruit Villager" }

// Description returns the flavour text shown on the upgrade card.
func (spawnVillagerUpgrade) Description() string {
	return "A settler moves in.\nSpawns a villager at an\nunoccupied house."
}

// Apply spawns a villager at a randomly chosen unoccupied house.
func (spawnVillagerUpgrade) Apply(env *game.Env) {
	var unoccupied []geom.Point
	for origin, occupied := range env.State.HouseOccupancy {
		if !occupied {
			unoccupied = append(unoccupied, origin)
		}
	}
	if len(unoccupied) == 0 {
		return
	}
	origin := unoccupied[env.RNG.Intn(len(unoccupied))]
	game.SpawnVillagerAtHouse(env, origin)
}
