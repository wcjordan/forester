package game

import (
	"testing"

	"forester/game/geom"
	"forester/game/internal/gametest"
)

func makeStateWithFoundation(t *testing.T) (*State, geom.Point) {
	t.Helper()
	w := NewWorld(20, 20)
	def := gametest.LogStorageDef{} // 4×4, BuildCost=20
	origin := geom.Point{X: 2, Y: 3}
	w.PlaceFoundation(origin.X, origin.Y, def)
	s := &State{
		World:               w,
		FoundationDeposited: make(map[point]int),
	}
	return s, origin
}

func TestAllFoundationsProgress_empty(t *testing.T) {
	w := NewWorld(10, 10)
	s := &State{
		World:               w,
		FoundationDeposited: make(map[point]int),
	}
	got := s.AllFoundationsProgress()
	if len(got) != 0 {
		t.Errorf("AllFoundationsProgress() len = %d, want 0", len(got))
	}
}

func TestAllFoundationsProgress_one(t *testing.T) {
	s, origin := makeStateWithFoundation(t)
	s.FoundationDeposited[origin] = 10 // 10/20 = 50%

	got := s.AllFoundationsProgress()
	if len(got) != 1 {
		t.Fatalf("AllFoundationsProgress() len = %d, want 1", len(got))
	}
	fi := got[0]
	if fi.Origin != origin {
		t.Errorf("Origin = %v, want %v", fi.Origin, origin)
	}
	if fi.Width != 4 || fi.Height != 4 {
		t.Errorf("Footprint = %dx%d, want 4x4", fi.Width, fi.Height)
	}
	if fi.Progress != 0.5 {
		t.Errorf("Progress = %v, want 0.5", fi.Progress)
	}
}

func TestAllFoundationsProgress_two(t *testing.T) {
	w := NewWorld(20, 20)
	def := gametest.LogStorageDef{} // 4×4, BuildCost=20
	o1 := geom.Point{X: 1, Y: 1}
	o2 := geom.Point{X: 10, Y: 10}
	w.PlaceFoundation(o1.X, o1.Y, def)
	w.PlaceFoundation(o2.X, o2.Y, def)
	s := &State{
		World:               w,
		FoundationDeposited: map[point]int{o1: 4, o2: 16},
	}

	got := s.AllFoundationsProgress()
	if len(got) != 2 {
		t.Fatalf("AllFoundationsProgress() len = %d, want 2", len(got))
	}
	byOrigin := make(map[geom.Point]float64, 2)
	for _, fi := range got {
		byOrigin[fi.Origin] = fi.Progress
	}
	if byOrigin[o1] != 0.2 {
		t.Errorf("progress at o1 = %v, want 0.2", byOrigin[o1])
	}
	if byOrigin[o2] != 0.8 {
		t.Errorf("progress at o2 = %v, want 0.8", byOrigin[o2])
	}
}

func TestFoundationProgressAt_originTile(t *testing.T) {
	s, origin := makeStateWithFoundation(t)
	s.FoundationDeposited[origin] = 5 // 5/20 = 0.25

	got, ok := s.FoundationProgressAt(origin)
	if !ok {
		t.Fatal("FoundationProgressAt(origin): ok = false, want true")
	}
	if got != 0.25 {
		t.Errorf("FoundationProgressAt(origin) = %v, want 0.25", got)
	}
}

func TestFoundationProgressAt_nonOriginTile(t *testing.T) {
	s, origin := makeStateWithFoundation(t)
	s.FoundationDeposited[origin] = 15 // 15/20 = 0.75

	// Any tile within the 4×4 footprint (not origin) should also return progress.
	inner := geom.Point{X: origin.X + 2, Y: origin.Y + 2}
	got, ok := s.FoundationProgressAt(inner)
	if !ok {
		t.Fatal("FoundationProgressAt(inner tile): ok = false, want true")
	}
	if got != 0.75 {
		t.Errorf("FoundationProgressAt(inner tile) = %v, want 0.75", got)
	}
}

func TestFoundationProgressAt_noFoundation(t *testing.T) {
	s, _ := makeStateWithFoundation(t)
	// Don't add any deposited amount — tile exists but no foundation progress entry.
	empty := geom.Point{X: 15, Y: 15}
	_, ok := s.FoundationProgressAt(empty)
	if ok {
		t.Error("FoundationProgressAt(empty tile): ok = true, want false")
	}
}
