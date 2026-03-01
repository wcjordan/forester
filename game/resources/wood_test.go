package resources

import (
	"math/rand"
	"testing"
	"time"

	"forester/game"
	"forester/game/geom"
	"forester/game/internal/gametest"
)

// regrowTick calls woodDef.Regrow with a timestamp far enough ahead to fire
// the regrowth logic regardless of the current cooldown state.
func regrowTick(env *game.Env, rng *rand.Rand, i int) {
	t0 := time.Time{}
	woodDef{}.Regrow(env, rng, t0.Add(time.Duration(i+1)*woodRegrowthCooldown*2))
}

func TestRegrowWood(t *testing.T) {
	// Use a 20×20 world and place Forest tiles at (0,0), which is ~14 tiles
	// from the spawn center (10,10) — well outside the no-grow radius of 8.
	t.Run("cut tree eventually grows", func(t *testing.T) {
		rng := rand.New(rand.NewSource(0))
		w := game.NewWorld(20, 20)
		w.Tiles[0][0] = game.Tile{Terrain: game.Forest, TreeSize: 0}
		env := &game.Env{State: &game.State{World: w}, Stores: game.NewStorageManager()}
		grew := false
		for i := range 1000 {
			regrowTick(env, rng, i)
			if w.Tiles[0][0].TreeSize > 0 {
				grew = true
				break
			}
		}
		if !grew {
			t.Error("cut tree (Forest/TreeSize=0) should eventually grow")
		}
	})

	t.Run("forest eventually grows toward woodMaxTreeSize", func(t *testing.T) {
		rng := rand.New(rand.NewSource(0))
		w := game.NewWorld(20, 20)
		w.Tiles[0][0] = game.Tile{Terrain: game.Forest, TreeSize: 5}
		env := &game.Env{State: &game.State{World: w}, Stores: game.NewStorageManager()}
		grew := false
		for i := range 1000 {
			regrowTick(env, rng, i)
			if w.Tiles[0][0].TreeSize > 5 {
				grew = true
				break
			}
		}
		if !grew {
			t.Error("forest should eventually grow toward woodMaxTreeSize")
		}
	})

	t.Run("forest at woodMaxTreeSize does not grow further", func(t *testing.T) {
		rng := rand.New(rand.NewSource(0))
		w := game.NewWorld(20, 20)
		w.Tiles[0][0] = game.Tile{Terrain: game.Forest, TreeSize: woodMaxTreeSize}
		env := &game.Env{State: &game.State{World: w}, Stores: game.NewStorageManager()}
		for i := range 1000 {
			regrowTick(env, rng, i)
		}
		if w.Tiles[0][0].TreeSize != woodMaxTreeSize {
			t.Errorf("TreeSize = %d, want %d", w.Tiles[0][0].TreeSize, woodMaxTreeSize)
		}
	})

	t.Run("grassland is unaffected", func(t *testing.T) {
		rng := rand.New(rand.NewSource(0))
		w := game.NewWorld(20, 20)
		w.Tiles[0][0] = game.Tile{Terrain: game.Grassland}
		env := &game.Env{State: &game.State{World: w}, Stores: game.NewStorageManager()}
		for i := range 1000 {
			regrowTick(env, rng, i)
		}
		tile := w.Tiles[0][0]
		if tile.Terrain != game.Grassland {
			t.Errorf("Terrain = %v, want Grassland", tile.Terrain)
		}
	})

	t.Run("cut tree within spawn no-grow zone converts to Grassland", func(t *testing.T) {
		rng := rand.New(rand.NewSource(0))
		// 20×20 world: spawn = (10,10). Tile at (10,10) is distance 0 ≤ 8.
		w := game.NewWorld(20, 20)
		w.Tiles[10][10] = game.Tile{Terrain: game.Forest, TreeSize: 0}
		env := &game.Env{State: &game.State{World: w}, Stores: game.NewStorageManager()}
		regrowTick(env, rng, 0)
		if w.Tiles[10][10].Terrain != game.Grassland {
			t.Errorf("Terrain = %v, want Grassland (cut tree in no-grow zone should convert)", w.Tiles[10][10].Terrain)
		}
	})

	t.Run("living forest within spawn no-grow zone does not grow or convert", func(t *testing.T) {
		rng := rand.New(rand.NewSource(0))
		w := game.NewWorld(20, 20)
		w.Tiles[10][10] = game.Tile{Terrain: game.Forest, TreeSize: 5}
		env := &game.Env{State: &game.State{World: w}, Stores: game.NewStorageManager()}
		for i := range 1000 {
			regrowTick(env, rng, i)
		}
		tile := w.Tiles[10][10]
		if tile.Terrain != game.Forest {
			t.Errorf("Terrain = %v, want Forest (living tree should not convert)", tile.Terrain)
		}
		if tile.TreeSize != 5 {
			t.Errorf("TreeSize = %d, want 5 (living tree in no-grow zone should not grow)", tile.TreeSize)
		}
	})

	t.Run("cut tree within building no-grow zone converts to Grassland", func(t *testing.T) {
		rng := rand.New(rand.NewSource(0))
		// 40×40 world: spawn = (20,20). Place a structure at (5,5) and a Forest
		// tile at (5,10) — distance 5 ≤ 8 from the structure, and distance
		// sqrt(225+100)=~18 from spawn (safely outside the spawn zone).
		w := game.NewWorld(40, 40)
		w.PlaceBuilt(5, 5, gametest.WallDef{Width: 1, Height: 1})
		w.Tiles[10][5] = game.Tile{Terrain: game.Forest, TreeSize: 0}
		env := &game.Env{State: &game.State{World: w}, Stores: game.NewStorageManager()}
		regrowTick(env, rng, 0)
		if w.Tiles[10][5].Terrain != game.Grassland {
			t.Errorf("Terrain = %v, want Grassland (cut tree in building no-grow zone should convert)", w.Tiles[10][5].Terrain)
		}
	})
}

// makeEnv builds a minimal Env with the player at (px, py) facing north.
func makeEnv(w *game.World, px, py int) (*game.Env, *game.Player) {
	p := game.NewPlayer(px, py)
	s := &game.State{
		Player:              p,
		World:               w,
		FoundationDeposited: make(map[geom.Point]int),
	}
	return &game.Env{State: s, Stores: game.NewStorageManager()}, p
}

func TestHarvestWood(t *testing.T) {
	makeWorld := func(terrain game.TerrainType, treeSize int) (*game.World, *game.Tile) {
		w := game.NewWorld(5, 5)
		w.Tiles[1][2] = game.Tile{Terrain: terrain, TreeSize: treeSize} // above player at (2,2)
		return w, &w.Tiles[1][2]
	}

	t.Run("harvests adjacent forest tile", func(t *testing.T) {
		w, tile := makeWorld(game.Forest, 5)
		env, p := makeEnv(w, 2, 2)
		woodDef{}.Harvest(env, time.Now())
		if p.Inventory[game.Wood] != 1 {
			t.Errorf("Inventory[Wood] = %d, want 1", p.Inventory[game.Wood])
		}
		if tile.TreeSize != 4 {
			t.Errorf("TreeSize = %d, want 4", tile.TreeSize)
		}
		if tile.Terrain != game.Forest {
			t.Errorf("Terrain = %v, want Forest (tree not depleted)", tile.Terrain)
		}
	})

	t.Run("stays Forest when tree depleted", func(t *testing.T) {
		w, tile := makeWorld(game.Forest, 1)
		env, p := makeEnv(w, 2, 2)
		woodDef{}.Harvest(env, time.Now())
		if p.Inventory[game.Wood] != 1 {
			t.Errorf("Inventory[Wood] = %d, want 1", p.Inventory[game.Wood])
		}
		if tile.TreeSize != 0 {
			t.Errorf("TreeSize = %d, want 0", tile.TreeSize)
		}
		if tile.Terrain != game.Forest {
			t.Errorf("Terrain = %v, want Forest (cut tree stays Forest)", tile.Terrain)
		}
	})

	t.Run("does not harvest from cut tree", func(t *testing.T) {
		w, tile := makeWorld(game.Forest, 0)
		env, p := makeEnv(w, 2, 2)
		woodDef{}.Harvest(env, time.Now())
		if p.Inventory[game.Wood] != 0 {
			t.Errorf("Inventory[Wood] = %d, want 0 (cut tree should not yield wood)", p.Inventory[game.Wood])
		}
		if tile.Terrain != game.Forest {
			t.Errorf("Terrain changed from Forest unexpectedly")
		}
		if tile.TreeSize != 0 {
			t.Errorf("TreeSize changed from 0 unexpectedly")
		}
	})

	t.Run("does not harvest from grassland", func(t *testing.T) {
		w, _ := makeWorld(game.Grassland, 0)
		env, p := makeEnv(w, 2, 2)
		woodDef{}.Harvest(env, time.Now())
		if p.Inventory[game.Wood] != 0 {
			t.Errorf("Inventory[Wood] = %d, want 0 (grassland should not yield wood)", p.Inventory[game.Wood])
		}
	})

	t.Run("safe at world edge — no panic on nil tile", func(t *testing.T) {
		w := game.NewWorld(3, 3)
		w.Tiles[0][0] = game.Tile{Terrain: game.Forest, TreeSize: 5}
		env, _ := makeEnv(w, 0, 0) // at corner; two neighbors are out of bounds
		woodDef{}.Harvest(env, time.Now())
	})

	t.Run("harvests the forward arc (straight and both diagonals)", func(t *testing.T) {
		w := game.NewWorld(5, 5)
		env, p := makeEnv(w, 2, 2)
		// Default facing is north (0,-1).
		// Forward arc: N (2,1), NW (1,1), NE (3,1).
		// Non-forward: S (2,3), E (3,2), W (1,2).
		for _, coord := range [][2]int{{2, 1}, {1, 1}, {3, 1}} {
			w.Tiles[coord[1]][coord[0]] = game.Tile{Terrain: game.Forest, TreeSize: 3}
		}
		for _, coord := range [][2]int{{2, 3}, {3, 2}, {1, 2}} {
			w.Tiles[coord[1]][coord[0]] = game.Tile{Terrain: game.Forest, TreeSize: 3}
		}
		woodDef{}.Harvest(env, time.Now())
		if p.Inventory[game.Wood] != 3 {
			t.Errorf("Inventory[Wood] = %d, want 3 (forward arc harvested)", p.Inventory[game.Wood])
		}
		// Forward arc tiles reduced.
		for _, coord := range [][2]int{{2, 1}, {1, 1}, {3, 1}} {
			if w.Tiles[coord[1]][coord[0]].TreeSize != 2 {
				t.Errorf("forward tile (%d,%d) TreeSize = %d, want 2", coord[0], coord[1], w.Tiles[coord[1]][coord[0]].TreeSize)
			}
		}
		// Non-forward tiles untouched.
		for _, coord := range [][2]int{{2, 3}, {3, 2}, {1, 2}} {
			if w.Tiles[coord[1]][coord[0]].TreeSize != 3 {
				t.Errorf("non-forward tile (%d,%d) should be untouched, TreeSize = %d, want 3", coord[0], coord[1], w.Tiles[coord[1]][coord[0]].TreeSize)
			}
		}
	})
}

func TestHarvestCapacityWood(t *testing.T) {
	t.Run("harvest stops at InitialCarryingCapacity", func(t *testing.T) {
		w := game.NewWorld(5, 5)
		w.Tiles[1][2] = game.Tile{Terrain: game.Forest, TreeSize: 10}
		env, p := makeEnv(w, 2, 2)
		p.Inventory[game.Wood] = game.InitialCarryingCapacity
		woodDef{}.Harvest(env, time.Now())
		if p.Inventory[game.Wood] != game.InitialCarryingCapacity {
			t.Errorf("Inventory[Wood] = %d, want %d (should not exceed InitialCarryingCapacity)",
				p.Inventory[game.Wood], game.InitialCarryingCapacity)
		}
		if w.Tiles[1][2].TreeSize != 10 {
			t.Errorf("TreeSize = %d, want 10 (should not harvest when full)", w.Tiles[1][2].TreeSize)
		}
	})

	t.Run("partial fill at near-max", func(t *testing.T) {
		w := game.NewWorld(5, 5)
		w.Tiles[1][2] = game.Tile{Terrain: game.Forest, TreeSize: 10}
		env, p := makeEnv(w, 2, 2)
		p.Inventory[game.Wood] = game.InitialCarryingCapacity - 1
		woodDef{}.Harvest(env, time.Now())
		if p.Inventory[game.Wood] != game.InitialCarryingCapacity {
			t.Errorf("Inventory[Wood] = %d, want %d (should fill to exactly InitialCarryingCapacity)",
				p.Inventory[game.Wood], game.InitialCarryingCapacity)
		}
	})
}
