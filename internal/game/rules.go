package game

import "fmt"

type PassDirection string

const (
	PassDirectionLeft   PassDirection = "left"
	PassDirectionRight  PassDirection = "right"
	PassDirectionAcross PassDirection = "across"
	PassDirectionHold   PassDirection = "hold"
)

// PassDirectionForRound returns the pass direction for a given round index.
func PassDirectionForRound(roundIndex int) PassDirection {
	switch roundIndex % 4 {
	case 0:
		return PassDirectionLeft
	case 1:
		return PassDirectionRight
	case 2:
		return PassDirectionAcross
	default:
		return PassDirectionHold
	}
}

// PassDirectionOffset returns the seat offset for a pass direction.
func PassDirectionOffset(dir PassDirection) int {
	switch dir {
	case PassDirectionLeft:
		return 1
	case PassDirectionRight:
		return PlayersPerTable - 1
	case PassDirectionAcross:
		return 2
	default:
		return 0
	}
}

// ExchangePasses computes the cards each seat receives given submitted passes and a direction.
// received[dst] contains the cards that seat dst gains. Hold returns an empty result.
func ExchangePasses(passes [PlayersPerTable][]Card, dir PassDirection) [PlayersPerTable][]Card {
	var received [PlayersPerTable][]Card
	if dir == PassDirectionHold {
		return received
	}
	offset := PassDirectionOffset(dir)
	for src, cards := range passes {
		dst := (src + offset) % PlayersPerTable
		received[dst] = append(received[dst], cards...)
	}
	return received
}

type Play struct {
	Seat int
	Card Card
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
		return fmt.Errorf("card %s is not in hand: %w", input.Card, ErrCardNotInHand)
	}

	isLead := len(input.Trick) == 0
	if isLead {
		if input.FirstTrick {
			twoClubs := Card{Suit: SuitClubs, Rank: 2}
			if input.Card != twoClubs {
				return ErrMustLeadTwoClubs
			}
			return nil
		}

		if input.Card.Suit == SuitHearts && !input.HeartsBroken && !allHearts(input.Hand) {
			return ErrHeartsNotBroken
		}

		return nil
	}

	leadSuit := input.Trick[0].Suit
	if HasSuit(input.Hand, leadSuit) && input.Card.Suit != leadSuit {
		return fmt.Errorf("%w %s", ErrMustFollowSuit, leadSuit)
	}

	if input.FirstTrick && IsPenaltyCard(input.Card) && hasNonPenalty(input.Hand) {
		return ErrPenaltyCardBlocked
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

func TrickWinner(plays []Play) (int, Points, error) {
	if len(plays) == 0 {
		return 0, Points(0), ErrEmptyTrick
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

	return winner.Seat, points, nil
}

func ApplyShootTheMoon(roundPoints [PlayersPerTable]Points) [PlayersPerTable]Points {
	adjusted := roundPoints
	moonShooter := -1

	for i, pts := range roundPoints {
		if pts == ShootTheMoonPoints {
			moonShooter = i
			break
		}
	}

	if moonShooter == -1 {
		return adjusted
	}

	for i := range adjusted {
		if i == moonShooter {
			adjusted[i] = Points(0)
		} else {
			adjusted[i] = ShootTheMoonPoints
		}
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
