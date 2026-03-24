package game

import (
	"fmt"
	"math/rand"
	"slices"
	"strings"
	"time"
)

type Suit string

const (
	SuitClubs    Suit = "C"
	SuitDiamonds Suit = "D"
	SuitSpades   Suit = "S"
	SuitHearts   Suit = "H"
)

var allSuits = []Suit{SuitClubs, SuitDiamonds, SuitSpades, SuitHearts}

type Card struct {
	Suit Suit
	Rank int
}

func (c Card) String() string {
	var rank string
	switch c.Rank {
	case 14:
		rank = "A"
	case 13:
		rank = "K"
	case 12:
		rank = "Q"
	case 11:
		rank = "J"
	case 10:
		rank = "T"
	default:
		rank = fmt.Sprintf("%d", c.Rank)
	}

	return rank + string(c.Suit)
}

func ParseCard(raw string) (Card, error) {
	raw = strings.ToUpper(strings.TrimSpace(raw))
	if len(raw) != 2 {
		return Card{}, ErrInvalidCardFormat
	}

	var rank int
	switch raw[0] {
	case 'A':
		rank = 14
	case 'K':
		rank = 13
	case 'Q':
		rank = 12
	case 'J':
		rank = 11
	case 'T':
		rank = 10
	case '2', '3', '4', '5', '6', '7', '8', '9':
		rank = int(raw[0] - '0')
	default:
		return Card{}, fmt.Errorf("%w %q", ErrInvalidCardRank, raw[0])
	}

	var suit Suit
	switch raw[1] {
	case 'C':
		suit = SuitClubs
	case 'D':
		suit = SuitDiamonds
	case 'S':
		suit = SuitSpades
	case 'H':
		suit = SuitHearts
	default:
		return Card{}, fmt.Errorf("%w %q", ErrInvalidCardSuit, raw[1])
	}

	return Card{Suit: suit, Rank: rank}, nil
}

func ParseCards(raw []string) ([]Card, error) {
	out := make([]Card, 0, len(raw))
	for _, value := range raw {
		card, err := ParseCard(value)
		if err != nil {
			return nil, err
		}
		out = append(out, card)
	}
	return out, nil
}

func CardStrings(cards []Card) []string {
	out := make([]string, len(cards))
	for i, card := range cards {
		out[i] = card.String()
	}
	return out
}

func BuildDeck() []Card {
	deck := make([]Card, 0, 52)
	for _, suit := range allSuits {
		for rank := 2; rank <= 14; rank++ {
			deck = append(deck, Card{Suit: suit, Rank: rank})
		}
	}
	return deck
}

// Deal builds a shuffled deck and distributes 13 cards to each of 4 seats.
func Deal(rng *rand.Rand) [4][]Card {
	deck := BuildDeck()
	Shuffle(deck, rng)
	var hands [4][]Card
	for i, card := range deck {
		hands[i%4] = append(hands[i%4], card)
	}
	return hands
}

func Shuffle(cards []Card, rng *rand.Rand) {
	if rng == nil {
		rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	rng.Shuffle(len(cards), func(i, j int) {
		cards[i], cards[j] = cards[j], cards[i]
	})
}

func SortCards(cards []Card) {
	slices.SortFunc(cards, compareCards)
}

func ContainsCard(cards []Card, target Card) bool {
	for _, card := range cards {
		if card == target {
			return true
		}
	}
	return false
}

func RemoveCard(cards []Card, target Card) ([]Card, bool) {
	for i, card := range cards {
		if card != target {
			continue
		}

		cards = append(cards[:i], cards[i+1:]...)
		return cards, true
	}

	return cards, false
}

func HasSuit(cards []Card, suit Suit) bool {
	for _, card := range cards {
		if card.Suit == suit {
			return true
		}
	}
	return false
}

func IsPenaltyCard(card Card) bool {
	if card.Suit == SuitHearts {
		return true
	}
	return card.Suit == SuitSpades && card.Rank == 12
}

func PenaltyPoints(card Card) Points {
	if card.Suit == SuitHearts {
		return Points(1)
	}
	if card.Suit == SuitSpades && card.Rank == 12 {
		return Points(13)
	}
	return Points(0)
}

func compareCards(a, b Card) int {
	if suitOrder(a.Suit) != suitOrder(b.Suit) {
		return suitOrder(a.Suit) - suitOrder(b.Suit)
	}
	return a.Rank - b.Rank
}

func suitOrder(suit Suit) int {
	switch suit {
	case SuitClubs:
		return 0
	case SuitDiamonds:
		return 1
	case SuitSpades:
		return 2
	case SuitHearts:
		return 3
	default:
		return 4
	}
}
