package game

import "forester/game/geom"

// spawnAnchoredPlacer is an optional interface for StructureDef implementations
// that want to be placed as close as possible to the world spawn point rather
// than on the path from the player toward the center.
type spawnAnchoredPlacer interface {
	UseSpawnAnchoredPlacement() bool
}

// findStructureDefByFoundationType returns the StructureDef registered for the
// given FoundationType, or nil if none is found.
// The explicit FoundationType check guards against accidentally passing a BuiltType,
// which would also resolve in the dual-keyed map but is not a valid lookup.
func findStructureDefByFoundationType(ft StructureType) StructureDef {
	reg, ok := structures[ft]
	if !ok || reg.Def.FoundationType() != ft {
		return nil
	}
	return reg.Def
}

// SpawnFoundationByType looks up the StructureDef registered for ft and places its
// foundation near the player. Returns true if placed, false if no valid location
// is found or ft is not registered. Callers may retry on the next tick.
// Intended for use in story beat actions registered via RegisterStoryBeat.
func SpawnFoundationByType(env *Env, ft StructureType) bool {
	def := findStructureDefByFoundationType(ft)
	if def == nil {
		return false
	}
	return spawnFoundationAt(env.State.World, env.State.Player.X, env.State.Player.Y, def)
}

// spawnFoundationAt finds a valid location for def and places its foundation tile.
// Placement is near the world spawn point if def implements spawnAnchoredPlacer,
// otherwise near the player. Returns true if a foundation was placed, false if no
// valid location was found (caller may retry on the next tick).
func spawnFoundationAt(world *World, playerX, playerY int, def StructureDef) bool {
	fw, fh := def.Footprint()
	var cx, cy int
	if sa, ok := def.(spawnAnchoredPlacer); ok && sa.UseSpawnAnchoredPlacement() {
		cx, cy = findValidLocationNearSpawn(world, playerX, playerY, fw, fh)
	} else {
		cx, cy = findValidLocationNearPlayer(world, playerX, playerY, fw, fh)
	}
	if cx >= 0 {
		world.PlaceFoundation(cx, cy, def)
		return true
	}
	return false
}

// maybeSpawnFoundation checks each registered structure definition and places a foundation
// when its ShouldSpawn world condition is met. Each ShouldSpawn implementation is responsible
// for its own idempotency (e.g. checking that no foundation is already pending).
func maybeSpawnFoundation(env *Env) {
	IterateStructures(func(def StructureDef, cb StructureCallbacks) {
		if cb.ShouldSpawn != nil && cb.ShouldSpawn(env) {
			spawnFoundationAt(env.State.World, env.State.Player.X, env.State.Player.Y, def)
		}
	})
}

// abs returns the absolute value of n.
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
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
// expanding Chebyshev rings, returning the top-left corner of the first valid
// footW×footH area found. Returns (-1, -1) if none found.
func findValidLocationNearSpawn(world *World, playerX, playerY, footW, footH int) (x, y int) {
	spawnX := world.Width / 2
	spawnY := world.Height / 2
	// anchorX/anchorY is the top-left that would center the footprint on spawn.
	anchorX := spawnX - footW/2
	anchorY := spawnY - footH/2
	maxR := world.Width + world.Height

	x, y, found := geom.SpiralSearchDo(anchorX, anchorY, maxR, func(px, py int) bool {
		return isValidArea(world, playerX, playerY, px, py, footW, footH)
	})
	if found {
		return x, y
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
