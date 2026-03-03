package render

import (
	"fmt"
	"strings"
)

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
