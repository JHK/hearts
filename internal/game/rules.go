package game

import "fmt"

func IsPenaltyCard(card Card) bool {
	if card.Suit == Hearts {
		return true
	}
	return card.Suit == Spades && card.Rank == 12
}

func HasSuit(hand []Card, suit Suit) bool {
	for _, card := range hand {
		if card.Suit == suit {
			return true
		}
	}
	return false
}

func CanLeadCard(hand []Card, card Card, heartsBroken bool, trickNumber int) error {
	if trickNumber == 0 {
		if card != (Card{Suit: Clubs, Rank: 2}) {
			return fmt.Errorf("first trick must start with 2C")
		}
		return nil
	}

	if card.Suit == Hearts && !heartsBroken && !allCardsAreHearts(hand) {
		return fmt.Errorf("cannot lead hearts before hearts are broken")
	}

	return nil
}

func CanPlayCard(hand []Card, card Card, lead Suit, trickNumber int) error {
	if HasSuit(hand, lead) && card.Suit != lead {
		return fmt.Errorf("must follow suit %s", lead)
	}

	if trickNumber == 0 && IsPenaltyCard(card) && !allCardsArePenalty(hand) {
		return fmt.Errorf("cannot play penalty cards on first trick unless required")
	}

	return nil
}

func allCardsAreHearts(hand []Card) bool {
	if len(hand) == 0 {
		return false
	}

	for _, card := range hand {
		if card.Suit != Hearts {
			return false
		}
	}
	return true
}

func allCardsArePenalty(hand []Card) bool {
	if len(hand) == 0 {
		return false
	}

	for _, card := range hand {
		if !IsPenaltyCard(card) {
			return false
		}
	}
	return true
}
