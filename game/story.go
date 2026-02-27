package game

// storyBeat is a one-shot trigger that fires exactly once when its condition is
// met. At most one beat fires per call to maybeAdvanceStory; beats are evaluated
// in strict order — if an incomplete beat's condition is not yet true, evaluation
// stops so that later beats cannot complete out of order.
//
// Action returns true when the beat is complete and should not fire again.
// Returning false (e.g. when spawnFoundationAt finds no valid location) causes
// the beat to retry on the next tick.
type storyBeat struct {
	ID        string
	Condition func(env *Env) bool
	Action    func(env *Env) bool
}

// storyBeats is the ordered list of early-game story triggers.
// It is a package-level variable so tests can swap it out if needed.
//
// Beat ordering mirrors the intended narrative:
//  1. Spawn log storage foundation (player inventory full)
//  2. Reward first log storage completion (carry upgrade)
//  3. Spawn first house foundation (enough wood in storage)
//  4. Reward first house completion (build/deposit speed upgrades)
var storyBeats = []storyBeat{
	{
		// Spawn the first log storage foundation when the player's inventory is full.
		ID: "initial_log_storage",
		Condition: func(env *Env) bool {
			p := env.State.Player
			return p.Inventory[Wood] >= p.MaxCarry
		},
		Action: func(env *Env) bool {
			def := findStructureDefByFoundationType(FoundationLogStorage)
			if def == nil {
				return false
			}
			return spawnFoundationAt(env.State.World, env.State.Player.X, env.State.Player.Y, def)
		},
	},
	{
		// Queue the carry upgrade offer when the first log storage is completed.
		ID: "first_log_storage_built",
		Condition: func(env *Env) bool {
			return env.State.World.HasStructureOfType(LogStorage)
		},
		Action: func(env *Env) bool {
			env.State.AddOffer([]string{"carry_capacity"})
			return true
		},
	},
	{
		// Spawn the first house foundation once enough wood has been deposited in storage.
		// NOTE: 50 matches houseBuildCost in game/structures/house.go so the player has
		// enough wood on hand after depositing to immediately build the house.
		// If houseBuildCost changes, update this threshold to match.
		ID: "initial_house",
		Condition: func(env *Env) bool {
			return env.Stores.Total(Wood) >= 50
		},
		Action: func(env *Env) bool {
			def := findStructureDefByFoundationType(FoundationHouse)
			if def == nil {
				return false
			}
			return spawnFoundationAt(env.State.World, env.State.Player.X, env.State.Player.Y, def)
		},
	},
	{
		// Queue the build/deposit speed upgrade offer when the first house is completed.
		ID: "first_house_built",
		Condition: func(env *Env) bool {
			return env.State.World.HasStructureOfType(House)
		},
		Action: func(env *Env) bool {
			env.State.AddOffer([]string{"build_speed", "deposit_speed"})
			return true
		},
	},
}

// maybeAdvanceStory evaluates storyBeats in order and fires the first beat
// whose condition is met. Completed beats are skipped. If an incomplete beat's
// condition is not yet true, evaluation stops (strict ordering). At most one
// beat fires per call; a beat is only marked complete when its action returns true.
func maybeAdvanceStory(env *Env) {
	if env.State.completedBeats == nil {
		env.State.completedBeats = make(map[string]bool)
	}
	for _, beat := range storyBeats {
		if env.State.completedBeats[beat.ID] {
			continue
		}
		// Strict ordering: stop at the first incomplete beat whose condition is not met.
		if !beat.Condition(env) {
			return
		}
		if beat.Action(env) {
			env.State.completedBeats[beat.ID] = true
		}
		return
	}
}
