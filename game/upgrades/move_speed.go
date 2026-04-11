package upgrades

import "forester/game"

func init() { game.RegisterUpgrade(moveSpeedUpgrade{}) }

// moveSpeedUpgrade increases player movement speed by 10%.
type moveSpeedUpgrade struct{}

// ID returns the unique identifier for this upgrade.
func (moveSpeedUpgrade) ID() string { return "move_speed" }

// Name returns the display name for this upgrade.
func (moveSpeedUpgrade) Name() string { return "Faster Movement" }

// Description returns the flavour text shown on the upgrade card.
func (moveSpeedUpgrade) Description() string {
	return "Your legs carry you further.\nMovement speed +10%."
}

// Apply increases the player's move speed by 10%.
func (moveSpeedUpgrade) Apply(env *game.Env) {
	env.State.Player.MoveSpeed /= 0.9
}
