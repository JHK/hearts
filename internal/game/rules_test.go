package game

import "testing"

func TestFirstTrickMustLeadTwoOfClubs(t *testing.T) {
	hand := []Card{{Suit: Clubs, Rank: 2}, {Suit: Hearts, Rank: 14}}

	err := CanLeadCard(hand, Card{Suit: Hearts, Rank: 14}, false, 0)
	if err == nil {
		t.Fatalf("expected first trick lead validation error")
	}

	err = CanLeadCard(hand, Card{Suit: Clubs, Rank: 2}, false, 0)
	if err != nil {
		t.Fatalf("expected 2C to be valid lead, got %v", err)
	}
}

func TestCannotLeadHeartsBeforeBroken(t *testing.T) {
	hand := []Card{{Suit: Hearts, Rank: 10}, {Suit: Clubs, Rank: 4}}

	err := CanLeadCard(hand, Card{Suit: Hearts, Rank: 10}, false, 2)
	if err == nil {
		t.Fatalf("expected hearts lead to be blocked before hearts are broken")
	}
}

func TestCanLeadHeartsWithHeartsOnly(t *testing.T) {
	hand := []Card{{Suit: Hearts, Rank: 10}, {Suit: Hearts, Rank: 4}}

	err := CanLeadCard(hand, Card{Suit: Hearts, Rank: 10}, false, 2)
	if err != nil {
		t.Fatalf("expected hearts lead to be allowed when hand is only hearts, got %v", err)
	}
}

func TestMustFollowSuitWhenPossible(t *testing.T) {
	hand := []Card{{Suit: Clubs, Rank: 10}, {Suit: Hearts, Rank: 4}}

	err := CanPlayCard(hand, Card{Suit: Hearts, Rank: 4}, Clubs, 4)
	if err == nil {
		t.Fatalf("expected follow-suit error")
	}

	err = CanPlayCard(hand, Card{Suit: Clubs, Rank: 10}, Clubs, 4)
	if err != nil {
		t.Fatalf("expected follow-suit play to be valid, got %v", err)
	}
}
