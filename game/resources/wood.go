package resources

import (
	"math/rand"
	"time"

	"forester/game"
	"forester/game/geom"
)

func init() { game.RegisterResource(woodDef{}) }

// woodDef implements game.ResourceDef for wood.
type woodDef struct{}

// woodHarvestPerStep is how much wood is taken from each adjacent Forest tile per harvest tick.
const woodHarvestPerStep = 1

// woodRegrowthCooldown is how often the regrowth tick fires.
const woodRegrowthCooldown = 500 * time.Millisecond

// woodRegrowthOdds is the 1-in-N chance each eligible Forest tile grows per regrowth tick.
const woodRegrowthOdds = 40

// woodMaxTreeSize is the maximum TreeSize a Forest tile can grow to.
const woodMaxTreeSize = 10

// Type returns the resource type for wood.
func (woodDef) Type() game.ResourceType { return game.Wood }

// Harvest harvests wood from the tile under the player and the three Forest tiles
// in front of the player: straight ahead and the two forward diagonals.
// Each tile loses woodHarvestPerStep wood; when TreeSize reaches 0 it stays Forest (cut tree).
// The harvested wood is added to the player's inventory.
// The harvest is skipped if the Harvest cooldown has not elapsed.
func (woodDef) Harvest(env *game.Env, now time.Time) {
	p := env.State.Player
	if !p.CooldownExpired(game.Harvest, now) {
		return
	}
	p.SetCooldown(game.Harvest, now.Add(p.HarvestInterval))
	if p.Inventory[game.Wood] >= p.MaxCarry {
		return
	}
	dx, dy := p.FacingDX, p.FacingDY
	// Four tiles: under the player, straight ahead, diagonal-left, diagonal-right.
	targets := [4][2]int{
		{p.TileX(), p.TileY()},
		{p.TileX() + dx, p.TileY() + dy},
		{p.TileX() + dx - dy, p.TileY() + dy + dx},
		{p.TileX() + dx + dy, p.TileY() + dy - dx},
	}
	for _, coord := range targets {
		tile := env.State.World.TileAt(coord[0], coord[1])
		if tile == nil || tile.Terrain != game.Forest {
			continue
		}
		canTake := min(woodHarvestPerStep, p.MaxCarry-p.Inventory[game.Wood])
		harvest := min(canTake, tile.TreeSize)
		tile.TreeSize -= harvest
		p.Inventory[game.Wood] += harvest
		if harvest > 0 {
			p.LastHarvestAt = now
			game.AwardXP(env, game.XPPerWoodChopped*harvest)
		}
	}
}

// Regrow advances tree regrowth probabilistically if the regrowth cooldown has elapsed.
// Each eligible Forest tile (including TreeSize=0 cut trees) has a 1/woodRegrowthOdds chance to grow,
// unless it is in the precomputed NoGrowTiles set (within noGrowRadius of the spawn point or any structure tile).
func (woodDef) Regrow(env *game.Env, rng *rand.Rand, now time.Time) {
	w := env.State.World
	if !w.RegrowElapsed(now) {
		return
	}
	w.MarkRegrowCooldown(woodRegrowthCooldown, now)
	for y := range w.Tiles {
		for x := range w.Tiles[y] {
			tile := &w.Tiles[y][x]
			if tile.Terrain != game.Forest || tile.TreeSize >= woodMaxTreeSize {
				continue
			}
			if _, blocked := w.NoGrowTiles[geom.Point{X: x, Y: y}]; blocked {
				// Cut trees (TreeSize=0) in no-grow zones convert to Grassland
				// so the cleared area stays open for village growth.
				if tile.TreeSize == 0 {
					tile.Terrain = game.Grassland
				}
				continue
			}
			if rng.Intn(woodRegrowthOdds) == 0 {
				tile.TreeSize++
			}
		}
	}
}
