package render

import (
	"fmt"
	"strings"
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

// buildProgressBar renders a text progress bar, e.g. "Building: ████░░░░ 75%".
func buildProgressBar(progress float64) string {
	const width = 8
	filled := clamp(int(progress*width), 0, width)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return fmt.Sprintf("Building: %s %d%%", bar, int(progress*100))
}
