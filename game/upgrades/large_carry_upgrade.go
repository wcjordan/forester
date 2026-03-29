package upgrades

import "forester/game"

func init() { game.RegisterUpgrade(largeCarryCapacityUpgrade{}) }

type largeCarryCapacityUpgrade struct{}

// ID returns the unique identifier for this upgrade.
func (largeCarryCapacityUpgrade) ID() string { return "large_carry_capacity" }

// Name returns the display name for this upgrade.
func (largeCarryCapacityUpgrade) Name() string { return "Depot Carry Bonus" }

// Description returns the flavour text shown on the upgrade card.
func (largeCarryCapacityUpgrade) Description() string {
	return "The depot's efficiency inspires you.\nMax wood carried: +100."
}

// Apply increases the player's carry capacity by 100.
func (largeCarryCapacityUpgrade) Apply(env *game.Env) { env.State.Player.MaxCarry += 100 }
