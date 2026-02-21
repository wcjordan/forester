package game

// LogStorageBuildTicks is the number of ticks (at 100ms each) to complete a Log Storage build (~3s).
const LogStorageBuildTicks = 30

// BuildOperation tracks an in-progress structure build.
type BuildOperation struct {
	X, Y          int
	Width, Height int
	Target        StructureType
	ProgressTicks int
	TotalTicks    int
}

// Progress returns build completion as a fraction in [0, 1].
func (b *BuildOperation) Progress() float64 {
	return float64(b.ProgressTicks) / float64(b.TotalTicks)
}

// Done returns true when the build is complete.
func (b *BuildOperation) Done() bool {
	return b.ProgressTicks >= b.TotalTicks
}
