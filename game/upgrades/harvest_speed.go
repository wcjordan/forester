package upgrades

import (
	"time"

	"forester/game"
)

func init() { game.RegisterUpgrade(harvestSpeedUpgrade{}) }

// harvestSpeedUpgrade reduces the time between player wood harvests by 10%.
type harvestSpeedUpgrade struct{}

// ID returns the unique identifier for this upgrade.
func (harvestSpeedUpgrade) ID() string { return "harvest_speed" }

// Name returns the display name for this upgrade.
func (harvestSpeedUpgrade) Name() string { return "Faster Chopping" }

// Description returns the flavour text shown on the upgrade card.
func (harvestSpeedUpgrade) Description() string {
	return "Your axe swings faster.\nWood harvest speed +10%."
}

// Apply reduces the player's harvest interval by 10%.
func (harvestSpeedUpgrade) Apply(env *game.Env) {
	p := env.State.Player
	p.HarvestInterval = time.Duration(float64(p.HarvestInterval) * 0.9)
}
