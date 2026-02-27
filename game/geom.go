package game

// abs returns the absolute value of n.
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

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

// manhattan returns the Manhattan distance between two points.
func manhattan(a, b Point) int {
	return abs(a.X-b.X) + abs(a.Y-b.Y)
}
