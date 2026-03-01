package upgrades

import "forester/game"

func init() { game.RegisterUpgrade(moveSpeedUpgrade{}) }

// moveSpeedUpgrade reduces all player movement cooldowns by 10%.
type moveSpeedUpgrade struct{}

// ID returns the unique identifier for this upgrade.
func (moveSpeedUpgrade) ID() string { return "move_speed" }

// Name returns the display name for this upgrade.
func (moveSpeedUpgrade) Name() string { return "Faster Movement" }

// Description returns the flavour text shown on the upgrade card.
func (moveSpeedUpgrade) Description() string {
	return "Your legs carry you further.\nMovement speed +10%."
}

// Apply reduces the player's move speed multiplier by 10%, increasing movement speed.
func (moveSpeedUpgrade) Apply(env *game.Env) {
	env.State.Player.MoveSpeedMultiplier *= 0.9
}
