package game

// ResourceType identifies a storable resource.
type ResourceType int

// Resource types.
const (
	Wood ResourceType = iota
)

// StorageInstance tracks one storage structure's fill level.
type StorageInstance struct {
	Resource ResourceType
	Capacity int
	Stored   int
}

// Deposit adds up to amount into this instance, capped at remaining capacity.
// Returns the amount actually deposited.
func (si *StorageInstance) Deposit(amount int) int {
	space := si.Capacity - si.Stored
	if space <= 0 || amount <= 0 {
		return 0
	}
	if amount > space {
		amount = space
	}
	si.Stored += amount
	return amount
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

// AddInstance registers a new storage instance with the given resource type and capacity.
func (r *ResourceStorage) AddInstance(resource ResourceType, capacity int) *StorageInstance {
	inst := &StorageInstance{Resource: resource, Capacity: capacity}
	r.Instances = append(r.Instances, inst)
	return inst
}
