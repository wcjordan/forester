package game

import (
	"math/rand"
	"testing"
)

func TestXPMilestoneAt(t *testing.T) {
	cases := []struct {
		n    int
		want int
	}{
		{0, 50},
		{1, 125},
		{2, 225},
		{3, 350},
		{4, 500},
		{5, 675},
	}
	for _, tc := range cases {
		got := xpMilestoneAt(tc.n)
		if got != tc.want {
			t.Errorf("xpMilestoneAt(%d) = %d, want %d", tc.n, got, tc.want)
		}
	}
}

func TestAwardXPMilestoneCrossed(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	s := newState()
	env := &Env{State: s, RNG: rng}

	// Award XP up to just below first milestone (50).
	AwardXP(env, 49)
	if s.XPMilestoneIdx != 0 {
		t.Errorf("expected no milestone yet; XPMilestoneIdx=%d", s.XPMilestoneIdx)
	}
	if len(s.pendingOfferIDs) != 0 {
		t.Errorf("expected no pending offers; got %d", len(s.pendingOfferIDs))
	}

	// Award 1 more XP to cross milestone 0 (at 50).
	AwardXP(env, 1)
	if s.XPMilestoneIdx != 1 {
		t.Errorf("expected XPMilestoneIdx=1 after crossing milestone 0; got %d", s.XPMilestoneIdx)
	}
	if len(s.pendingOfferIDs) != 1 {
		t.Errorf("expected 1 pending offer after crossing milestone; got %d", len(s.pendingOfferIDs))
	}
}

func TestAwardXPMultipleMilestones(t *testing.T) {
	rng := rand.New(rand.NewSource(0))
	s := newState()
	env := &Env{State: s, RNG: rng}

	// Award enough XP to cross milestones 0 (50) and 1 (125) at once.
	AwardXP(env, 130)
	if s.XPMilestoneIdx != 2 {
		t.Errorf("expected XPMilestoneIdx=2 after crossing milestones 0 and 1; got %d", s.XPMilestoneIdx)
	}
	if len(s.pendingOfferIDs) != 2 {
		t.Errorf("expected 2 pending offers; got %d", len(s.pendingOfferIDs))
	}
}

func TestPickCardOfferLength(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	s := newState()
	env := &Env{State: s, RNG: rng}

	offer := pickCardOffer(env)
	if len(offer) != 3 {
		t.Errorf("pickCardOffer returned %d IDs, want 3", len(offer))
	}
	// All IDs should be distinct.
	seen := make(map[string]bool)
	for _, id := range offer {
		if seen[id] {
			t.Errorf("duplicate ID in offer: %q", id)
		}
		seen[id] = true
	}
}

func TestPickCardOfferNoSpawnWithoutHouse(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	s := newState()
	env := &Env{State: s, RNG: rng}

	// Run many draws; spawn_villager should never appear when no house is unoccupied.
	for range 50 {
		offer := pickCardOffer(env)
		for _, id := range offer {
			if id == "spawn_villager" {
				t.Error("spawn_villager appeared in offer with no unoccupied houses")
			}
		}
	}
}
