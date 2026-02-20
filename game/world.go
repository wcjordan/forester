package game

import "time"

// RegrowthTickInterval controls how often the world advances tree regrowth.
const RegrowthTickInterval = 20 * time.Second

// maxTreeSize is the maximum TreeSize a Forest tile can grow to.
const maxTreeSize = 10

// World represents the game map as a 2D grid of tiles.
type World struct {
	Width  int
	Height int
	Tiles  [][]Tile
}

// NewWorld creates a world with the given dimensions, filled with grassland.
func NewWorld(width, height int) *World {
	tiles := make([][]Tile, height)
	for y := range tiles {
		tiles[y] = make([]Tile, width)
		for x := range tiles[y] {
			tiles[y][x] = Tile{Terrain: Grassland}
		}
	}

	return &World{
		Width:  width,
		Height: height,
		Tiles:  tiles,
	}
}

// InBounds returns true if the given coordinates are within the world.
func (w *World) InBounds(x, y int) bool {
	return x >= 0 && x < w.Width && y >= 0 && y < w.Height
}

// Regrow advances tree regrowth by one step across every tile.
// Stumps regrow into small Forest tiles; Forest tiles grow toward maxTreeSize.
func (w *World) Regrow() {
	for y := range w.Tiles {
		for x := range w.Tiles[y] {
			tile := &w.Tiles[y][x]
			switch tile.Terrain {
			case Stump:
				tile.Terrain = Forest
				tile.TreeSize = 1
			case Forest:
				if tile.TreeSize < maxTreeSize {
					tile.TreeSize++
				}
			}
		}
	}
}

// TileAt returns a pointer to the tile at the given coordinates.
// Returns nil if out of bounds.
func (w *World) TileAt(x, y int) *Tile {
	if !w.InBounds(x, y) {
		return nil
	}
	return &w.Tiles[y][x]
}

// SetStructure stamps a rectangle of tiles (x, y) to (x+width-1, y+height-1)
// with the given structure type. Out-of-bounds tiles are skipped.
func (w *World) SetStructure(x, y, width, height int, stype StructureType) {
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			tile := w.TileAt(x+dx, y+dy)
			if tile != nil {
				tile.Structure = stype
			}
		}
	}
}

// IsAdjacentToStructure returns true if any of the four cardinal neighbors of
// (x, y) has the given structure type.
func (w *World) IsAdjacentToStructure(x, y int, stype StructureType) bool {
	for _, d := range [4][2]int{{0, -1}, {0, 1}, {-1, 0}, {1, 0}} {
		tile := w.TileAt(x+d[0], y+d[1])
		if tile != nil && tile.Structure == stype {
			return true
		}
	}
	return false
}
