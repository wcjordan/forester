package game

import "math/rand"

// defaultSeed is used by newState to produce a consistent map each run.
// Pass a different seed to GenerateWorld for a different map.
const defaultSeed int64 = 42

// GenerateWorld creates a world with procedurally generated forest/grassland
// terrain using cellular automata. The same seed always produces the same map.
func GenerateWorld(width, height int, seed int64) *World {
	rng := rand.New(rand.NewSource(seed)) //nolint:gosec // game RNG, not crypto

	world := NewWorld(width, height)

	// Step 1: Seed each tile with ~60% forest probability.
	for y := range world.Tiles {
		for x := range world.Tiles[y] {
			if rng.Float64() < 0.60 {
				world.Tiles[y][x].Terrain = Forest
			}
		}
	}

	// Step 2: Smooth with 5 CA iterations.
	// Rule: a tile becomes Forest if it has >= 5 Forest neighbors (8-directional).
	// Out-of-bounds positions count as Forest, biasing edges toward Forest.
	const iterations = 5
	const forestThreshold = 5

	for range iterations {
		next := make([][]Tile, height)
		for y := range next {
			next[y] = make([]Tile, width)
			for x := range next[y] {
				if countForestNeighbors(world, x, y) >= forestThreshold {
					next[y][x].Terrain = Forest
				}
				// else Grassland (zero value, already set by make)
			}
		}
		world.Tiles = next
	}

	// Step 2.5: Assign random tree sizes to all Forest tiles.
	const minTreeSize = 4
	const maxTreeSize = 10
	for y := range world.Tiles {
		for x := range world.Tiles[y] {
			if world.Tiles[y][x].Terrain == Forest {
				world.Tiles[y][x].TreeSize = rng.Intn(maxTreeSize-minTreeSize+1) + minTreeSize
			}
		}
	}

	// Step 3: Clear a circle of Euclidean radius 5 at the center to guarantee a grassland spawn.
	cx, cy := width/2, height/2
	const spawnClearRadius = 5
	for dy := -spawnClearRadius; dy <= spawnClearRadius; dy++ {
		for dx := -spawnClearRadius; dx <= spawnClearRadius; dx++ {
			if dx*dx+dy*dy <= spawnClearRadius*spawnClearRadius {
				if tile := world.TileAt(cx+dx, cy+dy); tile != nil {
					tile.Terrain = Grassland
					tile.TreeSize = 0
				}
			}
		}
	}

	return world
}

// countForestNeighbors counts Forest neighbors in 8 directions.
// Out-of-bounds positions are treated as Forest.
func countForestNeighbors(w *World, x, y int) int {
	count := 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}
			nx, ny := x+dx, y+dy
			if !w.InBounds(nx, ny) || w.Tiles[ny][nx].Terrain == Forest {
				count++
			}
		}
	}
	return count
}
