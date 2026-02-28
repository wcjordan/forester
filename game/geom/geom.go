package geom

// Point is a 2D coordinate used as a map key for spatial indexes.
type Point struct{ X, Y int }

// chebyshevRingDo calls f(x, y) for every tile on the Chebyshev ring at
// distance r from (cx, cy). Ring 0 is just the center point.
func chebyshevRingDo(cx, cy, r int, f func(x, y int)) {
	if r == 0 {
		f(cx, cy)
		return
	}
	for dx := -r; dx <= r; dx++ {
		f(cx+dx, cy-r)
		f(cx+dx, cy+r)
	}
	for dy := -r + 1; dy <= r-1; dy++ {
		f(cx-r, cy+dy)
		f(cx+r, cy+dy)
	}
}

// FootprintBorderDo calls f(x, y) for every tile on the 1-tile Chebyshev border
// around a w×h footprint whose top-left tile is at (ox, oy). The footprint tiles
// themselves are never visited; each border tile is visited exactly once.
// This is the rectangular generalisation of chebyshevRingDo.
// Precondition: w >= 1 and h >= 1.
func FootprintBorderDo(ox, oy, w, h int, f func(x, y int)) {
	if w < 1 || h < 1 {
		panic("FootprintBorderDo: w and h must be >= 1")
	}
	for x := ox - 1; x <= ox+w; x++ {
		f(x, oy-1) // top row
		f(x, oy+h) // bottom row
	}
	for y := oy; y < oy+h; y++ {
		f(ox-1, y) // left column (no corners)
		f(ox+w, y) // right column (no corners)
	}
}

// SpiralSearchDo expands Chebyshev rings outward from (cx, cy) up to maxR,
// calling f(x, y) for each tile in ring order until f returns true.
// Returns the (x, y) where f first returned true and found=true,
// or (-1, -1, false) if f never returned true.
//
// The traversal order mirrors chebyshevRingDo: for each dx (−r to +r) the top
// tile (cy−r) is checked then the bottom tile (cy+r), followed by the left and
// right column tiles (dy from −r+1 to r−1).
func SpiralSearchDo(cx, cy, maxR int, f func(x, y int) bool) (x, y int, found bool) {
	for r := 0; r <= maxR; r++ {
		if r == 0 {
			if f(cx, cy) {
				return cx, cy, true
			}
			continue
		}
		for dx := -r; dx <= r; dx++ {
			if f(cx+dx, cy-r) {
				return cx + dx, cy - r, true
			}
			if f(cx+dx, cy+r) {
				return cx + dx, cy + r, true
			}
		}
		for dy := -r + 1; dy <= r-1; dy++ {
			if f(cx-r, cy+dy) {
				return cx - r, cy + dy, true
			}
			if f(cx+r, cy+dy) {
				return cx + r, cy + dy, true
			}
		}
	}
	return -1, -1, false
}

// ForFootprintCardinalNeighbors calls f for each tile that is cardinally
// (orthogonally) adjacent to the w×h footprint with top-left at (fx, fy).
// Corner tiles of the Chebyshev border are excluded (they are only diagonally
// adjacent). Each neighbor tile is visited exactly once.
func ForFootprintCardinalNeighbors(fx, fy, fw, fh int, f func(x, y int)) {
	// Top and bottom edges (no corners).
	for x := fx; x < fx+fw; x++ {
		f(x, fy-1)
		f(x, fy+fh)
	}
	// Left and right edges.
	for y := fy; y < fy+fh; y++ {
		f(fx-1, y)
		f(fx+fw, y)
	}
}
