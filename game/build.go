package game

import "math"

// checkGhostContact starts a build operation when the player steps onto a ghost tile.
func (s *State) checkGhostContact() {
	if s.Building != nil {
		return
	}
	tile := s.World.TileAt(s.Player.X, s.Player.Y)
	if tile == nil {
		return
	}
	def := findDefForGhost(tile.Structure)
	if def == nil {
		return
	}
	gx, gy, ok := s.ghostOriginFor(def.GhostType())
	if !ok {
		return
	}
	w, h := def.Footprint()
	s.Building = &BuildOperation{
		X: gx, Y: gy,
		Width: w, Height: h,
		Target:     def.BuiltType(),
		TotalTicks: def.BuildTicks(),
	}
	s.nudgePlayerOutside(gx, gy, w, h)
}

// nudgePlayerOutside moves the player to the closest in-bounds tile outside the rectangle.
func (s *State) nudgePlayerOutside(rx, ry, rw, rh int) {
	type candidate struct {
		x, y int
		dist int
	}
	best := candidate{dist: math.MaxInt32}
	px, py := s.Player.X, s.Player.Y

	// Check one-tile border around the footprint.
	for dy := -1; dy <= rh; dy++ {
		for dx := -1; dx <= rw; dx++ {
			// Only consider tiles on the perimeter of the extended border.
			if dx >= 0 && dx < rw && dy >= 0 && dy < rh {
				continue // inside footprint
			}
			cx, cy := rx+dx, ry+dy
			if !s.World.InBounds(cx, cy) {
				continue
			}
			t := s.World.TileAt(cx, cy)
			if t == nil || t.Structure == LogStorage {
				continue
			}
			d := (cx-px)*(cx-px) + (cy-py)*(cy-py)
			if d < best.dist {
				best = candidate{cx, cy, d}
			}
		}
	}
	if best.dist < math.MaxInt32 {
		s.Player.X = best.x
		s.Player.Y = best.y
	}
}

// AdvanceBuild increments the in-progress build and completes it when done.
func (s *State) AdvanceBuild() {
	if s.Building == nil {
		return
	}
	s.Building.ProgressTicks++
	if s.Building.Done() {
		s.World.SetStructure(s.Building.X, s.Building.Y, s.Building.Width, s.Building.Height, s.Building.Target)
		if def := findDefForBuilt(s.Building.Target); def != nil {
			origin := Point{s.Building.X, s.Building.Y}
			s.World.IndexStructure(s.Building.X, s.Building.Y, s.Building.Width, s.Building.Height, def)
			def.OnBuilt(s, origin)
		}
		s.Building = nil
	}
}

// findDefForBuilt returns the StructureDef whose BuiltType matches st, or nil.
func findDefForBuilt(st StructureType) StructureDef {
	for _, def := range structures {
		if def.BuiltType() == st {
			return def
		}
	}
	return nil
}

// findDefForGhost returns the StructureDef whose GhostType matches st, or nil.
func findDefForGhost(st StructureType) StructureDef {
	if st == NoStructure {
		return nil
	}
	for _, def := range structures {
		if def.GhostType() == st {
			return def
		}
	}
	return nil
}
