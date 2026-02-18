package game

// MoveMsg requests the player to move by (DX, DY) tiles.
type MoveMsg struct {
	DX, DY int
}
