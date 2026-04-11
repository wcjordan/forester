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
	if d.Player.PosX != 7 || d.Player.PosY != 9 {
		t.Errorf("Player position = (%g,%g), want (7,9)", d.Player.PosX, d.Player.PosY)
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

func TestSaveDataZoomLevel(t *testing.T) {
	g := testGame(t)
	g.ZoomLevel = 2.5
	d := g.SaveData()
	if d.ZoomLevel != 2.5 {
		t.Errorf("ZoomLevel = %v, want 2.5", d.ZoomLevel)
	}
	g2 := testGame(t)
	if err := g2.loadSaveData(d); err != nil {
		t.Fatalf("loadSaveData error: %v", err)
	}
	if g2.ZoomLevel != 2.5 {
		t.Errorf("loaded ZoomLevel = %v, want 2.5", g2.ZoomLevel)
	}
}

// registerTestSaveDef registers testSaveDef for the duration of the test.
func registerTestSaveDef(t *testing.T) {
	t.Helper()
	RegisterStructure(testSaveDef{}, StructureCallbacks{})
	t.Cleanup(func() {
		delete(structures, testSaveDef{}.FoundationType())
		delete(structures, testSaveDef{}.BuiltType())
	})
}

func TestLoadSaveDataRoundTrip(t *testing.T) {
	registerTestSaveDef(t)

	// Set up initial state.
	g := testGame(t)
	g.State.Player.SetTilePos(13, 17)
	g.State.Player.Inventory[Wood] = 5
	g.State.Player.MaxCarry = 30
	g.State.World.Tiles[2][3] = Tile{Terrain: Forest, TreeSize: 8, WalkCount: 2}
	origin := geom.Point{X: 6, Y: 6}
	g.State.World.PlaceBuilt(origin.X, origin.Y, testSaveDef{})
	g.State.XP = 200
	g.State.XPMilestoneIdx = 3
	g.State.completedBeats["beat1"] = true
	g.State.pendingOfferIDs = [][]string{{"x", "y"}}
	g.State.HouseOccupancy[geom.Point{X: 1, Y: 1}] = true
	g.State.FoundationDeposited[geom.Point{X: 2, Y: 2}] = 4
	g.Villagers.Spawn(8, 9)
	g.Villagers.Villagers[0].Wood = 2
	g.Villagers.Villagers[0].Task = VillagerWalkingToTree

	// Save then load into a fresh game.
	saved := g.SaveData()
	g2 := testGame(t)
	if err := g2.loadSaveData(saved); err != nil {
		t.Fatalf("loadSaveData error: %v", err)
	}

	// Player.
	p2 := g2.State.Player
	if p2.TileX() != 13 || p2.TileY() != 17 {
		t.Errorf("Player pos = (%d,%d), want (13,17)", p2.TileX(), p2.TileY())
	}
	if p2.Inventory[Wood] != 5 {
		t.Errorf("Inventory[Wood] = %d, want 5", p2.Inventory[Wood])
	}
	if p2.MaxCarry != 30 {
		t.Errorf("MaxCarry = %d, want 30", p2.MaxCarry)
	}

	// World tile.
	tile := g2.State.World.TileAt(3, 2)
	if tile.Terrain != Forest {
		t.Errorf("Terrain = %v, want Forest", tile.Terrain)
	}
	if tile.TreeSize != 8 {
		t.Errorf("TreeSize = %d, want 8", tile.TreeSize)
	}
	if tile.WalkCount != 2 {
		t.Errorf("WalkCount = %d, want 2", tile.WalkCount)
	}

	// Structure placement.
	if !g2.State.World.IsStructureOrigin(origin.X, origin.Y) {
		t.Errorf("structure not found at origin %v after load", origin)
	}
	wantStructType := testSaveDef{}.BuiltType()
	gotStructType := g2.State.World.TileAt(origin.X, origin.Y).Structure
	if gotStructType != wantStructType {
		t.Errorf("tile structure type = %q, want %q", gotStructType, wantStructType)
	}

	// XP.
	if g2.State.XP != 200 {
		t.Errorf("XP = %d, want 200", g2.State.XP)
	}
	if g2.State.XPMilestoneIdx != 3 {
		t.Errorf("XPMilestoneIdx = %d, want 3", g2.State.XPMilestoneIdx)
	}

	// Private state fields.
	if !g2.State.completedBeats["beat1"] {
		t.Error("completedBeats[\"beat1\"] = false, want true")
	}
	if len(g2.State.pendingOfferIDs) != 1 || g2.State.pendingOfferIDs[0][0] != "x" {
		t.Errorf("pendingOfferIDs = %v, want [[x y]]", g2.State.pendingOfferIDs)
	}
	if !g2.State.HouseOccupancy[geom.Point{X: 1, Y: 1}] {
		t.Error("HouseOccupancy[(1,1)] = false, want true")
	}
	if g2.State.FoundationDeposited[geom.Point{X: 2, Y: 2}] != 4 {
		t.Errorf("FoundationDeposited[(2,2)] = %d, want 4", g2.State.FoundationDeposited[geom.Point{X: 2, Y: 2}])
	}

	// Villager.
	if len(g2.Villagers.Villagers) != 1 {
		t.Fatalf("villager count = %d, want 1", len(g2.Villagers.Villagers))
	}
	v2 := g2.Villagers.Villagers[0]
	if v2.X != 8 || v2.Y != 9 {
		t.Errorf("Villager pos = (%d,%d), want (8,9)", v2.X, v2.Y)
	}
	if v2.Wood != 2 {
		t.Errorf("Villager Wood = %d, want 2", v2.Wood)
	}
	if v2.Task != VillagerWalkingToTree {
		t.Errorf("Villager Task = %v, want WalkingToTree", v2.Task)
	}
}
