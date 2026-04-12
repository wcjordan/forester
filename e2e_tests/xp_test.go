package e2e_tests

import (
	"strings"
	"testing"

	"forester/game"
	_ "forester/game/resources"
	_ "forester/game/structures"
	_ "forester/game/upgrades"
)

// TestXPMilestoneOffer verifies that the XP system produces 3-card milestone offers
// during normal gameplay and that the XP counter appears in the status bar.
//
// Flow:
//  1. Navigate to harvest position.
//  2. Tick until the first 3-card XP milestone offer (auto-accepting any story beat
//     1-card offers encountered along the way).
//  3. Verify the game is paused during the offer (no XP gain on tick).
//  4. Accept the offer; verify XP is shown in the status bar.
//  5. Verify the offer contains 3 distinct upgrade IDs from the expected pool.
func TestXPMilestoneOffer(t *testing.T) {
	// ── Setup: load from post-log-storage checkpoint ──────────────────────────
	// Player starts at (48,45), log storage built, MaxCarry=100, story beats for
	// log storage already completed (won't interrupt the harvest loop below).
	g, clock, m := loadFixture(t, "checkpoint_log_storage")

	// ── Phase 2: Harvest until a 3-card XP milestone offer appears ────────────
	// Story beat 1-card offers (carry_capacity etc.) are auto-accepted.
	// The first XP milestone (50 XP from chopping + building) produces a 3-card offer.
	const maxTicksToMilestone = 400
	found3CardOffer := false
	var milestoneOffer []game.UpgradeDef
	for i := range maxTicksToMilestone {
		tick(&m, clock)
		if g.HasPendingOffer() {
			offer := g.CurrentOffer()
			if len(offer) == 3 {
				found3CardOffer = true
				milestoneOffer = offer
				break
			}
			// Story beat offer (1 or 2 cards) — accept card 0 and continue.
			g.SelectCard(0)
		}
		if i == maxTicksToMilestone-1 {
			t.Fatalf("3-card XP milestone offer not reached after %d ticks; XP=%d",
				maxTicksToMilestone, g.State.XP)
		}
	}
	if !found3CardOffer {
		t.Fatal("game did not produce a 3-card XP milestone offer")
	}

	// ── Phase 3: Verify game is paused during offer ────────────────────────────
	xpBefore := g.State.XP
	tick(&m, clock)
	if g.State.XP != xpBefore {
		t.Error("game should be paused (no XP gain) while offer is pending")
	}

	// ── Phase 4: Verify 3 distinct cards from expected pool ───────────────────
	validIDs := map[string]bool{
		"harvest_speed":  true,
		"deposit_speed":  true,
		"move_speed":     true,
		"build_speed":    true,
		"spawn_villager": true,
	}
	seen := make(map[string]bool)
	for _, card := range milestoneOffer {
		id := card.ID()
		if !validIDs[id] {
			t.Errorf("unexpected card ID in XP offer: %q", id)
		}
		if seen[id] {
			t.Errorf("duplicate card ID in XP offer: %q", id)
		}
		seen[id] = true
	}

	// ── Phase 5: Accept offer; verify XP in status bar ────────────────────────
	g.SelectCard(0)
	tick(&m, clock)
	bar := statusBar(m)
	if !strings.Contains(bar, "XP:") {
		t.Errorf("status bar %q does not contain XP display", bar)
	}
}
