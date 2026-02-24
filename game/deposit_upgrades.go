package game

import "time"

func init() {
	upgradeRegistry["build_speed"] = buildSpeedUpgrade{}
	upgradeRegistry["deposit_speed"] = depositSpeedUpgrade{}
}

// buildSpeedUpgrade reduces the time between foundation deposits by 10%.
type buildSpeedUpgrade struct{}

// ID returns the unique identifier for this upgrade.
func (buildSpeedUpgrade) ID() string { return "build_speed" }

// Name returns the display name for this upgrade.
func (buildSpeedUpgrade) Name() string { return "Faster Construction" }

// Description returns the flavour text shown on the upgrade card.
func (buildSpeedUpgrade) Description() string {
	return "Your hands move quicker.\nFoundation build speed +10%."
}

// Apply reduces the player's foundation deposit interval by 10%.
func (buildSpeedUpgrade) Apply(p *Player) {
	p.BuildInterval = time.Duration(float64(p.BuildInterval) * 0.9)
}

// depositSpeedUpgrade reduces the time between storage deposits by 10%.
type depositSpeedUpgrade struct{}

// ID returns the unique identifier for this upgrade.
func (depositSpeedUpgrade) ID() string { return "deposit_speed" }

// Name returns the display name for this upgrade.
func (depositSpeedUpgrade) Name() string { return "Faster Depositing" }

// Description returns the flavour text shown on the upgrade card.
func (depositSpeedUpgrade) Description() string {
	return "You unload wood with ease.\nStorage deposit speed +10%."
}

// Apply reduces the player's storage deposit interval by 10%.
func (depositSpeedUpgrade) Apply(p *Player) {
	p.DepositInterval = time.Duration(float64(p.DepositInterval) * 0.9)
}
