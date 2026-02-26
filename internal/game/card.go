package game

import (
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
)

type Suit string

const (
	Clubs    Suit = "C"
	Diamonds Suit = "D"
	Spades   Suit = "S"
	Hearts   Suit = "H"
)

var ErrInvalidCard = errors.New("invalid card")

var suitSortOrder = map[Suit]int{
	Clubs:    0,
	Diamonds: 1,
	Spades:   2,
	Hearts:   3,
}

type Card struct {
	Suit Suit `json:"suit"`
	Rank int  `json:"rank"`
}

func (c Card) String() string {
	var rank string
	switch c.Rank {
	case 11:
		rank = "J"
	case 12:
		rank = "Q"
	case 13:
		rank = "K"
	case 14:
		rank = "A"
	default:
		rank = strconv.Itoa(c.Rank)
	}
	return rank + string(c.Suit)
}

func ParseCard(raw string) (Card, error) {
	raw = strings.ToUpper(strings.TrimSpace(raw))
	if len(raw) < 2 || len(raw) > 3 {
		return Card{}, ErrInvalidCard
	}

	rankText := raw[:len(raw)-1]
	suitText := raw[len(raw)-1:]

	var rank int
	switch rankText {
	case "J":
		rank = 11
	case "Q":
		rank = 12
	case "K":
		rank = 13
	case "A":
		rank = 14
	default:
		value, err := strconv.Atoi(rankText)
		if err != nil {
			return Card{}, ErrInvalidCard
		}
		rank = value
	}

	if rank < 2 || rank > 14 {
		return Card{}, ErrInvalidCard
	}

	card := Card{Suit: Suit(suitText), Rank: rank}
	if !isValidSuit(card.Suit) {
		return Card{}, fmt.Errorf("%w: unknown suit", ErrInvalidCard)
	}

	return card, nil
}

func NewDeck() []Card {
	deck := make([]Card, 0, 52)
	for _, suit := range []Suit{Clubs, Diamonds, Spades, Hearts} {
		for rank := 2; rank <= 14; rank++ {
			deck = append(deck, Card{Suit: suit, Rank: rank})
		}
	}
	return deck
}

func NewShuffledDeck(rng *rand.Rand) []Card {
	deck := NewDeck()
	rng.Shuffle(len(deck), func(i, j int) {
		deck[i], deck[j] = deck[j], deck[i]
	})
	return deck
}

func SortCards(cards []Card) {
	sort.Slice(cards, func(i, j int) bool {
		if cards[i].Suit != cards[j].Suit {
			return suitSortOrder[cards[i].Suit] < suitSortOrder[cards[j].Suit]
		}
		return cards[i].Rank < cards[j].Rank
	})
}

func CardsToStrings(cards []Card) []string {
	out := make([]string, 0, len(cards))
	for _, card := range cards {
		out = append(out, card.String())
	}
	return out
}

func ContainsCard(cards []Card, card Card) bool {
	for _, candidate := range cards {
		if candidate == card {
			return true
		}
	}
	return false
}

func RemoveCard(cards []Card, card Card) ([]Card, bool) {
	for i, candidate := range cards {
		if candidate == card {
			return append(cards[:i], cards[i+1:]...), true
		}
	}
	return cards, false
}

func isValidSuit(suit Suit) bool {
	_, ok := suitSortOrder[suit]
	return ok
}
