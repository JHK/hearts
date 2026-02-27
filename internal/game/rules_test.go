package game

import "testing"

func TestValidatePlayFirstTrickLeadMustBeTwoClubs(t *testing.T) {
	hand := []Card{{Suit: SuitClubs, Rank: 2}, {Suit: SuitHearts, Rank: 5}}

	err := ValidatePlay(ValidatePlayInput{
		Hand:       hand,
		Card:       Card{Suit: SuitHearts, Rank: 5},
		FirstTrick: true,
	})
	if err == nil {
		t.Fatalf("expected 5H lead to be rejected on first trick")
	}

	err = ValidatePlay(ValidatePlayInput{
		Hand:       hand,
		Card:       Card{Suit: SuitClubs, Rank: 2},
		FirstTrick: true,
	})
	if err != nil {
		t.Fatalf("expected 2C lead to be accepted, got %v", err)
	}
}

func TestValidatePlayMustFollowSuit(t *testing.T) {
	hand := []Card{{Suit: SuitClubs, Rank: 10}, {Suit: SuitHearts, Rank: 4}}

	err := ValidatePlay(ValidatePlayInput{
		Hand:       hand,
		Card:       Card{Suit: SuitHearts, Rank: 4},
		Trick:      []Card{{Suit: SuitClubs, Rank: 2}},
		FirstTrick: false,
	})
	if err == nil {
		t.Fatalf("expected follow-suit validation error")
	}
}

func TestTrickWinnerAndPoints(t *testing.T) {
	winner, points, err := TrickWinner([]Play{
		{PlayerID: "p1", Card: Card{Suit: SuitSpades, Rank: 10}},
		{PlayerID: "p2", Card: Card{Suit: SuitSpades, Rank: 14}},
		{PlayerID: "p3", Card: Card{Suit: SuitHearts, Rank: 2}},
		{PlayerID: "p4", Card: Card{Suit: SuitSpades, Rank: 12}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if winner != "p2" {
		t.Fatalf("expected p2 to win, got %s", winner)
	}
	if points != 14 {
		t.Fatalf("expected 14 points, got %d", points)
	}
}

func TestApplyShootTheMoon(t *testing.T) {
	adjusted := ApplyShootTheMoon(map[PlayerID]Points{
		"p1": 26,
		"p2": 0,
		"p3": 0,
		"p4": 0,
	})

	if adjusted["p1"] != 0 || adjusted["p2"] != 26 || adjusted["p3"] != 26 || adjusted["p4"] != 26 {
		t.Fatalf("unexpected adjusted scores: %v", adjusted)
	}
}
