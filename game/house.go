package game

import "time"

// HouseBuildCost is the number of wood required to complete a House foundation.
const HouseBuildCost = 50

// HouseSpawnThreshold is the amount of wood that must be stored in Log Storage
// before a House foundation will spawn.
const HouseSpawnThreshold = 50

func init() { structures = append(structures, houseDef{}) }

// houseDef implements StructureDef for the House structure.
type houseDef struct{}

// FoundationType returns the foundation tile type for House.
func (houseDef) FoundationType() StructureType { return FoundationHouse }

// BuiltType returns the built tile type for House.
func (houseDef) BuiltType() StructureType { return House }

// Footprint returns the 2×2 dimensions of a House.
func (houseDef) Footprint() (w, h int) { return 2, 2 }

// BuildCost returns the number of wood required to complete a House foundation.
func (houseDef) BuildCost() int { return HouseBuildCost }

// ShouldSpawn returns true when enough wood is currently stored in Log Storage.
func (houseDef) ShouldSpawn(env *Env) bool {
	return env.Stores.Total(Wood) >= HouseSpawnThreshold
}

// UseSpawnAnchoredPlacement signals that the house foundation should be placed
// as close as possible to the world spawn point rather than near the player.
func (houseDef) UseSpawnAnchoredPlacement() bool { return true }

// OnBuilt queues a 2-card milestone offer when a House is completed.
func (houseDef) OnBuilt(env *Env, _ Point) {
	env.State.AddOffer([]string{"build_speed", "deposit_speed"})
}

// OnPlayerInteraction handles adjacent-player interaction.
// When adjacent to a foundation, deposits one wood toward the build cost each cooldown tick.
// When adjacent to a built house, nothing happens (no storage).
func (d houseDef) OnPlayerInteraction(env *Env, origin Point, now time.Time) {
	tile := env.State.World.TileAt(origin.X, origin.Y)
	if tile == nil || tile.Structure != FoundationHouse {
		return
	}
	p := env.State.Player
	if !p.CooldownExpired(Build, now) {
		return
	}
	if p.Wood == 0 {
		return
	}
	env.State.FoundationDeposited[origin]++
	p.Wood--
	p.QueueCooldown(Build, now.Add(p.BuildInterval))
	if env.State.FoundationDeposited[origin] >= d.BuildCost() {
		w, h := d.Footprint()
		env.State.World.SetStructure(origin.X, origin.Y, w, h, House)
		env.State.World.IndexStructure(origin.X, origin.Y, w, h, d)
		delete(env.State.FoundationDeposited, origin)
		d.OnBuilt(env, origin)
	}
}
