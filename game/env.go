package game

import "math/rand"

// Env is the runtime context passed to StructureDef methods and upgrade Apply callbacks.
// State holds serializable truth data; Stores and Villagers hold derived runtime state.
// RNG is the shared game random source, used for deterministic card shuffling and spawning.
type Env struct {
	State     *State
	Stores    *StorageManager
	Villagers *VillagerManager
	RNG       *rand.Rand
}
