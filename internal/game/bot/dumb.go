package bot

import (
	"slices"

	"github.com/JHK/hearts/internal/game"
)

type Dumb struct{}

var dumbBotNames = []string{"Linus", "Ken", "Dennis", "Anita", "Bob", "Dave", "Carol", "Ted"}

func (d *Dumb) Kind() StrategyKind { return StrategyDumb }

func chooseSmartLead(hand []game.Card, legal []game.Card) game.Card {
	nonHearts := filterCards(legal, func(card game.Card) bool {
		return card.Suit != game.SuitHearts
	})

	pool := legal
	if len(nonHearts) > 0 {
		pool = nonHearts
	}

	hasQueenSpades := game.ContainsCard(hand, game.Card{Suit: game.SuitSpades, Rank: 12})
	suitCounts := make(map[game.Suit]int, 4)
	for _, card := range hand {
		suitCounts[card.Suit]++
	}

	best := pool[0]
	for _, card := range pool[1:] {
		if compareLeadCard(card, best, suitCounts, hasQueenSpades) < 0 {
			best = card
		}
	}

	return best
}

func chooseSmartFollow(trick []game.Card, legal []game.Card) game.Card {
	leadSuit := trick[0].Suit
	winningRank := 0
	penaltyInTrick := false

	for _, played := range trick {
		if game.IsPenaltyCard(played) {
			penaltyInTrick = true
		}
		if played.Suit == leadSuit && played.Rank > winningRank {
			winningRank = played.Rank
		}
	}

	under, over := splitAgainstWinner(legal, winningRank)
	if penaltyInTrick {
		if len(under) > 0 {
			return under[len(under)-1]
		}
		return legal[0]
	}

	if len(trick) == game.PlayersPerTable-1 && len(over) > 0 {
		return over[0]
	}

	if len(under) > 0 {
		return under[len(under)-1]
	}

	return legal[0]
}

func chooseSmartDiscard(legal []game.Card) game.Card {
	queenSpades := game.Card{Suit: game.SuitSpades, Rank: 12}
	if game.ContainsCard(legal, queenSpades) {
		return queenSpades
	}

	penalties := filterCards(legal, game.IsPenaltyCard)
	if len(penalties) > 0 {
		return highestRiskCard(penalties)
	}

	spades := filterCards(legal, func(card game.Card) bool {
		return card.Suit == game.SuitSpades
	})
	if len(spades) > 0 {
		return highestRiskCard(spades)
	}

	return highestRiskCard(legal)
}

// NewDumbBot creates a dumb bot for testing.
func NewDumbBot() *Dumb {
	return &Dumb{}
}

func (s *Dumb) ChoosePlay(input game.TurnInput) (game.Card, error) {
	legal := game.LegalPlays(input.Hand, input.Trick, input.HeartsBroken, input.FirstTrick)
	if len(legal) == 0 {
		return game.Card{}, ErrNoLegalPlays
	}

	if len(input.Trick) == 0 {
		return chooseSmartLead(input.Hand, legal), nil
	}

	leadSuit := input.Trick[0].Suit
	if game.HasSuit(input.Hand, leadSuit) {
		return chooseSmartFollow(input.Trick, legal), nil
	}

	return chooseSmartDiscard(legal), nil
}

func (s *Dumb) ChoosePass(input game.PassInput) ([]game.Card, error) {
	if len(input.Hand) < 3 {
		return nil, ErrNotEnoughCards
	}

	suitCounts := make(map[game.Suit]int, 4)
	for _, card := range input.Hand {
		suitCounts[card.Suit]++
	}

	candidates := append([]game.Card(nil), input.Hand...)
	slices.SortFunc(candidates, func(a, b game.Card) int {
		aRisk := passRisk(a, suitCounts)
		bRisk := passRisk(b, suitCounts)
		if aRisk != bRisk {
			return bRisk - aRisk
		}
		if a.Rank != b.Rank {
			return b.Rank - a.Rank
		}
		return smartSuitPriority(b.Suit) - smartSuitPriority(a.Suit)
	})

	return append([]game.Card(nil), candidates[:3]...), nil
}
