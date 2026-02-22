package game

// ResourceType identifies a storable resource.
type ResourceType int

// Resource types.
const (
	Wood ResourceType = iota
)

// StorageInstance tracks one storage structure's fill level.
type StorageInstance struct {
	Capacity int
	Stored   int
}

// ResourceStorage aggregates all storage instances for one resource type.
type ResourceStorage struct {
	Instances []*StorageInstance
}

// Total returns total stored across all instances.
func (r *ResourceStorage) Total() int {
	total := 0
	for _, inst := range r.Instances {
		total += inst.Stored
	}
	return total
}

// Deposit adds up to amount into the first available (non-full) instance.
// Returns the amount actually deposited.
func (r *ResourceStorage) Deposit(amount int) int {
	for _, inst := range r.Instances {
		space := inst.Capacity - inst.Stored
		if space <= 0 {
			continue
		}
		if amount > space {
			amount = space
		}
		inst.Stored += amount
		return amount
	}
	return 0
}

// AddInstance registers a new storage instance with the given capacity.
func (r *ResourceStorage) AddInstance(capacity int) *StorageInstance {
	inst := &StorageInstance{Capacity: capacity}
	r.Instances = append(r.Instances, inst)
	return inst
}
