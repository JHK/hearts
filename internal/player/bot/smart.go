package bot

import (
	"fmt"
	"slices"

	"github.com/JHK/hearts/internal/game"
)

type Smart struct{}

func NewSmartBot() Strategy {
	return &Smart{}
}

func (s *Smart) ChoosePlay(input TurnInput) (game.Card, error) {
	legal := game.LegalPlays(input.Hand, input.Trick, input.HeartsBroken, input.FirstTrick)
	if len(legal) == 0 {
		return game.Card{}, fmt.Errorf("no legal plays")
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

func (s *Smart) ChoosePass(input PassInput) ([]game.Card, error) {
	if len(input.Hand) < 3 {
		return nil, fmt.Errorf("not enough cards to pass")
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

func compareLeadCard(a, b game.Card, suitCounts map[game.Suit]int, hasQueenSpades bool) int {
	aDanger := leadDangerPenalty(a, suitCounts, hasQueenSpades)
	bDanger := leadDangerPenalty(b, suitCounts, hasQueenSpades)
	if aDanger != bDanger {
		return aDanger - bDanger
	}

	aCount := suitCounts[a.Suit]
	bCount := suitCounts[b.Suit]
	if aCount != bCount {
		return aCount - bCount
	}

	if a.Rank != b.Rank {
		return a.Rank - b.Rank
	}

	return smartSuitPriority(a.Suit) - smartSuitPriority(b.Suit)
}

func leadDangerPenalty(card game.Card, suitCounts map[game.Suit]int, hasQueenSpades bool) int {
	if card.Suit != game.SuitSpades || !hasQueenSpades {
		return 0
	}

	penalty := 1
	if suitCounts[game.SuitSpades] <= 2 {
		penalty = 0
	}

	if card.Rank >= 11 {
		penalty += 2
	}

	return penalty
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

func splitAgainstWinner(cards []game.Card, winningRank int) ([]game.Card, []game.Card) {
	under := make([]game.Card, 0, len(cards))
	over := make([]game.Card, 0, len(cards))

	for _, card := range cards {
		if card.Rank < winningRank {
			under = append(under, card)
			continue
		}
		if card.Rank > winningRank {
			over = append(over, card)
		}
	}

	return under, over
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

func highestRiskCard(cards []game.Card) game.Card {
	best := cards[0]
	for _, card := range cards[1:] {
		if compareDiscardCard(card, best) > 0 {
			best = card
		}
	}
	return best
}

func compareDiscardCard(a, b game.Card) int {
	aScore := discardRisk(a)
	bScore := discardRisk(b)
	if aScore != bScore {
		return aScore - bScore
	}
	if a.Rank != b.Rank {
		return a.Rank - b.Rank
	}
	return smartSuitPriority(a.Suit) - smartSuitPriority(b.Suit)
}

func discardRisk(card game.Card) int {
	score := card.Rank * 4
	if card.Suit == game.SuitHearts {
		score += 40
	}
	if card.Suit == game.SuitSpades {
		score += 20
		if card.Rank >= 13 {
			score += 20
		}
	}
	return score
}

func passRisk(card game.Card, suitCounts map[game.Suit]int) int {
	score := card.Rank * 4

	if card.Suit == game.SuitSpades {
		if card.Rank == 12 {
			score += 240
		}
		if card.Rank == 13 || card.Rank == 14 {
			score += 120
		}
		if card.Rank >= 11 {
			score += 30
		}
	}

	if card.Suit == game.SuitHearts {
		score += 20 + card.Rank*2
	}

	switch suitCounts[card.Suit] {
	case 1:
		score += 45
	case 2:
		score += 25
	}

	if card.Suit == game.SuitClubs && card.Rank <= 4 {
		score -= 25
	}
	if card.Suit == game.SuitDiamonds && card.Rank <= 4 {
		score -= 10
	}

	return score
}

func smartSuitPriority(suit game.Suit) int {
	switch suit {
	case game.SuitClubs:
		return 0
	case game.SuitDiamonds:
		return 1
	case game.SuitSpades:
		return 2
	case game.SuitHearts:
		return 3
	default:
		return -1
	}
}

func filterCards(cards []game.Card, keep func(card game.Card) bool) []game.Card {
	out := make([]game.Card, 0, len(cards))
	for _, card := range cards {
		if keep(card) {
			out = append(out, card)
		}
	}
	return out
}
