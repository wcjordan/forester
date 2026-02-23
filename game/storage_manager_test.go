package game

import "testing"

func TestStorageManagerRoundTrip(t *testing.T) {
	// Build a world with two LogStorage instances at distinct origins.
	w := NewWorld(20, 20)
	originA := Point{2, 2}
	originB := Point{10, 10}
	w.SetStructure(originA.X, originA.Y, 4, 4, LogStorage)
	w.IndexStructure(originA.X, originA.Y, 4, 4, logStorageDef{})
	w.SetStructure(originB.X, originB.Y, 4, 4, LogStorage)
	w.IndexStructure(originB.X, originB.Y, 4, 4, logStorageDef{})

	// Register both with a manager and deposit into A only.
	m := NewStorageManager()
	m.Register(originA, Wood, LogStorageCapacity)
	m.Register(originB, Wood, LogStorageCapacity)
	deposited := m.DepositAt(originA, 7)
	if deposited != 7 {
		t.Fatalf("DepositAt returned %d, want 7", deposited)
	}

	// Save, then reconstruct into a fresh manager.
	saved := m.SaveData()
	m2 := NewStorageManager()
	m2.LoadFrom(saved, w)

	// Totals should match.
	if m2.Total(Wood) != m.Total(Wood) {
		t.Errorf("Total(Wood) after LoadFrom = %d, want %d", m2.Total(Wood), m.Total(Wood))
	}

	// Per-origin stored amounts should match.
	instA2 := m2.FindByOrigin(originA)
	if instA2 == nil {
		t.Fatal("FindByOrigin(originA) returned nil after LoadFrom")
	}
	if instA2.Stored != 7 {
		t.Errorf("instA.Stored after LoadFrom = %d, want 7", instA2.Stored)
	}

	instB2 := m2.FindByOrigin(originB)
	if instB2 == nil {
		t.Fatal("FindByOrigin(originB) returned nil after LoadFrom")
	}
	if instB2.Stored != 0 {
		t.Errorf("instB.Stored after LoadFrom = %d, want 0", instB2.Stored)
	}

	// Capacities should be restored from the structure def.
	if instA2.Capacity != LogStorageCapacity {
		t.Errorf("instA.Capacity after LoadFrom = %d, want %d", instA2.Capacity, LogStorageCapacity)
	}
}
