package game

import "container/heap"

// findPath returns a path from (fromX,fromY) to (toX,toY), exclusive of start,
// inclusive of goal, using A* with Manhattan heuristic and terrain costs.
// Returns nil if the goal is unreachable, or []Point{} if start == goal.
func findPath(w *World, fromX, fromY, toX, toY int) []Point {
	if fromX == toX && fromY == toY {
		return []Point{}
	}

	// Fast-fail: out-of-bounds or blocked endpoints can never be part of a valid path.
	if !w.InBounds(fromX, fromY) || !w.InBounds(toX, toY) {
		return nil
	}
	goalTile := w.TileAt(toX, toY)
	if goalTile == nil || goalTile.Structure != NoStructure {
		return nil
	}

	type key = Point

	gCost := make(map[key]int)
	cameFrom := make(map[key]key)

	start := Point{fromX, fromY}
	goal := Point{toX, toY}

	gCost[start] = 0

	pq := &priorityQueue{}
	heap.Push(pq, &pqNode{pt: start, f: manhattan(start, goal), g: 0})

	dirs := [4]Point{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}

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
			nb := Point{cur.pt.X + d.X, cur.pt.Y + d.Y}
			tile := w.TileAt(nb.X, nb.Y)
			if tile == nil || tile.Structure != NoStructure {
				continue
			}
			tentativeG := cur.g + tileCost(tile)
			if best, ok := gCost[nb]; !ok || tentativeG < best {
				gCost[nb] = tentativeG
				cameFrom[nb] = cur.pt
				f := tentativeG + manhattan(nb, goal)
				heap.Push(pq, &pqNode{pt: nb, f: f, g: tentativeG})
			}
		}
	}

	return nil // unreachable
}

// tileCost returns the movement cost to enter a tile.
// Forest tiles with trees cost 2; everything else costs 1.
func tileCost(t *Tile) int {
	if t.Terrain == Forest && t.TreeSize > 0 {
		return 2
	}
	return 1
}

// manhattan returns the Manhattan distance between two points.
func manhattan(a, b Point) int {
	return abs(a.X-b.X) + abs(a.Y-b.Y)
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
	f, g  int
	index int
}

type priorityQueue []*pqNode

func (pq priorityQueue) Len() int { return len(pq) }
func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].f < pq[j].f
}
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
