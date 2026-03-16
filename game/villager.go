package game

import (
	"math/rand"
	"time"

	"forester/game/geom"
)

// VillagerMaxCarry is the maximum wood a villager can carry at once.
const VillagerMaxCarry = 5

// villagerDepositTypes lists StructureTypes that villagers treat as wood-deposit destinations.
// Populated via RegisterVillagerDepositType from external packages (e.g. game/structures).
var villagerDepositTypes []StructureType

// villagerDeliveryTypes lists StructureTypes that villagers treat as build-delivery targets.
// Populated via RegisterVillagerDeliveryType from external packages (e.g. game/structures).
var villagerDeliveryTypes []StructureType

// RegisterVillagerDepositType marks stype as a destination for villagers depositing harvested wood.
// Call from an init() in an external package (e.g. game/structures).
func RegisterVillagerDepositType(stype StructureType) {
	villagerDepositTypes = append(villagerDepositTypes, stype)
}

// RegisterVillagerDeliveryType marks stype as a foundation target for villagers delivering wood.
// Call from an init() in an external package (e.g. game/structures).
func RegisterVillagerDeliveryType(stype StructureType) {
	villagerDeliveryTypes = append(villagerDeliveryTypes, stype)
}

// villagerMoveCooldown is how often a villager takes one movement or harvest step.
const villagerMoveCooldown = 300 * time.Millisecond

// villagerPathMaxFailures is the number of consecutive FindPath misses before a villager
// gives up on its current target and returns to VillagerIdle.
const villagerPathMaxFailures = 5

// villagerPathBackoffBase is the initial retry delay after the first FindPath failure.
// Each subsequent failure doubles the delay, capped at villagerPathBackoffMax.
const villagerPathBackoffBase = villagerMoveCooldown

// villagerPathBackoffMax caps the per-failure retry delay.
const villagerPathBackoffMax = 5 * time.Second

// VillagerTask identifies the villager's current activity.
type VillagerTask int

// VillagerTask values.
const (
	VillagerIdle              VillagerTask = iota // choosing next task
	VillagerFindTree                              // finding a tree to harvest
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
	case VillagerFindTree:
		return "FindTree"
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
	path         []point // nil = recompute; []point{} = at goal
	pathFailures int     // consecutive FindPath misses for current target
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

// SpawnVillagerAtHouse spawns a villager at the first clear tile on the Chebyshev
// border of the given house, then marks the house as occupied in HouseOccupancy.
// Returns true if a villager was spawned.
func SpawnVillagerAtHouse(env *Env, origin geom.Point) bool {
	entry, ok := env.State.World.structureIndex[point(origin)]
	if !ok {
		return false
	}
	fw, fh := entry.Def.Footprint()
	spawned := false
	geom.FootprintBorderDo(origin.X, origin.Y, fw, fh, func(bx, by int) {
		if spawned {
			return
		}
		if env.State.World.IsBlocked(bx, by) {
			return
		}
		env.Villagers.Spawn(bx, by)
		env.State.HouseOccupancy[origin] = true
		spawned = true
	})
	return spawned
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

	case VillagerFindTree:
		if !v.tryAssignChopTask(env) {
			if v.Wood > 0 {
				v.headToStorage(env)
			} else {
				v.Task = VillagerIdle
			}
		}

	case VillagerWalkingToTree:
		// Target no longer valid (exhausted or gone).
		if !env.State.World.isHarvestable(v.TargetX, v.TargetY) {
			v.Task = VillagerFindTree
			return
		}
		if v.X == v.TargetX && v.Y == v.TargetY {
			// On the tree tile: harvest one wood.
			tile := env.State.World.TileAt(v.TargetX, v.TargetY)
			take := min(1, tile.TreeSize, VillagerMaxCarry-v.Wood)
			tile.TreeSize -= take
			v.Wood += take
			if v.Wood >= VillagerMaxCarry {
				v.headToStorage(env)
			} else if tile.TreeSize == 0 {
				// Tree exhausted but not full; find another tree.
				v.Task = VillagerFindTree
			}
		} else {
			v.move(env.State.World, now)
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
			v.move(env.State.World, now)
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
				} else if v.Wood > 0 {
					// Couldn't fetch more but already carrying wood; deliver it.
					if !v.headToHouse(env) {
						v.Wood -= env.Stores.DepositAt(origin, v.Wood)
						v.Task = VillagerIdle
					}
					return
				}
			}
			v.Task = VillagerIdle
		} else {
			v.move(env.State.World, now)
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
							AwardXP(env, XPBuildCompleteVillager)
							FinalizeFoundation(env, entry.Def, foundOrigin)
						}
					}
				}
			}
			v.Task = VillagerIdle
		} else {
			v.move(env.State.World, now)
		}
	}
}

// pickTask selects the villager's next task probabilistically based on log storage fill.
// P(chop wood → storage) = 1 - fillRatio; P(fetch storage → house) = fillRatio.
// Falls back to the other task when the preferred one has no valid target.
func (v *Villager) pickTask(env *Env, rng *rand.Rand) {
	// Already carrying wood: deliver to a house or deposit before doing anything else.
	if v.Wood > 0 {
		if v.headToHouse(env) {
			return
		}
		v.headToStorage(env)
		return
	}

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

// setTarget assigns a new movement target and resets path state so A* reruns from scratch.
func (v *Villager) setTarget(tx, ty int) {
	v.TargetX, v.TargetY = tx, ty
	v.path = nil
	v.pathFailures = 0
}

func (v *Villager) tryAssignChopTask(env *Env) bool {
	tx, ty, ok := findNearbyTree(env.State.World, v.X, v.Y)
	if !ok {
		return false
	}
	v.Task = VillagerWalkingToTree
	v.setTarget(tx, ty)
	return true
}

func (v *Villager) tryAssignDeliverTask(env *Env) bool {
	if v.Wood >= VillagerMaxCarry {
		return false
	}
	if env.Stores.Total(Wood) == 0 {
		return false
	}
	hasDeliveryTarget := false
	for _, dt := range villagerDeliveryTypes {
		if len(env.State.World.StructureTypeIndex[dt]) > 0 {
			hasDeliveryTarget = true
			break
		}
	}
	if !hasDeliveryTarget {
		return false
	}
	for _, depositType := range villagerDepositTypes {
		tx, ty, ok := nearestClearTileAdjacent(env.State.World, depositType, v.X, v.Y, nil)
		if ok {
			v.Task = VillagerWalkingToStorage
			v.setTarget(tx, ty)
			return true
		}
	}
	return false
}

// headToStorage transitions the villager to CarryingToStorage toward the nearest non-full storage.
// If no storage with space is found, goes idle.
func (v *Villager) headToStorage(env *Env) {
	isFull := func(origin point) bool {
		inst := env.Stores.FindByOrigin(origin)
		return inst == nil || inst.Stored >= inst.Capacity
	}
	for _, depositType := range villagerDepositTypes {
		tx, ty, ok := nearestClearTileAdjacent(env.State.World, depositType, v.X, v.Y, isFull)
		if ok {
			v.Task = VillagerCarryingToStorage
			v.setTarget(tx, ty)
			return
		}
	}
	v.Task = VillagerIdle
}

// headToHouse transitions the villager to DeliveringToHouse toward the nearest
// registered delivery target (e.g. house foundation). Returns false if none found.
func (v *Villager) headToHouse(env *Env) bool {
	for _, deliveryType := range villagerDeliveryTypes {
		tx, ty, ok := nearestClearTileAdjacent(env.State.World, deliveryType, v.X, v.Y, nil)
		if ok {
			v.Task = VillagerDeliveringToHouse
			v.setTarget(tx, ty)
			return true
		}
	}
	return false
}

// move advances the villager one step along its cached A* path toward (TargetX, TargetY).
// Recomputes the path if nil (new target) or if the next step becomes blocked.
// On repeated FindPath failures it applies exponential backoff and resets to idle after
// villagerPathMaxFailures consecutive misses.
func (v *Villager) move(world *World, now time.Time) {
	if v.path == nil {
		v.path = geom.FindPath(world, v.X, v.Y, v.TargetX, v.TargetY)
		if v.path == nil {
			v.pathFailures++
			if v.pathFailures >= villagerPathMaxFailures {
				v.Task = VillagerIdle
				v.pathFailures = 0
				return
			}
			backoff := min(villagerPathBackoffBase<<(v.pathFailures-1), villagerPathBackoffMax)
			v.moveCooldown = now.Add(backoff)
			return
		}
		v.pathFailures = 0
	}
	if len(v.path) == 0 {
		return // already at goal
	}
	next := v.path[0]
	if world.IsBlocked(next.X, next.Y) {
		v.path = nil // next step blocked; recompute next tick
		return
	}
	v.X, v.Y = next.X, next.Y
	v.path = v.path[1:]
	if t := world.TileAt(v.X, v.Y); t != nil && isRoadEligible(t) {
		t.WalkCount++
	}
}

// findNearbyTree returns the world coordinates of the first Forest tile with
// TreeSize > 0 found by expanding Chebyshev rings outward from (fromX, fromY).
// Returns ok=false if none found.
func findNearbyTree(world *World, fromX, fromY int) (x, y int, ok bool) {
	maxR := world.Width + world.Height
	return geom.SpiralSearchDo(fromX, fromY, maxR, func(tx, ty int) bool {
		if !world.InBounds(tx, ty) {
			return false
		}
		tile := world.TileAt(tx, ty)
		return tile != nil && tile.Terrain == Forest && tile.TreeSize > 0
	})
}

// nearestClearTileAdjacent returns the clear tile (no structure, in-bounds) on the
// perimeter of the nearest instance of stype to (fromX, fromY).
// Uses structureInstanceIndex for O(instances) iteration over the footprint perimeter.
// skip, if non-nil, is called with the structure origin; returning true skips that instance.
// Returns ok=false if no such tile exists.
func nearestClearTileAdjacent(world *World, stype StructureType, fromX, fromY int, skip func(point) bool) (tx, ty int, ok bool) {
	bestDist2 := 0
	found := false
	for origin := range world.structureInstanceIndex[stype] {
		if skip != nil && skip(origin) {
			continue
		}
		entry, hasEntry := world.structureIndex[origin]
		if !hasEntry {
			continue
		}
		fw, fh := entry.Def.Footprint()
		geom.ForFootprintCardinalNeighbors(origin.X, origin.Y, fw, fh, func(px, py int) {
			if world.IsBlocked(px, py) {
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

// adjacentStructureOriginOfType returns the origin of the first structure of stype
// cardinally adjacent to (x, y). Returns ok=false if none found.
func adjacentStructureOriginOfType(world *World, x, y int, stype StructureType) (point, bool) {
	for _, d := range [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
		nx, ny := x+d[0], y+d[1]
		entry, found := world.structureIndex[point{X: nx, Y: ny}]
		if !found {
			continue
		}
		tile := world.TileAt(entry.Origin.X, entry.Origin.Y)
		if tile != nil && tile.Structure == stype {
			return entry.Origin, true
		}
	}
	return point{}, false
}

// storageOriginAdjacent returns the origin of any registered deposit-target structure
// cardinally adjacent to (x, y). Returns ok=false if none found.
func storageOriginAdjacent(world *World, x, y int) (point, bool) {
	for _, stype := range villagerDepositTypes {
		if origin, ok := adjacentStructureOriginOfType(world, x, y, stype); ok {
			return origin, true
		}
	}
	return point{}, false
}

// foundationHouseOriginAdjacent returns the origin of any registered delivery-target
// structure cardinally adjacent to (x, y). Returns ok=false if none found.
func foundationHouseOriginAdjacent(world *World, x, y int) (point, bool) {
	for _, stype := range villagerDeliveryTypes {
		if origin, ok := adjacentStructureOriginOfType(world, x, y, stype); ok {
			return origin, true
		}
	}
	return point{}, false
}
