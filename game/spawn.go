package game

import "sort"

// SpawnAnchoredPlacer is an optional interface for StructureDef implementations
// that want to be placed as close as possible to the world spawn point rather
// than on the path from the player toward the center.
type SpawnAnchoredPlacer interface {
	UseSpawnAnchoredPlacement() bool
}

// findStructureDefByFoundationType returns the StructureDef registered for the
// given FoundationType, or nil if none is found.
// The explicit FoundationType check guards against accidentally passing a BuiltType,
// which would also resolve in the dual-keyed map but is not a valid lookup.
func findStructureDefByFoundationType(ft StructureType) StructureDef {
	def := structures[ft]
	if def == nil || def.FoundationType() != ft {
		return nil
	}
	return def
}

// spawnFoundationAt finds a valid location for def and places its foundation tile.
// Placement is near the world spawn point if def implements SpawnAnchoredPlacer,
// otherwise near the player. Returns true if a foundation was placed, false if no
// valid location was found (caller may retry on the next tick).
func spawnFoundationAt(world *World, playerX, playerY int, def StructureDef) bool {
	w, h := def.Footprint()
	var cx, cy int
	if sa, ok := def.(SpawnAnchoredPlacer); ok && sa.UseSpawnAnchoredPlacement() {
		cx, cy = findValidLocationNearSpawn(world, playerX, playerY, w, h)
	} else {
		cx, cy = findValidLocationNearPlayer(world, playerX, playerY, w, h)
	}
	if cx >= 0 {
		world.SetStructure(cx, cy, w, h, def.FoundationType())
		world.IndexStructure(cx, cy, w, h, def)
		return true
	}
	return false
}

// maybeSpawnFoundation checks each registered structure definition and places a foundation
// when its ShouldSpawn world condition is met. Each ShouldSpawn implementation is responsible
// for its own idempotency (e.g. checking that no foundation is already pending).
func maybeSpawnFoundation(env *Env) {
	IterateStructures(func(def StructureDef) {
		if def.ShouldSpawn(env) {
			spawnFoundationAt(env.State.World, env.State.Player.X, env.State.Player.Y, def)
		}
	})
}

// findValidLocationNearPlayer walks from the player position toward the world center,
// returning the top-left corner of the first valid area of the given dimensions.
// Returns (-1, -1) if no valid location is found.
func findValidLocationNearPlayer(world *World, playerX, playerY, footW, footH int) (x, y int) {
	spawnX := world.Width / 2
	spawnY := world.Height / 2

	dx := spawnX - playerX
	dy := spawnY - playerY
	steps := abs(dx)
	if abs(dy) > steps {
		steps = abs(dy)
	}
	if steps == 0 {
		steps = 1
	}

	for i := 0; i <= steps; i++ {
		tx := playerX + dx*i/steps
		ty := playerY + dy*i/steps
		if isValidArea(world, playerX, playerY, tx, ty, footW, footH) {
			return tx, ty
		}
	}
	return -1, -1
}

// findValidLocationNearSpawn searches outward from the world spawn point in
// expanding Chebyshev rings, returning the top-left corner of the closest valid
// footW×footH area by Euclidean distance from footprint center to spawn.
// Stops as soon as the first valid location is found. Returns (-1, -1) if none found.
func findValidLocationNearSpawn(world *World, playerX, playerY, footW, footH int) (x, y int) {
	spawnX := world.Width / 2
	spawnY := world.Height / 2
	// anchorX/anchorY is the top-left that would center the footprint on spawn.
	anchorX := spawnX - footW/2
	anchorY := spawnY - footH/2

	footprintDist2 := func(px, py int) float64 {
		cx := float64(px) + float64(footW)/2 - float64(spawnX)
		cy := float64(py) + float64(footH)/2 - float64(spawnY)
		return cx*cx + cy*cy
	}

	type pos struct{ x, y int }
	maxR := world.Width + world.Height

	for r := 0; r <= maxR; r++ {
		var ring []pos
		chebyshevRingDo(anchorX, anchorY, r, func(px, py int) {
			ring = append(ring, pos{px, py})
		})

		// Sort this ring by Euclidean distance so we check the closest positions first.
		sort.Slice(ring, func(i, j int) bool {
			di := footprintDist2(ring[i].x, ring[i].y)
			dj := footprintDist2(ring[j].x, ring[j].y)
			if di != dj {
				return di < dj
			}
			if ring[i].y != ring[j].y {
				return ring[i].y < ring[j].y
			}
			return ring[i].x < ring[j].x
		})

		for _, p := range ring {
			if isValidArea(world, playerX, playerY, p.x, p.y, footW, footH) {
				return p.x, p.y
			}
		}
	}
	return -1, -1
}

// isValidArea returns true if the footW×footH area with top-left at (x, y) satisfies
// all placement constraints:
//   - Every footprint tile is in-bounds, grassland, and has no structure.
//   - No footprint tile overlaps the player's current position.
//   - The full 1-tile Chebyshev border around the footprint contains no structures
//     (ensures at least a 1-tile gap from every existing structure in all 8 directions).
func isValidArea(world *World, playerX, playerY, x, y, footW, footH int) bool {
	// Check footprint tiles.
	for dy := 0; dy < footH; dy++ {
		for dx := 0; dx < footW; dx++ {
			tx, ty := x+dx, y+dy
			// Player overlap check.
			if tx == playerX && ty == playerY {
				return false
			}
			tile := world.TileAt(tx, ty)
			if tile == nil || tile.Terrain != Grassland || tile.Structure != NoStructure {
				return false
			}
		}
	}

	// Check 1-tile Chebyshev border for existing structures.
	for by := y - 1; by <= y+footH; by++ {
		for bx := x - 1; bx <= x+footW; bx++ {
			// Skip the footprint itself (already checked above).
			if bx >= x && bx < x+footW && by >= y && by < y+footH {
				continue
			}
			tile := world.TileAt(bx, by)
			if tile != nil && tile.Structure != NoStructure {
				return false
			}
		}
	}

	return true
}
