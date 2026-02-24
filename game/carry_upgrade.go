package game

func init() { upgradeRegistry["carry_capacity"] = carryCapacityUpgrade{} }

type carryCapacityUpgrade struct{}

// ID returns the unique identifier for this upgrade.
func (carryCapacityUpgrade) ID() string { return "carry_capacity" }

// Name returns the display name for this upgrade.
func (carryCapacityUpgrade) Name() string { return "Expanded Carry Capacity" }

// Description returns the flavour text shown on the upgrade card.
func (carryCapacityUpgrade) Description() string {
	return "Your back grows stronger.\nMax wood carried: 20 \u2192 100."
}

// Apply sets the player's carry capacity to 100.
func (carryCapacityUpgrade) Apply(p *Player) { p.MaxCarry = 100 }
