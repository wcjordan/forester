package game

// UpgradeDef defines an upgrade that can be applied to the player.
type UpgradeDef interface {
	ID() string
	Name() string
	Description() string
	Apply(p *Player)
}

// CardOffer is a set of upgrade choices presented at once; the player picks one.
type CardOffer []UpgradeDef
