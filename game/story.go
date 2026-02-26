package game

// StoryBeat is a one-shot trigger that fires exactly once when its condition is
// met. Beats are evaluated in order; only the first unmet beat is checked per
// call to maybeAdvanceStory.
type StoryBeat struct {
	ID        string
	Condition func(env *Env) bool
	Action    func(env *Env)
}

// storyBeats is the ordered list of early-game story triggers.
// It is a package-level variable so tests can swap it out if needed.
var storyBeats = []StoryBeat{
	{
		// Spawn the first log storage foundation when the player's inventory is full.
		ID: "initial_log_storage",
		Condition: func(env *Env) bool {
			p := env.State.Player
			return p.Wood >= p.MaxCarry
		},
		Action: func(env *Env) {
			def := findStructureDefByFoundationType(FoundationLogStorage)
			if def != nil {
				env.State.spawnFoundationAt(def)
			}
		},
	},
	{
		// Spawn the first house foundation once enough wood has been deposited in storage.
		// 50 matches the house build cost so the player has enough on hand after depositing.
		ID: "initial_house",
		Condition: func(env *Env) bool {
			return env.Stores.Total(Wood) >= 50
		},
		Action: func(env *Env) {
			def := findStructureDefByFoundationType(FoundationHouse)
			if def != nil {
				env.State.spawnFoundationAt(def)
			}
		},
	},
}

// maybeAdvanceStory iterates storyBeats in order, skips completed beats, and
// fires the first beat whose condition is now met. At most one beat fires per call.
func (s *State) maybeAdvanceStory(env *Env) {
	if s.CompletedBeats == nil {
		s.CompletedBeats = make(map[string]bool)
	}
	for _, beat := range storyBeats {
		if s.CompletedBeats[beat.ID] {
			continue
		}
		if !beat.Condition(env) {
			continue
		}
		beat.Action(env)
		s.CompletedBeats[beat.ID] = true
		return
	}
}
