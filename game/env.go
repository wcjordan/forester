package game

// Env is the runtime context passed to StructureDef methods.
// State holds serializable truth data; Stores holds derived runtime storage state.
type Env struct {
	State  *State
	Stores *StorageManager
}
