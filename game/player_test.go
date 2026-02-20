package game

import "testing"

func TestMovePlayer(t *testing.T) {
	w := NewWorld(10, 10)
	p := NewPlayer(5, 5)

	p.MovePlayer(1, 0, w)
	if p.X != 6 || p.Y != 5 {
		t.Errorf("after move right: got (%d,%d), want (6,5)", p.X, p.Y)
	}

	p.MovePlayer(0, 1, w)
	if p.X != 6 || p.Y != 6 {
		t.Errorf("after move down: got (%d,%d), want (6,6)", p.X, p.Y)
	}

	p.MovePlayer(-1, 0, w)
	if p.X != 5 || p.Y != 6 {
		t.Errorf("after move left: got (%d,%d), want (5,6)", p.X, p.Y)
	}

	p.MovePlayer(0, -1, w)
	if p.X != 5 || p.Y != 5 {
		t.Errorf("after move up: got (%d,%d), want (5,5)", p.X, p.Y)
	}
}

func TestMovePlayerBounds(t *testing.T) {
	w := NewWorld(10, 10)

	// At left/top edge — cannot move further.
	p := NewPlayer(0, 0)
	p.MovePlayer(-1, 0, w)
	if p.X != 0 {
		t.Errorf("moved past left edge: X = %d, want 0", p.X)
	}
	p.MovePlayer(0, -1, w)
	if p.Y != 0 {
		t.Errorf("moved past top edge: Y = %d, want 0", p.Y)
	}

	// At right/bottom edge — cannot move further.
	p = NewPlayer(9, 9)
	p.MovePlayer(1, 0, w)
	if p.X != 9 {
		t.Errorf("moved past right edge: X = %d, want 9", p.X)
	}
	p.MovePlayer(0, 1, w)
	if p.Y != 9 {
		t.Errorf("moved past bottom edge: Y = %d, want 9", p.Y)
	}
}

func TestHarvestAdjacent(t *testing.T) {
	// Helper: make a small world with a Forest tile adjacent to player.
	makeWorld := func(terrain TerrainType, treeSize int) (*World, *Tile) {
		w := NewWorld(5, 5)
		w.Tiles[1][2] = Tile{Terrain: terrain, TreeSize: treeSize} // above player at (2,2)
		return w, &w.Tiles[1][2]
	}

	t.Run("harvests adjacent forest tile", func(t *testing.T) {
		w, tile := makeWorld(Forest, 5)
		p := NewPlayer(2, 2)
		p.HarvestAdjacent(w)
		if p.Wood != 1 {
			t.Errorf("Wood = %d, want 1", p.Wood)
		}
		if tile.TreeSize != 4 {
			t.Errorf("TreeSize = %d, want 4", tile.TreeSize)
		}
		if tile.Terrain != Forest {
			t.Errorf("Terrain = %v, want Forest (tree not depleted)", tile.Terrain)
		}
	})

	t.Run("converts to stump when tree depleted", func(t *testing.T) {
		w, tile := makeWorld(Forest, 1)
		p := NewPlayer(2, 2)
		p.HarvestAdjacent(w)
		if p.Wood != 1 {
			t.Errorf("Wood = %d, want 1", p.Wood)
		}
		if tile.TreeSize != 0 {
			t.Errorf("TreeSize = %d, want 0", tile.TreeSize)
		}
		if tile.Terrain != Stump {
			t.Errorf("Terrain = %v, want Stump", tile.Terrain)
		}
	})

	t.Run("does not harvest from stump", func(t *testing.T) {
		w, tile := makeWorld(Stump, 0)
		p := NewPlayer(2, 2)
		p.HarvestAdjacent(w)
		if p.Wood != 0 {
			t.Errorf("Wood = %d, want 0 (stump should not yield wood)", p.Wood)
		}
		if tile.Terrain != Stump {
			t.Errorf("Terrain changed from Stump unexpectedly")
		}
	})

	t.Run("does not harvest from grassland", func(t *testing.T) {
		w, _ := makeWorld(Grassland, 0)
		p := NewPlayer(2, 2)
		p.HarvestAdjacent(w)
		if p.Wood != 0 {
			t.Errorf("Wood = %d, want 0 (grassland should not yield wood)", p.Wood)
		}
	})

	t.Run("safe at world edge — no panic on nil tile", func(t *testing.T) {
		w := NewWorld(3, 3)
		w.Tiles[0][0] = Tile{Terrain: Forest, TreeSize: 5}
		p := NewPlayer(0, 0) // at corner; two neighbors are out of bounds
		p.HarvestAdjacent(w) // must not panic
	})

	t.Run("harvests the forward arc (straight and both diagonals)", func(t *testing.T) {
		w := NewWorld(5, 5)
		p := NewPlayer(2, 2)
		// Default facing is north (0,-1).
		// Forward arc: N (2,1), NW (1,1), NE (3,1).
		// Non-forward: S (2,3), E (3,2), W (1,2).
		for _, coord := range [][2]int{{2, 1}, {1, 1}, {3, 1}} {
			w.Tiles[coord[1]][coord[0]] = Tile{Terrain: Forest, TreeSize: 3}
		}
		for _, coord := range [][2]int{{2, 3}, {3, 2}, {1, 2}} {
			w.Tiles[coord[1]][coord[0]] = Tile{Terrain: Forest, TreeSize: 3}
		}
		p.HarvestAdjacent(w)
		if p.Wood != 3 {
			t.Errorf("Wood = %d, want 3 (forward arc harvested)", p.Wood)
		}
		// Forward arc tiles reduced.
		for _, coord := range [][2]int{{2, 1}, {1, 1}, {3, 1}} {
			if w.Tiles[coord[1]][coord[0]].TreeSize != 2 {
				t.Errorf("forward tile (%d,%d) TreeSize = %d, want 2", coord[0], coord[1], w.Tiles[coord[1]][coord[0]].TreeSize)
			}
		}
		// Non-forward tiles untouched.
		for _, coord := range [][2]int{{2, 3}, {3, 2}, {1, 2}} {
			if w.Tiles[coord[1]][coord[0]].TreeSize != 3 {
				t.Errorf("non-forward tile (%d,%d) should be untouched, TreeSize = %d, want 3", coord[0], coord[1], w.Tiles[coord[1]][coord[0]].TreeSize)
			}
		}
	})
}

func TestHarvestCapacity(t *testing.T) {
	t.Run("harvest stops at MaxWood", func(t *testing.T) {
		w := NewWorld(5, 5)
		w.Tiles[1][2] = Tile{Terrain: Forest, TreeSize: 10}
		p := NewPlayer(2, 2)
		p.Wood = MaxWood
		p.HarvestAdjacent(w)
		if p.Wood != MaxWood {
			t.Errorf("Wood = %d, want %d (should not exceed MaxWood)", p.Wood, MaxWood)
		}
		if w.Tiles[1][2].TreeSize != 10 {
			t.Errorf("TreeSize = %d, want 10 (should not harvest when full)", w.Tiles[1][2].TreeSize)
		}
	})

	t.Run("partial fill at near-max", func(t *testing.T) {
		w := NewWorld(5, 5)
		w.Tiles[1][2] = Tile{Terrain: Forest, TreeSize: 10}
		p := NewPlayer(2, 2)
		p.Wood = MaxWood - 1
		p.HarvestAdjacent(w)
		if p.Wood != MaxWood {
			t.Errorf("Wood = %d, want %d (should fill to exactly MaxWood)", p.Wood, MaxWood)
		}
	})
}

func TestMoveCooldowns(t *testing.T) {
	forestCooldown, ok := MoveCooldowns[Forest]
	if !ok {
		t.Fatal("MoveCooldowns missing Forest entry")
	}
	grassCooldown, ok := MoveCooldowns[Grassland]
	if !ok {
		t.Fatal("MoveCooldowns missing Grassland entry")
	}
	stumpCooldown, ok := MoveCooldowns[Stump]
	if !ok {
		t.Fatal("MoveCooldowns missing Stump entry")
	}
	if forestCooldown <= grassCooldown {
		t.Errorf("Forest cooldown (%v) should be longer than Grassland (%v)", forestCooldown, grassCooldown)
	}
	if forestCooldown <= stumpCooldown {
		t.Errorf("Forest cooldown (%v) should be longer than Stump (%v)", forestCooldown, stumpCooldown)
	}
	if grassCooldown != stumpCooldown {
		t.Errorf("Grassland (%v) and Stump (%v) cooldowns should be equal", grassCooldown, stumpCooldown)
	}
}

func TestNewPlayer(t *testing.T) {
	p := NewPlayer(10, 20)

	if p.X != 10 {
		t.Errorf("X = %d, want 10", p.X)
	}
	if p.Y != 20 {
		t.Errorf("Y = %d, want 20", p.Y)
	}
	if p.Wood != 0 {
		t.Errorf("Wood = %d, want 0", p.Wood)
	}
	if p.FacingDX != 0 || p.FacingDY != -1 {
		t.Errorf("facing = (%d,%d), want (0,-1)", p.FacingDX, p.FacingDY)
	}
}
