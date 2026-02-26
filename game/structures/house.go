package structures

import (
	"time"

	"forester/game"
)

const houseBuildCost = 50

func init() { game.RegisterStructure(houseDef{}) }

// houseDef implements game.StructureDef for the House structure.
type houseDef struct{}

// FoundationType returns the foundation tile type for House.
func (houseDef) FoundationType() game.StructureType { return game.FoundationHouse }

// BuiltType returns the built tile type for House.
func (houseDef) BuiltType() game.StructureType { return game.House }

// Footprint returns the 2×2 dimensions of a House.
func (houseDef) Footprint() (w, h int) { return 2, 2 }

// BuildCost returns the number of wood required to complete a House foundation.
func (houseDef) BuildCost() int { return houseBuildCost }

// ShouldSpawn is the world condition for spawning additional houses.
// Returns true when at least one house has been built and no house foundation is pending.
// The first house is handled by the story beat system; this drives all subsequent spawns.
func (houseDef) ShouldSpawn(env *game.Env) bool {
	built := len(env.State.World.StructureTypeIndex[game.House])
	pending := len(env.State.World.StructureTypeIndex[game.FoundationHouse])
	return built >= 1 && pending == 0
}

// UseSpawnAnchoredPlacement signals that the house foundation should be placed
// as close as possible to the world spawn point rather than near the player.
func (houseDef) UseSpawnAnchoredPlacement() bool { return true }

// OnBuilt is called when a House is completed. It spawns a villager at the
// first clear tile adjacent to the house footprint.
func (d houseDef) OnBuilt(env *game.Env, origin game.Point) {
	fw, fh := d.Footprint()
	px, py := env.State.Player.X, env.State.Player.Y
	// Candidates: tiles just outside each edge of the footprint.
	candidates := []game.Point{
		{X: origin.X - 1, Y: origin.Y},
		{X: origin.X - 1, Y: origin.Y + fh - 1},
		{X: origin.X + fw, Y: origin.Y},
		{X: origin.X + fw, Y: origin.Y + fh - 1},
		{X: origin.X, Y: origin.Y - 1},
		{X: origin.X + fw - 1, Y: origin.Y - 1},
		{X: origin.X, Y: origin.Y + fh},
		{X: origin.X + fw - 1, Y: origin.Y + fh},
	}
	for _, c := range candidates {
		if c.X == px && c.Y == py {
			continue
		}
		tile := env.State.World.TileAt(c.X, c.Y)
		if tile == nil || tile.Structure != game.NoStructure {
			continue
		}
		env.State.SpawnVillager(c.X, c.Y)
		return
	}
}

// OnPlayerInteraction handles adjacent-player interaction.
// When adjacent to a foundation, deposits one wood toward the build cost each cooldown tick.
// When adjacent to a built house, nothing happens (no storage).
func (d houseDef) OnPlayerInteraction(env *game.Env, origin game.Point, now time.Time) {
	tile := env.State.World.TileAt(origin.X, origin.Y)
	if tile == nil || tile.Structure != game.FoundationHouse {
		return
	}
	p := env.State.Player
	if !p.CooldownExpired(game.Build, now) {
		return
	}
	if p.Wood == 0 {
		return
	}
	env.State.FoundationDeposited[origin]++
	p.Wood--
	p.QueueCooldown(game.Build, now.Add(p.BuildInterval))
	if env.State.FoundationDeposited[origin] >= d.BuildCost() {
		w, h := d.Footprint()
		env.State.World.SetStructure(origin.X, origin.Y, w, h, game.House)
		env.State.World.IndexStructure(origin.X, origin.Y, w, h, d)
		delete(env.State.FoundationDeposited, origin)
		d.OnBuilt(env, origin)
	}
}
