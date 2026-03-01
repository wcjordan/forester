package game

// XP award amounts per action type.
const (
	XPPerWoodChopped        = 1
	XPPerWoodDeposited      = 1
	XPBuildCompletePlayer   = 10
	XPBuildCompleteVillager = 20
)

// xpCardPool is the base set of upgrade IDs offered at every XP milestone.
// spawn_villager is added conditionally when an unoccupied house exists.
var xpCardPool = []string{"harvest_speed", "deposit_speed", "move_speed", "build_speed"}

// xpMilestoneAt returns the cumulative XP required to reach milestone n (0-indexed).
// Gaps grow by 25 each milestone: 50, 75, 100, 125, ...
// Formula: 50 + 50n + 25·n·(n+1)/2
func xpMilestoneAt(n int) int {
	return 50 + 50*n + 25*n*(n+1)/2
}

// hasUnoccupiedHouse reports whether any built house has no villager assigned.
func hasUnoccupiedHouse(env *Env) bool {
	for _, occupied := range env.State.HouseOccupancy {
		if !occupied {
			return true
		}
	}
	return false
}

// pickCardOffer selects 3 distinct upgrade IDs from the eligible pool, shuffled with env.RNG.
// spawn_villager is included in the pool only when an unoccupied house exists.
func pickCardOffer(env *Env) []string {
	pool := make([]string, len(xpCardPool))
	copy(pool, xpCardPool)
	if hasUnoccupiedHouse(env) {
		pool = append(pool, "spawn_villager")
	}
	env.RNG.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	if len(pool) > 3 {
		pool = pool[:3]
	}
	return pool
}

// AwardXP adds amount to the player's XP total and queues a 3-card offer for each
// XP milestone crossed. Multiple milestones can be crossed in one call.
// Called from resource and structure packages that receive *Env.
func AwardXP(env *Env, amount int) {
	env.State.XP += amount
	for env.State.XP >= xpMilestoneAt(env.State.XPMilestoneIdx) {
		env.State.AddOffer(pickCardOffer(env))
		env.State.XPMilestoneIdx++
	}
}
