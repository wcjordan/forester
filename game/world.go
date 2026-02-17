package game

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

// TileAt returns a pointer to the tile at the given coordinates.
// Returns nil if out of bounds.
func (w *World) TileAt(x, y int) *Tile {
	if !w.InBounds(x, y) {
		return nil
	}
	return &w.Tiles[y][x]
}
