package game

import (
	"math/rand"
	"testing"

	"forester/game/geom"
)

// testSaveDef is a minimal StructureDef for save/load tests.
type testSaveDef struct{}

func (testSaveDef) FoundationType() StructureType { return "test_save_foundation" }
func (testSaveDef) BuiltType() StructureType      { return "test_save_built" }
func (testSaveDef) Footprint() (w, h int)         { return 1, 1 }
func (testSaveDef) BuildCost() int                { return 5 }

func testGame(t *testing.T) *Game {
	t.Helper()
	g := NewWithClockAndRNG(RealClock{}, rand.New(rand.NewSource(0)))
	// Replace with a small deterministic world for fast tests.
	g.State.World = NewWorld(20, 20)
	g.State.Player = NewPlayer(7, 9)
	g.State.FoundationDeposited = make(map[geom.Point]int)
	g.State.HouseOccupancy = make(map[geom.Point]bool)
	g.State.completedBeats = make(map[string]bool)
	return g
}

func TestSaveDataPlayerPosition(t *testing.T) {
	g := testGame(t)
	d := g.SaveData()
	if d.Player.X != 7 || d.Player.Y != 9 {
		t.Errorf("Player position = (%d,%d), want (7,9)", d.Player.X, d.Player.Y)
	}
}

func TestSaveDataPlayerInventory(t *testing.T) {
	g := testGame(t)
	g.State.Player.Inventory[Wood] = 12
	d := g.SaveData()
	if d.Player.Inventory[Wood] != 12 {
		t.Errorf("Player inventory Wood = %d, want 12", d.Player.Inventory[Wood])
	}
}

func TestSaveDataTileTerrain(t *testing.T) {
	g := testGame(t)
	g.State.World.Tiles[4][6] = Tile{Terrain: Forest, TreeSize: 7, WalkCount: 3}
	d := g.SaveData()
	tile := d.World.Tiles[4][6]
	if tile.Terrain != Forest {
		t.Errorf("Terrain = %v, want Forest", tile.Terrain)
	}
	if tile.TreeSize != 7 {
		t.Errorf("TreeSize = %d, want 7", tile.TreeSize)
	}
	if tile.WalkCount != 3 {
		t.Errorf("WalkCount = %d, want 3", tile.WalkCount)
	}
}

func TestSaveDataStructurePlacement(t *testing.T) {
	g := testGame(t)
	origin := geom.Point{X: 5, Y: 5}
	g.State.World.PlaceBuilt(origin.X, origin.Y, testSaveDef{})

	d := g.SaveData()
	if len(d.World.Structures) != 1 {
		t.Fatalf("len(Structures) = %d, want 1", len(d.World.Structures))
	}
	got := d.World.Structures[0]
	if got.Origin != origin {
		t.Errorf("Origin = %v, want %v", got.Origin, origin)
	}
	wantType := testSaveDef{}.BuiltType()
	if got.Type != wantType {
		t.Errorf("Type = %q, want %q", got.Type, wantType)
	}
}

func TestSaveDataXP(t *testing.T) {
	g := testGame(t)
	g.State.XP = 150
	g.State.XPMilestoneIdx = 2
	d := g.SaveData()
	if d.XP != 150 {
		t.Errorf("XP = %d, want 150", d.XP)
	}
	if d.XPMilestoneIdx != 2 {
		t.Errorf("XPMilestoneIdx = %d, want 2", d.XPMilestoneIdx)
	}
}

func TestSaveDataVillagers(t *testing.T) {
	g := testGame(t)
	g.Villagers.Spawn(3, 4)
	v := g.Villagers.Villagers[0]
	v.Wood = 3
	v.Task = VillagerCarryingToStorage
	v.TargetX = 10
	v.TargetY = 11

	d := g.SaveData()
	if len(d.Villagers) != 1 {
		t.Fatalf("len(Villagers) = %d, want 1", len(d.Villagers))
	}
	vd := d.Villagers[0]
	if vd.X != 3 || vd.Y != 4 {
		t.Errorf("Villager pos = (%d,%d), want (3,4)", vd.X, vd.Y)
	}
	if vd.Wood != 3 {
		t.Errorf("Villager Wood = %d, want 3", vd.Wood)
	}
	if vd.Task != VillagerCarryingToStorage {
		t.Errorf("Villager Task = %v, want CarryingToStorage", vd.Task)
	}
	if vd.TargetX != 10 || vd.TargetY != 11 {
		t.Errorf("Villager target = (%d,%d), want (10,11)", vd.TargetX, vd.TargetY)
	}
}

func TestSaveDataPendingOfferIDs(t *testing.T) {
	g := testGame(t)
	g.State.pendingOfferIDs = [][]string{{"a", "b"}, {"c"}}
	d := g.SaveData()
	if len(d.PendingOfferIDs) != 2 {
		t.Fatalf("len(PendingOfferIDs) = %d, want 2", len(d.PendingOfferIDs))
	}
	if d.PendingOfferIDs[0][1] != "b" {
		t.Errorf("PendingOfferIDs[0][1] = %q, want \"b\"", d.PendingOfferIDs[0][1])
	}
}

func TestSaveDataCompletedBeats(t *testing.T) {
	g := testGame(t)
	g.State.completedBeats["intro"] = true
	d := g.SaveData()
	if !d.CompletedBeats["intro"] {
		t.Error("CompletedBeats[\"intro\"] = false, want true")
	}
}
