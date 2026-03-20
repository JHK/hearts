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
	p1, p2, p3, p4 := NewPlayer(), NewPlayer(), NewPlayer(), NewPlayer()
	winner, points, err := TrickWinner([]Play{
		{Player: p1, Card: Card{Suit: SuitSpades, Rank: 10}},
		{Player: p2, Card: Card{Suit: SuitSpades, Rank: 14}},
		{Player: p3, Card: Card{Suit: SuitHearts, Rank: 2}},
		{Player: p4, Card: Card{Suit: SuitSpades, Rank: 12}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if winner != p2 {
		t.Fatalf("expected p2 to win")
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
