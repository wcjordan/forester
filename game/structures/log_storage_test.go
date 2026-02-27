package structures

import (
	"testing"
	"time"

	"forester/game"
)

func TestFoundationBuildMechanic(t *testing.T) {
	// Set up a world with an indexed foundation at (5,5) and player just west at (4,5).
	makeFoundationState := func(wood int) (*game.State, *game.StorageManager) {
		w := game.NewWorld(20, 20)
		w.SetStructure(5, 5, 4, 4, game.FoundationLogStorage)
		w.IndexStructure(5, 5, 4, 4, logStorageDef{})
		p := game.NewPlayer(4, 5)
		p.Inventory[game.Wood] = wood
		s := &game.State{
			Player:              p,
			World:               w,
			FoundationDeposited: make(map[game.Point]int),
		}
		return s, game.NewStorageManager()
	}

	t.Run("foundation blocks player movement", func(t *testing.T) {
		s, _ := makeFoundationState(0)
		s.Player.Move(1, 0, s.World, time.Now()) // try to step into (5,5) — foundation tile
		if s.Player.X != 4 {
			t.Errorf("player X = %d, want 4 (foundation should block movement)", s.Player.X)
		}
	})

	t.Run("adjacent deposit reduces player wood", func(t *testing.T) {
		s, stores := makeFoundationState(5)
		env := &game.Env{State: s, Stores: stores}
		s.TickAdjacentStructures(env, time.Now())
		if s.Player.Inventory[game.Wood] != 4 {
			t.Errorf("Inventory[Wood] = %d, want 4 after one deposit", s.Player.Inventory[game.Wood])
		}
		origin := game.Point{X: 5, Y: 5}
		if s.FoundationDeposited[origin] != 1 {
			t.Errorf("FoundationDeposited = %d, want 1", s.FoundationDeposited[origin])
		}
	})

	t.Run("deposit respects cooldown", func(t *testing.T) {
		s, stores := makeFoundationState(5)
		env := &game.Env{State: s, Stores: stores}
		t0 := time.Now()
		s.TickAdjacentStructures(env, t0)
		s.TickAdjacentStructures(env, t0) // same timestamp — cooldown blocks
		origin := game.Point{X: 5, Y: 5}
		if s.FoundationDeposited[origin] != 1 {
			t.Errorf("FoundationDeposited = %d, want 1 (second tick should be blocked by cooldown)", s.FoundationDeposited[origin])
		}
	})

	t.Run("foundation completes after BuildCost deposits", func(t *testing.T) {
		s, stores := makeFoundationState(logStorageBuildCost)
		env := &game.Env{State: s, Stores: stores}
		t0 := time.Now()
		for i := range logStorageBuildCost {
			s.TickAdjacentStructures(env, t0.Add(time.Duration(i)*(game.DepositTickInterval+time.Millisecond)))
		}
		if s.HasStructureOfType(game.FoundationLogStorage) {
			t.Error("FoundationLogStorage tiles should be gone after build completes")
		}
		if !s.HasStructureOfType(game.LogStorage) {
			t.Error("LogStorage tiles should exist after build completes")
		}
		if s.Player.Inventory[game.Wood] != 0 {
			t.Errorf("player Inventory[Wood] = %d, want 0 (all wood deposited)", s.Player.Inventory[game.Wood])
		}
	})
}

func TestTickAdjacentStructures(t *testing.T) {
	// makeDepositState creates a state with one LogStorage at origin (5,4) above the player.
	makeDepositState := func(wood int) (*game.State, *game.StorageManager) {
		w := game.NewWorld(10, 10)
		origin := game.Point{X: 5, Y: 4}
		w.SetStructure(origin.X, origin.Y, 4, 4, game.LogStorage)
		w.IndexStructure(origin.X, origin.Y, 4, 4, logStorageDef{})
		p := game.NewPlayer(5, 5)
		p.Inventory[game.Wood] = wood
		s := &game.State{Player: p, World: w, FoundationDeposited: make(map[game.Point]int)}
		stores := game.NewStorageManager()
		stores.Register(origin, game.Wood, logStorageCapacity)
		return s, stores
	}

	t.Run("no deposit when Wood is 0", func(t *testing.T) {
		s, stores := makeDepositState(0)
		env := &game.Env{State: s, Stores: stores}
		s.TickAdjacentStructures(env, time.Now())
		if stores.Total(game.Wood) != 0 {
			t.Errorf("Total(Wood) = %d, want 0 when Wood == 0", stores.Total(game.Wood))
		}
	})

	t.Run("no deposit when not adjacent to LogStorage", func(t *testing.T) {
		w := game.NewWorld(10, 10)
		p := game.NewPlayer(5, 5)
		p.Inventory[game.Wood] = 5
		s := &game.State{Player: p, World: w, FoundationDeposited: make(map[game.Point]int)}
		stores := game.NewStorageManager()
		env := &game.Env{State: s, Stores: stores}
		s.TickAdjacentStructures(env, time.Now())
		if stores.Total(game.Wood) != 0 {
			t.Errorf("Total(Wood) = %d, want 0 when not adjacent", stores.Total(game.Wood))
		}
	})

	t.Run("deposits one wood when adjacent", func(t *testing.T) {
		s, stores := makeDepositState(5)
		env := &game.Env{State: s, Stores: stores}
		s.TickAdjacentStructures(env, time.Now())
		if s.Player.Inventory[game.Wood] != 4 {
			t.Errorf("Inventory[Wood] = %d, want 4 after deposit", s.Player.Inventory[game.Wood])
		}
		if stores.Total(game.Wood) != 1 {
			t.Errorf("Total(Wood) = %d, want 1", stores.Total(game.Wood))
		}
	})

	t.Run("deposits one at a time with cooldown between ticks", func(t *testing.T) {
		s, stores := makeDepositState(3)
		env := &game.Env{State: s, Stores: stores}
		t0 := time.Now()
		s.TickAdjacentStructures(env, t0)
		s.TickAdjacentStructures(env, t0.Add(game.DepositTickInterval+time.Millisecond))
		if s.Player.Inventory[game.Wood] != 1 {
			t.Errorf("Inventory[Wood] = %d, want 1 after 2 deposits", s.Player.Inventory[game.Wood])
		}
		if stores.Total(game.Wood) != 2 {
			t.Errorf("Total(Wood) = %d, want 2", stores.Total(game.Wood))
		}
	})

	t.Run("two adjacent instances each trigger an interaction", func(t *testing.T) {
		// Player at (5,5). LogStorage A above (y=4), LogStorage B below (y=6).
		// Cooldowns are queued during interactions and committed after all are
		// processed, so both instances fire within the same tick.
		// Each instance has its own StorageByOrigin entry so deposits route correctly.
		w := game.NewWorld(20, 20)
		originA := game.Point{X: 5, Y: 4}
		originB := game.Point{X: 5, Y: 6}
		w.SetStructure(originA.X, originA.Y, 1, 1, game.LogStorage)
		w.IndexStructure(originA.X, originA.Y, 1, 1, logStorageDef{})
		w.SetStructure(originB.X, originB.Y, 1, 1, game.LogStorage)
		w.IndexStructure(originB.X, originB.Y, 1, 1, logStorageDef{})
		p := game.NewPlayer(5, 5)
		p.Inventory[game.Wood] = 5
		s := &game.State{Player: p, World: w, FoundationDeposited: make(map[game.Point]int)}
		stores := game.NewStorageManager()
		stores.Register(originA, game.Wood, logStorageCapacity)
		stores.Register(originB, game.Wood, logStorageCapacity)
		instA := stores.FindByOrigin(originA)
		instB := stores.FindByOrigin(originB)
		env := &game.Env{State: s, Stores: stores}
		s.TickAdjacentStructures(env, time.Now())
		// Two instances adjacent → two deposits, one per instance.
		if s.Player.Inventory[game.Wood] != 3 {
			t.Errorf("Inventory[Wood] = %d, want 3 after two-instance deposit", s.Player.Inventory[game.Wood])
		}
		if stores.Total(game.Wood) != 2 {
			t.Errorf("Total(Wood) = %d, want 2", stores.Total(game.Wood))
		}
		if instA.Stored != 1 {
			t.Errorf("instA.Stored = %d, want 1", instA.Stored)
		}
		if instB.Stored != 1 {
			t.Errorf("instB.Stored = %d, want 1", instB.Stored)
		}
	})
}

func TestDepositRoutesToSpecificInstance(t *testing.T) {
	// Two storages: A at (2,4) and B at (8,4). Player at (2,5) — adjacent to A only.
	// Deposit should go into A, not B.
	w := game.NewWorld(15, 10)
	originA := game.Point{X: 2, Y: 4}
	originB := game.Point{X: 8, Y: 4}
	w.SetStructure(originA.X, originA.Y, 1, 1, game.LogStorage)
	w.IndexStructure(originA.X, originA.Y, 1, 1, logStorageDef{})
	w.SetStructure(originB.X, originB.Y, 1, 1, game.LogStorage)
	w.IndexStructure(originB.X, originB.Y, 1, 1, logStorageDef{})
	p := game.NewPlayer(2, 5)
	p.Inventory[game.Wood] = 3
	s := &game.State{Player: p, World: w, FoundationDeposited: make(map[game.Point]int)}
	stores := game.NewStorageManager()
	stores.Register(originA, game.Wood, logStorageCapacity)
	stores.Register(originB, game.Wood, logStorageCapacity)
	instA := stores.FindByOrigin(originA)
	instB := stores.FindByOrigin(originB)
	env := &game.Env{State: s, Stores: stores}

	s.TickAdjacentStructures(env, time.Now())

	if instA.Stored != 1 {
		t.Errorf("instA.Stored = %d, want 1 (adjacent storage)", instA.Stored)
	}
	if instB.Stored != 0 {
		t.Errorf("instB.Stored = %d, want 0 (non-adjacent storage)", instB.Stored)
	}
}

func TestDepositRespectsInstanceCapacity(t *testing.T) {
	// Storage at capacity — deposit should be refused and no cooldown queued.
	w := game.NewWorld(10, 10)
	origin := game.Point{X: 5, Y: 4}
	w.SetStructure(origin.X, origin.Y, 1, 1, game.LogStorage)
	w.IndexStructure(origin.X, origin.Y, 1, 1, logStorageDef{})
	p := game.NewPlayer(5, 5)
	p.Inventory[game.Wood] = 5
	s := &game.State{Player: p, World: w, FoundationDeposited: make(map[game.Point]int)}
	stores := game.NewStorageManager()
	stores.Register(origin, game.Wood, 2)
	stores.DepositAt(origin, 2) // fill to capacity via the proper API
	inst := stores.FindByOrigin(origin)
	env := &game.Env{State: s, Stores: stores}

	s.TickAdjacentStructures(env, time.Now())

	if inst.Stored != 2 {
		t.Errorf("inst.Stored = %d, want 2 (full — no deposit)", inst.Stored)
	}
	if p.Inventory[game.Wood] != 5 {
		t.Errorf("player Inventory[Wood] = %d, want 5 (no deposit taken)", p.Inventory[game.Wood])
	}
}

func TestStorageManagerRoundTrip(t *testing.T) {
	// Build a world with two LogStorage instances at distinct origins.
	w := game.NewWorld(20, 20)
	originA := game.Point{X: 2, Y: 2}
	originB := game.Point{X: 10, Y: 10}
	w.SetStructure(originA.X, originA.Y, 4, 4, game.LogStorage)
	w.IndexStructure(originA.X, originA.Y, 4, 4, logStorageDef{})
	w.SetStructure(originB.X, originB.Y, 4, 4, game.LogStorage)
	w.IndexStructure(originB.X, originB.Y, 4, 4, logStorageDef{})

	// Register both with a manager and deposit into A only.
	m := game.NewStorageManager()
	m.Register(originA, game.Wood, logStorageCapacity)
	m.Register(originB, game.Wood, logStorageCapacity)
	deposited := m.DepositAt(originA, 7)
	if deposited != 7 {
		t.Fatalf("DepositAt returned %d, want 7", deposited)
	}

	// Save, then reconstruct into a fresh manager.
	saved := m.SaveData()
	m2 := game.NewStorageManager()
	m2.LoadFrom(saved, w)

	// Totals should match.
	if m2.Total(game.Wood) != m.Total(game.Wood) {
		t.Errorf("Total(Wood) after LoadFrom = %d, want %d", m2.Total(game.Wood), m.Total(game.Wood))
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
	if instA2.Capacity != logStorageCapacity {
		t.Errorf("instA.Capacity after LoadFrom = %d, want %d", instA2.Capacity, logStorageCapacity)
	}
}

func TestDepositCooldown(t *testing.T) {
	makeDepositState := func(wood int) (*game.State, *game.StorageManager) {
		w := game.NewWorld(10, 10)
		origin := game.Point{X: 5, Y: 4}
		w.SetStructure(origin.X, origin.Y, 4, 4, game.LogStorage) // storage above player
		w.IndexStructure(origin.X, origin.Y, 4, 4, logStorageDef{})
		p := game.NewPlayer(5, 5)
		p.Inventory[game.Wood] = wood
		s := &game.State{Player: p, World: w, FoundationDeposited: make(map[game.Point]int)}
		stores := game.NewStorageManager()
		stores.Register(origin, game.Wood, logStorageCapacity)
		return s, stores
	}

	t.Run("does not deposit when cooldown has not passed", func(t *testing.T) {
		s, stores := makeDepositState(5)
		env := &game.Env{State: s, Stores: stores}
		now := time.Now()
		s.Player.SetCooldown(game.Deposit, now.Add(time.Hour))
		s.TickAdjacentStructures(env, now)
		if stores.Total(game.Wood) != 0 {
			t.Errorf("Total(Wood) = %d, want 0 when cooldown active", stores.Total(game.Wood))
		}
	})

	t.Run("deposits when cooldown has passed", func(t *testing.T) {
		s, stores := makeDepositState(5) // Cooldowns zero value = expired
		env := &game.Env{State: s, Stores: stores}
		s.TickAdjacentStructures(env, time.Now())
		if stores.Total(game.Wood) != 1 {
			t.Errorf("Total(Wood) = %d, want 1", stores.Total(game.Wood))
		}
	})

	t.Run("sets cooldown after deposit", func(t *testing.T) {
		s, stores := makeDepositState(5)
		env := &game.Env{State: s, Stores: stores}
		now := time.Now()
		s.TickAdjacentStructures(env, now)
		if !s.Player.Cooldowns[game.Deposit].After(now) {
			t.Error("Cooldowns[Deposit] should be set to a future time after deposit")
		}
	})

	t.Run("does not set cooldown when nothing deposited", func(t *testing.T) {
		s, stores := makeDepositState(0) // no wood to deposit
		env := &game.Env{State: s, Stores: stores}
		s.TickAdjacentStructures(env, time.Now())
		var zero time.Time
		if s.Player.Cooldowns[game.Deposit] != zero {
			t.Error("Cooldowns[Deposit] should remain zero when nothing was deposited")
		}
	})
}
