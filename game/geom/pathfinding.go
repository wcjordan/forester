package geom

import "container/heap"

// Grid is the minimal interface that FindPath requires from a tile map.
// InBounds reports whether (x, y) is within the grid.
// IsBlocked reports whether the tile at (x, y) cannot be traversed
// (returns true for out-of-bounds coordinates as well).
// MoveCost returns the movement cost to enter (x, y); only called for
// non-blocked tiles. Must return a value >= 1.0 so the Manhattan heuristic
// remains admissible.
type Grid interface {
	InBounds(x, y int) bool
	IsBlocked(x, y int) bool
	MoveCost(x, y int) float64
}

// FindPath returns a path from (fromX,fromY) to (toX,toY), exclusive of
// start, inclusive of goal, using A* with Manhattan heuristic and terrain
// costs from g. Returns nil if the goal is unreachable, or []Point{} if
// start == goal.
func FindPath(g Grid, fromX, fromY, toX, toY int) []Point {
	if fromX == toX && fromY == toY {
		return []Point{}
	}

	// Fast-fail: out-of-bounds or blocked endpoints can never be part of a valid path.
	if !g.InBounds(fromX, fromY) || g.IsBlocked(toX, toY) {
		return nil
	}

	type key = Point

	gCost := make(map[key]float64)
	cameFrom := make(map[key]key)

	start := Point{X: fromX, Y: fromY}
	goal := Point{X: toX, Y: toY}

	gCost[start] = 0

	pq := &priorityQueue{}
	heap.Push(pq, &pqNode{pt: start, f: float64(manhattan(start, goal)), g: 0})

	dirs := [4]Point{{X: 0, Y: -1}, {X: 0, Y: 1}, {X: -1, Y: 0}, {X: 1, Y: 0}}

	for pq.Len() > 0 {
		cur := heap.Pop(pq).(*pqNode)

		// Stale-node guard: skip if a cheaper path to this node was already found.
		if best, ok := gCost[cur.pt]; ok && cur.g > best {
			continue
		}

		if cur.pt == goal {
			return reconstructPath(cameFrom, start, goal)
		}

		for _, d := range dirs {
			nb := Point{X: cur.pt.X + d.X, Y: cur.pt.Y + d.Y}
			if g.IsBlocked(nb.X, nb.Y) {
				continue
			}
			tentativeG := cur.g + g.MoveCost(nb.X, nb.Y)
			if best, ok := gCost[nb]; !ok || tentativeG < best {
				gCost[nb] = tentativeG
				cameFrom[nb] = cur.pt
				f := tentativeG + float64(manhattan(nb, goal))
				heap.Push(pq, &pqNode{pt: nb, f: f, g: tentativeG})
			}
		}
	}

	return nil // unreachable
}

// reconstructPath walks cameFrom backwards from goal to start and returns the
// path exclusive of start, inclusive of goal.
func reconstructPath(cameFrom map[Point]Point, start, goal Point) []Point {
	var path []Point
	cur := goal
	for cur != start {
		path = append(path, cur)
		cur = cameFrom[cur]
	}
	// Reverse so path goes from start→goal direction.
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}

// --- Priority queue (min-heap on f) ---

type pqNode struct {
	pt    Point
	f, g  float64
	index int
}

type priorityQueue []*pqNode

func (pq priorityQueue) Len() int           { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool { return pq[i].f < pq[j].f }
func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

// Push implements heap.Interface.
func (pq *priorityQueue) Push(x any) {
	n := x.(*pqNode)
	n.index = len(*pq)
	*pq = append(*pq, n)
}

// Pop implements heap.Interface.
func (pq *priorityQueue) Pop() any {
	old := *pq
	n := old[len(old)-1]
	old[len(old)-1] = nil
	*pq = old[:len(old)-1]
	return n
}
