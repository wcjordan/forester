package game

import (
	"fmt"
	"sort"
)

// storyBeat is a one-shot trigger that fires exactly once when its condition is
// met. At most one beat fires per call to maybeAdvanceStory; beats are evaluated
// in strict order — if an incomplete beat's condition is not yet true, evaluation
// stops so that later beats cannot complete out of order.
//
// Action returns true when the beat is complete and should not fire again.
// Returning false (e.g. when spawnFoundationAt finds no valid location) causes
// the beat to retry on the next tick.
type storyBeat struct {
	Order     int
	ID        string
	Condition func(env *Env) bool
	Action    func(env *Env) bool
}

// storyBeats is the ordered list of story triggers, sorted by Order.
// Beats are registered via RegisterStoryBeat, typically from init() functions
// in game/structures subpackages. It is a package-level variable so tests can
// swap it out if needed.
var storyBeats []storyBeat

// RegisterStoryBeat adds a story beat to the ordered sequence. The order field
// controls narrative sequencing: lower values fire before higher values.
// Call from an init() function in an external package (e.g. game/structures).
// The beat is inserted in sorted position so init() call order does not matter.
// Panics if id is already registered, matching the safety guarantee of RegisterStructure.
func RegisterStoryBeat(order int, id string, condition, action func(*Env) bool) {
	for _, b := range storyBeats {
		if b.ID == id {
			panic(fmt.Sprintf("RegisterStoryBeat: ID %q already registered", id))
		}
	}
	beat := storyBeat{Order: order, ID: id, Condition: condition, Action: action}
	i := sort.Search(len(storyBeats), func(i int) bool {
		return storyBeats[i].Order >= order
	})
	storyBeats = append(storyBeats, storyBeat{})
	copy(storyBeats[i+1:], storyBeats[i:])
	storyBeats[i] = beat
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
