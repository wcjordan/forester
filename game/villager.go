package game

import (
	"math/rand"
	"time"
)

// VillagerMaxCarry is the maximum wood a villager can carry at once.
const VillagerMaxCarry = 5

// villagerMoveCooldown is how often a villager takes one movement or harvest step.
const villagerMoveCooldown = 300 * time.Millisecond

// VillagerTask identifies the villager's current activity.
type VillagerTask int

// VillagerTask values.
const (
	VillagerIdle              VillagerTask = iota // choosing next task
	VillagerWalkingToTree                         // moving to tree tile to harvest
	VillagerCarryingToStorage                     // carrying harvested wood to log storage
	VillagerWalkingToStorage                      // walking to log storage to fetch wood
	VillagerDeliveringToHouse                     // carrying fetched wood to a house foundation
)

// String returns a short readable name for the task.
func (t VillagerTask) String() string {
	switch t {
	case VillagerIdle:
		return "Idle"
	case VillagerWalkingToTree:
		return "WalkingToTree"
	case VillagerCarryingToStorage:
		return "CarryingToStorage"
	case VillagerWalkingToStorage:
		return "WalkingToStorage"
	case VillagerDeliveringToHouse:
		return "DeliveringToHouse"
	default:
		return "Unknown"
	}
}

// Villager is an autonomous agent that collects and delivers wood.
type Villager struct {
	X, Y         int
	Wood         int
	Task         VillagerTask
	TargetX      int
	TargetY      int
	moveCooldown time.Time
	path         []Point // nil = recompute; []Point{} = at goal
}

// VillagerManager manages the set of autonomous villagers at runtime.
type VillagerManager struct {
	Villagers []*Villager
}

// NewVillagerManager creates an empty VillagerManager.
func NewVillagerManager() *VillagerManager {
	return &VillagerManager{}
}

// Spawn appends a new idle villager at (x, y).
func (vm *VillagerManager) Spawn(x, y int) {
	vm.Villagers = append(vm.Villagers, &Villager{X: x, Y: y})
}

// Tick advances every villager by one step.
func (vm *VillagerManager) Tick(env *Env, rng *rand.Rand, now time.Time) {
	for _, v := range vm.Villagers {
		v.Tick(env, rng, now)
	}
}

// Count returns the number of villagers.
func (vm *VillagerManager) Count() int {
	return len(vm.Villagers)
}

// Tick advances this villager by one game step, gated on moveCooldown.
func (v *Villager) Tick(env *Env, rng *rand.Rand, now time.Time) {
	if now.Before(v.moveCooldown) {
		return
	}
	v.moveCooldown = now.Add(villagerMoveCooldown)

	switch v.Task {
	case VillagerIdle:
		v.pickTask(env, rng)

	case VillagerWalkingToTree:
		tile := env.State.World.TileAt(v.TargetX, v.TargetY)
		// Target no longer valid (exhausted or gone).
		if tile == nil || tile.Terrain != Forest || tile.TreeSize == 0 {
			if v.Wood > 0 {
				v.headToStorage(env)
			} else {
				v.Task = VillagerIdle
			}
			return
		}
		if v.X == v.TargetX && v.Y == v.TargetY {
			// On the tree tile: harvest one wood.
			take := min(1, tile.TreeSize)
			tile.TreeSize -= take
			v.Wood += take
			if v.Wood >= VillagerMaxCarry || tile.TreeSize == 0 {
				v.headToStorage(env)
			}
		} else {
			v.move(env.State.World)
		}

	case VillagerCarryingToStorage:
		if v.X == v.TargetX && v.Y == v.TargetY {
			origin, ok := storageOriginAdjacent(env.State.World, v.X, v.Y)
			if ok {
				v.Wood -= env.Stores.DepositAt(origin, v.Wood)
			}
			if v.Wood > 0 {
				// Partial deposit (storage full); retry at another storage.
				v.headToStorage(env)
			} else {
				v.Task = VillagerIdle
			}
		} else {
			v.move(env.State.World)
		}

	case VillagerWalkingToStorage:
		if v.X == v.TargetX && v.Y == v.TargetY {
			origin, ok := storageOriginAdjacent(env.State.World, v.X, v.Y)
			if ok {
				space := VillagerMaxCarry - v.Wood
				fetched := env.Stores.WithdrawFrom(origin, space)
				if fetched > 0 {
					v.Wood += fetched
					if !v.headToHouse(env) {
						// No house foundation; return the wood and go idle.
						v.Wood -= env.Stores.DepositAt(origin, v.Wood)
						v.Task = VillagerIdle
					}
					return
				}
			}
			v.Task = VillagerIdle
		} else {
			v.move(env.State.World)
		}

	case VillagerDeliveringToHouse:
		if v.X == v.TargetX && v.Y == v.TargetY {
			foundOrigin, ok := foundationHouseOriginAdjacent(env.State.World, v.X, v.Y)
			if ok {
				entry, hasEntry := env.State.World.structureIndex[foundOrigin]
				if hasEntry {
					buildCost := entry.Def.BuildCost()
					current := env.State.FoundationDeposited[foundOrigin]
					needed := buildCost - current
					deposit := min(v.Wood, needed)
					if deposit > 0 {
						env.State.FoundationDeposited[foundOrigin] += deposit
						v.Wood -= deposit
						if env.State.FoundationDeposited[foundOrigin] >= buildCost {
							FinalizeFoundation(env, entry.Def, foundOrigin)
						}
					}
				}
			}
			v.Task = VillagerIdle
		} else {
			v.move(env.State.World)
		}
	}
}

// pickTask selects the villager's next task probabilistically based on log storage fill.
// P(chop wood → storage) = 1 - fillRatio; P(fetch storage → house) = fillRatio.
// Falls back to the other task when the preferred one has no valid target.
func (v *Villager) pickTask(env *Env, rng *rand.Rand) {
	// Set this higher to encourage building earlier.  2.0 encourages building when > 25% full
	const fillFactor = 2.0

	total := env.Stores.Total(Wood)
	storageCap := env.Stores.TotalCapacity(Wood)

	fillRatio := 0.0
	if storageCap > 0 {
		fillRatio = fillFactor * float64(total) / float64(storageCap)
		if fillRatio > 1 {
			fillRatio = 1
		}
	}

	wantBuild := rng.Float64() <= fillRatio
	if wantBuild {
		if v.tryAssignDeliverTask(env) {
			return
		}
	}
	v.tryAssignChopTask(env)
}

func (v *Villager) tryAssignChopTask(env *Env) bool {
	tx, ty, ok := findNearestTree(env.State.World, v.X, v.Y)
	if !ok {
		return false
	}
	v.Task = VillagerWalkingToTree
	v.TargetX, v.TargetY = tx, ty
	v.path = nil
	return true
}

func (v *Villager) tryAssignDeliverTask(env *Env) bool {
	if env.Stores.Total(Wood) == 0 {
		return false
	}
	if len(env.State.World.StructureTypeIndex[FoundationHouse]) == 0 {
		return false
	}
	tx, ty, ok := nearestClearTileAdjacent(env.State.World, LogStorage, v.X, v.Y)
	if !ok {
		return false
	}
	v.Task = VillagerWalkingToStorage
	v.TargetX, v.TargetY = tx, ty
	v.path = nil
	return true
}

// headToStorage transitions the villager to CarryingToStorage toward the nearest storage.
// If no storage is found, drops wood and goes idle.
func (v *Villager) headToStorage(env *Env) {
	tx, ty, ok := nearestClearTileAdjacent(env.State.World, LogStorage, v.X, v.Y)
	if !ok {
		v.Task = VillagerIdle
		return
	}
	v.Task = VillagerCarryingToStorage
	v.TargetX, v.TargetY = tx, ty
	v.path = nil
}

// headToHouse transitions the villager to DeliveringToHouse toward the nearest house foundation.
// Returns false if no house foundation is found.
func (v *Villager) headToHouse(env *Env) bool {
	tx, ty, ok := nearestClearTileAdjacent(env.State.World, FoundationHouse, v.X, v.Y)
	if !ok {
		return false
	}
	v.Task = VillagerDeliveringToHouse
	v.TargetX, v.TargetY = tx, ty
	v.path = nil
	return true
}

// move advances the villager one step along its cached A* path toward (TargetX, TargetY).
// Recomputes the path if nil (new target) or if the next step becomes blocked.
func (v *Villager) move(world *World) {
	if v.path == nil {
		v.path = findPath(world, v.X, v.Y, v.TargetX, v.TargetY)
		if v.path == nil {
			return // unreachable; try again next tick
		}
	}
	if len(v.path) == 0 {
		return // already at goal
	}
	next := v.path[0]
	tile := world.TileAt(next.X, next.Y)
	if tile == nil || tile.Structure != NoStructure {
		v.path = nil // next step blocked; recompute next tick
		return
	}
	v.X, v.Y = next.X, next.Y
	v.path = v.path[1:]
}

// findNearestTree returns the world coordinates of the closest Forest tile with
// TreeSize > 0, measured by Euclidean distance from (fromX, fromY).
// Expands Chebyshev rings outward; stops once all remaining tiles are provably
// farther than the current best. Returns ok=false if none found.
func findNearestTree(world *World, fromX, fromY int) (x, y int, ok bool) {
	bestDist2 := 0
	found := false
	maxR := world.Width + world.Height
	for r := 0; r <= maxR; r++ {
		// All tiles at Chebyshev distance r have Euclidean distance >= r.
		// Once r^2 >= bestDist2 no closer tile can exist.
		if found && r*r >= bestDist2 {
			break
		}
		chebyshevRingDo(fromX, fromY, r, func(tx, ty int) {
			if !world.InBounds(tx, ty) {
				return
			}
			tile := world.TileAt(tx, ty)
			if tile == nil || tile.Terrain != Forest || tile.TreeSize <= 0 {
				return
			}
			dx, dy := tx-fromX, ty-fromY
			d2 := dx*dx + dy*dy
			if !found || d2 < bestDist2 {
				bestDist2 = d2
				x, y = tx, ty
				found = true
			}
		})
	}
	return x, y, found
}

// nearestClearTileAdjacent returns the clear tile (no structure, in-bounds) on the
// perimeter of the nearest instance of stype to (fromX, fromY).
// Uses structureInstanceIndex for O(instances) iteration over the footprint perimeter.
// Returns ok=false if no such tile exists.
func nearestClearTileAdjacent(world *World, stype StructureType, fromX, fromY int) (tx, ty int, ok bool) {
	bestDist2 := 0
	found := false
	for origin := range world.structureInstanceIndex[stype] {
		entry, hasEntry := world.structureIndex[origin]
		if !hasEntry {
			continue
		}
		fw, fh := entry.Def.Footprint()
		forFootprintCardinalNeighbors(origin.X, origin.Y, fw, fh, func(px, py int) {
			tile := world.TileAt(px, py)
			if tile == nil || tile.Structure != NoStructure {
				return
			}
			dx, dy := px-fromX, py-fromY
			d2 := dx*dx + dy*dy
			if !found || d2 < bestDist2 {
				bestDist2 = d2
				tx, ty = px, py
				found = true
			}
		})
	}
	return tx, ty, found
}

// forFootprintCardinalNeighbors calls f for each tile that is cardinally
// (orthogonally) adjacent to the w×h footprint with top-left at (fx, fy).
// Corner tiles of the Chebyshev border are excluded because they are only
// diagonally adjacent and villagers cannot interact with a structure from there.
func forFootprintCardinalNeighbors(fx, fy, fw, fh int, f func(x, y int)) {
	// Top and bottom edges (no corners).
	for x := fx; x < fx+fw; x++ {
		f(x, fy-1)
		f(x, fy+fh)
	}
	// Left and right edges.
	for y := fy; y < fy+fh; y++ {
		f(fx-1, y)
		f(fx+fw, y)
	}
}

// storageOriginAdjacent returns the origin of a LogStorage tile cardinally adjacent
// to (x, y), using the structureIndex for lookup. Returns ok=false if none found.
func storageOriginAdjacent(world *World, x, y int) (Point, bool) {
	for _, d := range [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
		nx, ny := x+d[0], y+d[1]
		entry, found := world.structureIndex[Point{nx, ny}]
		if !found {
			continue
		}
		tile := world.TileAt(entry.Origin.X, entry.Origin.Y)
		if tile != nil && tile.Structure == LogStorage {
			return entry.Origin, true
		}
	}
	return Point{}, false
}

// foundationHouseOriginAdjacent returns the origin of a FoundationHouse tile cardinally
// adjacent to (x, y), using the structureIndex for lookup. Returns ok=false if none found.
func foundationHouseOriginAdjacent(world *World, x, y int) (Point, bool) {
	for _, d := range [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
		nx, ny := x+d[0], y+d[1]
		entry, found := world.structureIndex[Point{nx, ny}]
		if !found {
			continue
		}
		tile := world.TileAt(entry.Origin.X, entry.Origin.Y)
		if tile != nil && tile.Structure == FoundationHouse {
			return entry.Origin, true
		}
	}
	return Point{}, false
}
