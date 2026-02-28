package game

import (
	"sort"
	"testing"
)

// collectRing returns all (x,y) pairs visited by chebyshevRingDo sorted for
// deterministic comparison.
func collectRing(cx, cy, r int) [][2]int {
	var pts [][2]int
	chebyshevRingDo(cx, cy, r, func(x, y int) {
		pts = append(pts, [2]int{x, y})
	})
	return pts
}

// collectBorder returns all (x,y) pairs visited by FootprintBorderDo sorted
// for deterministic comparison.
func collectBorder(ox, oy, w, h int) [][2]int {
	var pts [][2]int
	FootprintBorderDo(ox, oy, w, h, func(x, y int) {
		pts = append(pts, [2]int{x, y})
	})
	return pts
}

func sortPts(pts [][2]int) [][2]int {
	sort.Slice(pts, func(i, j int) bool {
		if pts[i][0] != pts[j][0] {
			return pts[i][0] < pts[j][0]
		}
		return pts[i][1] < pts[j][1]
	})
	return pts
}

// TestChebyshevRingDo_Ring0 verifies ring 0 is just the center point.
func TestChebyshevRingDo_Ring0(t *testing.T) {
	pts := collectRing(3, 4, 0)
	if len(pts) != 1 || pts[0] != [2]int{3, 4} {
		t.Errorf("ring 0: got %v, want [{3 4}]", pts)
	}
}

// TestChebyshevRingDo_Ring1 verifies ring 1 visits exactly the 8 neighbours.
func TestChebyshevRingDo_Ring1(t *testing.T) {
	pts := sortPts(collectRing(0, 0, 1))
	want := sortPts([][2]int{
		{-1, -1}, {0, -1}, {1, -1},
		{-1, 0}, {1, 0},
		{-1, 1}, {0, 1}, {1, 1},
	})
	if len(pts) != len(want) {
		t.Fatalf("ring 1 count = %d, want %d", len(pts), len(want))
	}
	for i := range pts {
		if pts[i] != want[i] {
			t.Errorf("ring 1[%d] = %v, want %v", i, pts[i], want[i])
		}
	}
}

// TestChebyshevRingDo_NoDuplicates verifies no tile is visited twice.
func TestChebyshevRingDo_NoDuplicates(t *testing.T) {
	for r := 1; r <= 4; r++ {
		pts := collectRing(5, 5, r)
		seen := map[[2]int]int{}
		for _, p := range pts {
			seen[p]++
		}
		for p, n := range seen {
			if n > 1 {
				t.Errorf("ring %d: tile %v visited %d times", r, p, n)
			}
		}
	}
}

// TestFootprintBorderDo_1x1 verifies a 1×1 footprint border matches
// chebyshevRingDo at r=1 (same set of tiles, order may differ).
func TestFootprintBorderDo_1x1(t *testing.T) {
	cx, cy := 5, 5
	got := sortPts(collectBorder(cx, cy, 1, 1))
	want := sortPts(collectRing(cx, cy, 1))
	if len(got) != len(want) {
		t.Fatalf("1×1 border count = %d, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("1×1 border[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

// TestFootprintBorderDo_2x2 verifies a 2×2 footprint border visits exactly
// the expected 12 perimeter tiles and excludes the footprint interior.
func TestFootprintBorderDo_2x2(t *testing.T) {
	ox, oy := 0, 0
	got := sortPts(collectBorder(ox, oy, 2, 2))
	want := sortPts([][2]int{
		{-1, -1}, {0, -1}, {1, -1}, {2, -1}, // top row
		{-1, 2}, {0, 2}, {1, 2}, {2, 2}, // bottom row
		{-1, 0}, {-1, 1}, // left column
		{2, 0}, {2, 1}, // right column
	})
	if len(got) != len(want) {
		t.Fatalf("2×2 border count = %d, want %d; got %v", len(got), len(want), got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Errorf("2×2 border[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

// TestFootprintBorderDo_3x2 verifies a non-square footprint visits the right tiles.
func TestFootprintBorderDo_3x2(t *testing.T) {
	// 3 wide, 2 tall at origin (0,0). Border tiles: outer box (5×4=20) minus inner (3×2=6) = 14.
	got := collectBorder(0, 0, 3, 2)
	// No duplicates.
	seen := map[[2]int]int{}
	seen := map[[2]int]int{}
	for _, p := range got {
		seen[p]++
	}
	for p, n := range seen {
		if n > 1 {
			t.Errorf("3×2 border: tile %v visited %d times", p, n)
		}
	}
	// Footprint tiles must not appear.
	footprint := map[[2]int]bool{
		{0, 0}: true, {1, 0}: true, {2, 0}: true,
		{0, 1}: true, {1, 1}: true, {2, 1}: true,
	}
	for _, p := range got {
		if footprint[p] {
			t.Errorf("3×2 border: footprint tile %v was visited", p)
		}
	}
	// Expected count: outer box (5×4=20) minus inner (3×2=6) = 14.
	if len(got) != 14 {
		t.Errorf("3×2 border count = %d, want 14", len(got))
	}
}
