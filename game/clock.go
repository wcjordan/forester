package game

import "time"

// Clock abstracts time.Now() to allow deterministic testing.
type Clock interface {
	Now() time.Time
}

// RealClock implements Clock using the system clock.
type RealClock struct{}

// Now returns the current system time.
func (RealClock) Now() time.Time { return time.Now() }

// FakeClock implements Clock with a manually-advanced time value.
// Starts at 2024-01-01T00:00:00Z so zero-value time.Time fields
// (cooldowns not yet set) are always treated as already expired.
//
// Note: FakeClock has internal mutable state and must be used as a pointer
// (*FakeClock) when passed around (for example, as a Clock) so that calls to
// Now and Advance operate on the same underlying clock value.
type FakeClock struct {
	t time.Time
}

// NewFakeClock returns a FakeClock starting at 2024-01-01T00:00:00Z.
func NewFakeClock() *FakeClock {
	return &FakeClock{t: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)}
}

// Now returns the current fake time.
func (f *FakeClock) Now() time.Time { return f.t }

// Advance moves the fake clock forward by d.
func (f *FakeClock) Advance(d time.Duration) { f.t = f.t.Add(d) }
