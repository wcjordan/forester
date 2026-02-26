package game

// StorageState is the serializable truth for all storage structures.
// Amounts maps the origin (top-left corner) of each storage structure to
// the amount currently stored. Resource type and capacity are derived on
// load from the world's StructureIndex via StorageDef.
type StorageState struct {
	Amounts map[Point]int // origin → stored amount
}

// StorageManager manages storage instances at runtime.
// It owns the live amounts (truth) and derived lookup structures.
type StorageManager struct {
	amounts    map[Point]int                     // live truth: origin → stored amount
	byOrigin   map[Point]*StorageInstance        // derived: origin → instance
	byResource map[ResourceType]*ResourceStorage // derived: resource → aggregator
}

// NewStorageManager creates an empty StorageManager.
func NewStorageManager() *StorageManager {
	return &StorageManager{
		amounts:    make(map[Point]int),
		byOrigin:   make(map[Point]*StorageInstance),
		byResource: make(map[ResourceType]*ResourceStorage),
	}
}

// Register creates a new storage instance for the structure at origin.
// Called from StructureDef.OnBuilt when a storage structure is completed.
func (m *StorageManager) Register(origin Point, resource ResourceType, capacity int) {
	inst := &StorageInstance{Resource: resource, Capacity: capacity, Stored: 0}
	m.byOrigin[origin] = inst
	m.amounts[origin] = 0
	if m.byResource[resource] == nil {
		m.byResource[resource] = &ResourceStorage{}
	}
	m.byResource[resource].Instances = append(m.byResource[resource].Instances, inst)
}

// DepositAt deposits up to amount into the instance at origin.
// Returns the amount actually deposited. Keeps amounts in sync with the instance.
func (m *StorageManager) DepositAt(origin Point, amount int) int {
	inst := m.byOrigin[origin]
	if inst == nil {
		return 0
	}
	deposited := inst.Deposit(amount)
	m.amounts[origin] += deposited
	return deposited
}

// FindByOrigin returns the storage instance at the given origin, or nil if none.
func (m *StorageManager) FindByOrigin(origin Point) *StorageInstance {
	return m.byOrigin[origin]
}

// TotalCapacity returns the sum of capacities across all instances for the given resource.
func (m *StorageManager) TotalCapacity(r ResourceType) int {
	rs := m.byResource[r]
	if rs == nil {
		return 0
	}
	total := 0
	for _, inst := range rs.Instances {
		total += inst.Capacity
	}
	return total
}

// Total returns the total amount stored across all instances for the given resource.
func (m *StorageManager) Total(r ResourceType) int {
	rs := m.byResource[r]
	if rs == nil {
		return 0
	}
	return rs.Total()
}

// SaveData returns a snapshot of the current storage truth.
func (m *StorageManager) SaveData() StorageState {
	snap := make(map[Point]int, len(m.amounts))
	for k, v := range m.amounts {
		snap[k] = v
	}
	return StorageState{Amounts: snap}
}

// LoadFrom rebuilds derived structures from saved storage state and the world.
// It uses the world's StructureIndex to find storage structure origins and
// queries each def's StorageDef implementation for resource type and capacity.
// Origins in the world that are not in saved state are initialized with 0.
func (m *StorageManager) LoadFrom(s StorageState, world *World) {
	m.amounts = make(map[Point]int)
	m.byOrigin = make(map[Point]*StorageInstance)
	m.byResource = make(map[ResourceType]*ResourceStorage)

	seen := make(map[Point]bool)
	for _, entry := range world.StructureIndex {
		if seen[entry.Origin] {
			continue
		}
		seen[entry.Origin] = true
		sd, ok := entry.Def.(StorageDef)
		if !ok {
			continue
		}
		amount := s.Amounts[entry.Origin]
		resource := sd.StorageResource()
		capacity := sd.StorageCapacity()
		inst := &StorageInstance{Resource: resource, Capacity: capacity, Stored: amount}
		m.byOrigin[entry.Origin] = inst
		m.amounts[entry.Origin] = amount
		if m.byResource[resource] == nil {
			m.byResource[resource] = &ResourceStorage{}
		}
		m.byResource[resource].Instances = append(m.byResource[resource].Instances, inst)
	}
}
