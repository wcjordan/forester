package render

import (
	"time"

	"forester/game"
)

// statusDuration is how long a save/load status message is shown.
const statusDuration = 2 * time.Second

// saveStatusText returns the display string for a SaveStatusCode, or "" for none.
func saveStatusText(code game.SaveStatusCode) string {
	switch code {
	case game.SaveStatusSaved:
		return "Game saved"
	case game.SaveStatusLoaded:
		return "Game loaded"
	case game.SaveStatusReset:
		return "New game started"
	case game.SaveStatusSaveFailed:
		return "Save failed"
	case game.SaveStatusLoadFailed:
		return "Load failed"
	default:
		return ""
	}
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// FoundationProgressRGB returns a linearly interpolated amber→gold RGB color
// for the given build progress (0.0–1.0). Used by both the TUI and Ebiten
// renderers so both use the same color progression.
//
// 0%  → dark amber  (80,  60,  0)
// 100% → bright gold (255, 215, 0)
func FoundationProgressRGB(progress float64) (r, g, b uint8) {
	p := clampF(progress, 0, 1)
	r = uint8(80 + p*175)
	g = uint8(60 + p*155)
	b = 0
	return
}
