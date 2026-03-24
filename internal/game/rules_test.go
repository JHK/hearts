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
		{Seat: 0, Card: Card{Suit: SuitSpades, Rank: 10}},
		{Seat: 1, Card: Card{Suit: SuitSpades, Rank: 14}},
		{Seat: 2, Card: Card{Suit: SuitHearts, Rank: 2}},
		{Seat: 3, Card: Card{Suit: SuitSpades, Rank: 12}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if winner != 1 {
		t.Fatalf("expected seat 1 to win, got %d", winner)
	}
	if points != 14 {
		t.Fatalf("expected 14 points, got %d", points)
	}
}

func TestApplyShootTheMoon(t *testing.T) {
	adjusted := ApplyShootTheMoon([PlayersPerTable]Points{26, 0, 0, 0})

	if adjusted[0] != 0 || adjusted[1] != 26 || adjusted[2] != 26 || adjusted[3] != 26 {
		t.Fatalf("unexpected adjusted scores: %v", adjusted)
	}
}
