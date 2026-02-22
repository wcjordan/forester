package e2e_tests

import (
	"fmt"
	"os"
	"time"

	"forester/render"
)

// visualMode is enabled by setting E2E_VISUAL=1 in the environment.
// When on, each game frame is rendered to stderr with a short delay so the
// test plays back like a terminal animation.
var visualMode = os.Getenv("E2E_VISUAL") != ""

// frameDelay controls how long each frame is displayed. Override with
// E2E_VISUAL_DELAY (e.g. "50ms", "500ms"). Default: 120ms.
var frameDelay = func() time.Duration {
	if s := os.Getenv("E2E_VISUAL_DELAY"); s != "" {
		if d, err := time.ParseDuration(s); err == nil {
			return d
		}
	}
	return 120 * time.Millisecond
}()

// renderFrame clears the terminal, prints an optional label, then renders the
// current game view and pauses. No-op when E2E_VISUAL is not set.
func renderFrame(m render.Model, label string) {
	if !visualMode {
		return
	}
	// Move cursor to top-left and clear screen.
	fmt.Fprint(os.Stderr, "\x1b[H\x1b[2J")
	if label != "" {
		fmt.Fprintf(os.Stderr, "\x1b[1;33m▶ %s\x1b[0m\n", label)
	}
	fmt.Fprintln(os.Stderr, m.View())
	time.Sleep(frameDelay)
}

// announcePhase renders a phase header with a longer pause so the viewer has
// time to read it before the action begins.
func announcePhase(m render.Model, label string) {
	if !visualMode {
		return
	}
	fmt.Fprint(os.Stderr, "\x1b[H\x1b[2J")
	fmt.Fprintf(os.Stderr, "\x1b[1;36m══ %s ══\x1b[0m\n", label)
	fmt.Fprintln(os.Stderr, m.View())
	time.Sleep(frameDelay * 5)
}
