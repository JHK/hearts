package game

import "fmt"

type Play struct {
	PlayerID PlayerID
	Card     Card
}

type ValidatePlayInput struct {
	Hand         []Card
	Card         Card
	Trick        []Card
	HeartsBroken bool
	FirstTrick   bool
}

func ValidatePlay(input ValidatePlayInput) error {
	if !ContainsCard(input.Hand, input.Card) {
		return fmt.Errorf("card %s is not in hand", input.Card)
	}

	isLead := len(input.Trick) == 0
	if isLead {
		if input.FirstTrick {
			twoClubs := Card{Suit: SuitClubs, Rank: 2}
			if input.Card != twoClubs {
				return fmt.Errorf("first trick must be led with 2C")
			}
			return nil
		}

		if input.Card.Suit == SuitHearts && !input.HeartsBroken && !allHearts(input.Hand) {
			return fmt.Errorf("hearts are not broken")
		}

		return nil
	}

	leadSuit := input.Trick[0].Suit
	if HasSuit(input.Hand, leadSuit) && input.Card.Suit != leadSuit {
		return fmt.Errorf("must follow suit %s", leadSuit)
	}

	if input.FirstTrick && IsPenaltyCard(input.Card) && hasNonPenalty(input.Hand) {
		return fmt.Errorf("penalty cards are blocked on first trick when alternatives exist")
	}

	return nil
}

func LegalPlays(hand []Card, trick []Card, heartsBroken bool, firstTrick bool) []Card {
	legal := make([]Card, 0, len(hand))
	for _, card := range hand {
		err := ValidatePlay(ValidatePlayInput{
			Hand:         hand,
			Card:         card,
			Trick:        trick,
			HeartsBroken: heartsBroken,
			FirstTrick:   firstTrick,
		})
		if err == nil {
			legal = append(legal, card)
		}
	}

	SortCards(legal)
	return legal
}

func TrickWinner(plays []Play) (PlayerID, Points, error) {
	if len(plays) == 0 {
		return PlayerID(""), Points(0), fmt.Errorf("trick has no plays")
	}

	leadSuit := plays[0].Card.Suit
	winner := plays[0]
	points := Points(0)

	for _, play := range plays {
		points += PenaltyPoints(play.Card)
		if play.Card.Suit == leadSuit && play.Card.Rank > winner.Card.Rank {
			winner = play
		}
	}

	return winner.PlayerID, points, nil
}

func ApplyShootTheMoon(roundPoints map[PlayerID]Points) map[PlayerID]Points {
	adjusted := make(map[PlayerID]Points, len(roundPoints))
	moonShooter := PlayerID("")

	for playerID, points := range roundPoints {
		adjusted[playerID] = points
		if points == ShootTheMoonPoints {
			moonShooter = playerID
		}
	}

	if moonShooter == "" {
		return adjusted
	}

	for playerID := range adjusted {
		if playerID == moonShooter {
			adjusted[playerID] = Points(0)
			continue
		}
		adjusted[playerID] = ShootTheMoonPoints
	}

	return adjusted
}

func allHearts(hand []Card) bool {
	for _, card := range hand {
		if card.Suit != SuitHearts {
			return false
		}
	}
	return len(hand) > 0
}

func hasNonPenalty(hand []Card) bool {
	for _, card := range hand {
		if !IsPenaltyCard(card) {
			return true
		}
	}
	return false
}
