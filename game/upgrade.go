package game

// UpgradeDef defines an upgrade that can be applied to the player.
type UpgradeDef interface {
	ID() string
	Name() string
	Description() string
	Apply(p *Player)
}

// upgradeRegistry maps stable upgrade IDs to their runtime definitions.
// Each upgrade file self-registers via init().
var upgradeRegistry = map[string]UpgradeDef{}

// RegisterUpgrade adds an UpgradeDef to the global registry.
// Call this from an init() function in an external package (e.g. game/upgrades).
func RegisterUpgrade(u UpgradeDef) { upgradeRegistry[u.ID()] = u }
