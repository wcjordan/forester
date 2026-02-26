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
// Panics on nil, empty ID, or duplicate registration.
func RegisterUpgrade(u UpgradeDef) {
	if u == nil {
		panic("RegisterUpgrade: upgrade is nil")
	}
	id := u.ID()
	if id == "" {
		panic("RegisterUpgrade: upgrade ID is empty")
	}
	if _, exists := upgradeRegistry[id]; exists {
		panic("RegisterUpgrade: duplicate upgrade ID: " + id)
	}
	upgradeRegistry[id] = u
}
