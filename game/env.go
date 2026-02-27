package game

// Env is the runtime context passed to StructureDef methods.
// State holds serializable truth data; Stores and Villagers hold derived runtime state.
type Env struct {
	State     *State
	Stores    *StorageManager
	Villagers *VillagerManager
}
