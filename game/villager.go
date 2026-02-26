package game

import (
	"math/rand"
	"time"
)

// VillagerMaxCarry is the maximum wood a villager can carry at once.
const VillagerMaxCarry = 5

// VillagerMoveCooldown is how often a villager takes one movement or harvest step.
const VillagerMoveCooldown = 300 * time.Millisecond

// VillagerTask identifies the villager's current activity.
type VillagerTask int

// VillagerTask values.
const (
	VillagerIdle              VillagerTask = iota // choosing next task
	VillagerWalkingToTree                         // moving to tree tile to harvest
	VillagerCarryingToStorage                     // carrying harvested wood to log storage
	VillagerWalkingToStorage                      // walking to log storage to fetch wood
	VillagerDeliveringToHouse                     // carrying fetched wood to a house
)

// Villager is an autonomous agent that collects and delivers wood.
type Villager struct {
	X, Y         int
	Wood         int
	Task         VillagerTask
	TargetX      int
	TargetY      int
	moveCooldown time.Time
}

// SpawnVillager appends a new idle villager at (x, y) to the state.
func (s *State) SpawnVillager(x, y int) {
	s.Villagers = append(s.Villagers, &Villager{X: x, Y: y})
}

// TickVillagers advances every villager by one step.
func (s *State) TickVillagers(env *Env, rng *rand.Rand, now time.Time) {
	for _, v := range s.Villagers {
		v.Tick(env, rng, now)
	}
}

// Tick advances this villager by one game step, gated on moveCooldown.
func (v *Villager) Tick(env *Env, rng *rand.Rand, now time.Time) {
	if now.Before(v.moveCooldown) {
		return
	}
	v.moveCooldown = now.Add(VillagerMoveCooldown)

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
				deposited := env.Stores.DepositAt(origin, v.Wood)
				v.Wood -= deposited
			} else {
				// Storage gone; drop wood and go idle.
				v.Wood = 0
			}
			v.Task = VillagerIdle
		} else {
			v.move(env.State.World)
		}

	case VillagerWalkingToStorage:
		if v.X == v.TargetX && v.Y == v.TargetY {
			origin, ok := storageOriginAdjacent(env.State.World, v.X, v.Y)
			if ok {
				fetched := env.Stores.WithdrawFrom(origin, VillagerMaxCarry)
				if fetched > 0 {
					v.Wood = fetched
					if !v.headToHouse(env) {
						// No house; return the wood and go idle.
						env.Stores.DepositAt(origin, v.Wood)
						v.Wood = 0
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
			v.Wood = 0 // wood consumed by house
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
	total := env.Stores.Total(Wood)
	storageCap := env.Stores.TotalCapacity(Wood)

	fillRatio := 0.0
	if storageCap > 0 {
		fillRatio = float64(total) / float64(storageCap)
		if fillRatio > 1 {
			fillRatio = 1
		}
	}

	wantChop := rng.Float64() > fillRatio
	if wantChop {
		if v.tryAssignChopTask(env) {
			return
		}
		v.tryAssignDeliverTask(env)
	} else {
		if v.tryAssignDeliverTask(env) {
			return
		}
		v.tryAssignChopTask(env)
	}
}

func (v *Villager) tryAssignChopTask(env *Env) bool {
	tx, ty, ok := findNearestTree(env.State.World, v.X, v.Y)
	if !ok {
		return false
	}
	v.Task = VillagerWalkingToTree
	v.TargetX, v.TargetY = tx, ty
	return true
}

func (v *Villager) tryAssignDeliverTask(env *Env) bool {
	if env.Stores.Total(Wood) == 0 {
		return false
	}
	if len(env.State.World.StructureTypeIndex[House]) == 0 {
		return false
	}
	tx, ty, ok := nearestClearTileAdjacent(env.State.World, LogStorage, v.X, v.Y)
	if !ok {
		return false
	}
	v.Task = VillagerWalkingToStorage
	v.TargetX, v.TargetY = tx, ty
	return true
}

// headToStorage transitions the villager to CarryingToStorage toward the nearest storage.
// If no storage is found, drops wood and goes idle.
func (v *Villager) headToStorage(env *Env) {
	tx, ty, ok := nearestClearTileAdjacent(env.State.World, LogStorage, v.X, v.Y)
	if !ok {
		v.Wood = 0
		v.Task = VillagerIdle
		return
	}
	v.Task = VillagerCarryingToStorage
	v.TargetX, v.TargetY = tx, ty
}

// headToHouse transitions the villager to DeliveringToHouse toward the nearest house.
// Returns false if no house is found.
func (v *Villager) headToHouse(env *Env) bool {
	tx, ty, ok := nearestClearTileAdjacent(env.State.World, House, v.X, v.Y)
	if !ok {
		return false
	}
	v.Task = VillagerDeliveringToHouse
	v.TargetX, v.TargetY = tx, ty
	return true
}

// move takes one cardinal step toward (TargetX, TargetY), preferring the axis
// with larger remaining distance. Falls back to the secondary axis if blocked.
func (v *Villager) move(world *World) {
	diffX := v.TargetX - v.X
	diffY := v.TargetY - v.Y

	type step struct{ dx, dy int }
	var primary, secondary step
	if abs(diffX) >= abs(diffY) {
		if diffX > 0 {
			primary = step{1, 0}
		} else {
			primary = step{-1, 0}
		}
		if diffY > 0 {
			secondary = step{0, 1}
		} else {
			secondary = step{0, -1}
		}
	} else {
		if diffY > 0 {
			primary = step{0, 1}
		} else {
			primary = step{0, -1}
		}
		if diffX > 0 {
			secondary = step{1, 0}
		} else {
			secondary = step{-1, 0}
		}
	}

	for _, s := range []step{primary, secondary} {
		if s.dx == 0 && s.dy == 0 {
			continue
		}
		nx, ny := v.X+s.dx, v.Y+s.dy
		tile := world.TileAt(nx, ny)
		if tile != nil && tile.Structure == NoStructure {
			v.X, v.Y = nx, ny
			return
		}
	}
}

// findNearestTree returns the world coordinates of the closest Forest tile with
// TreeSize > 0, measured from (fromX, fromY). Returns ok=false if none found.
func findNearestTree(world *World, fromX, fromY int) (x, y int, ok bool) {
	bestDist2 := 0
	found := false
	for wy := range world.Tiles {
		for wx := range world.Tiles[wy] {
			tile := &world.Tiles[wy][wx]
			if tile.Terrain != Forest || tile.TreeSize <= 0 {
				continue
			}
			dx, dy := wx-fromX, wy-fromY
			d2 := dx*dx + dy*dy
			if !found || d2 < bestDist2 {
				bestDist2 = d2
				x, y = wx, wy
				found = true
			}
		}
	}
	return x, y, found
}

// nearestClearTileAdjacent returns the clear tile (no structure, in-bounds) adjacent
// to the nearest instance of stype to (fromX, fromY).
// Returns ok=false if no such tile exists.
func nearestClearTileAdjacent(world *World, stype StructureType, fromX, fromY int) (tx, ty int, ok bool) {
	bestDist2 := 0
	found := false
	for pt := range world.StructureTypeIndex[stype] {
		for _, d := range [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
			nx, ny := pt.X+d[0], pt.Y+d[1]
			tile := world.TileAt(nx, ny)
			if tile == nil || tile.Structure != NoStructure {
				continue
			}
			dx, dy := nx-fromX, ny-fromY
			d2 := dx*dx + dy*dy
			if !found || d2 < bestDist2 {
				bestDist2 = d2
				tx, ty = nx, ny
				found = true
			}
		}
	}
	return tx, ty, found
}

// storageOriginAdjacent returns the origin of a LogStorage tile cardinally adjacent
// to (x, y), using the StructureIndex for lookup. Returns ok=false if none found.
func storageOriginAdjacent(world *World, x, y int) (Point, bool) {
	for _, d := range [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
		nx, ny := x+d[0], y+d[1]
		entry, found := world.StructureIndex[Point{nx, ny}]
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
