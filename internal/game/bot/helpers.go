package bot

import (
	"math/rand"
	"time"

	"github.com/JHK/hearts/internal/game"
)

func randomFrom(pool []string) string {
	return pool[rand.New(rand.NewSource(time.Now().UnixNano())).Intn(len(pool))]
}

// filterCards returns cards for which keep returns true.
func filterCards(cards []game.Card, keep func(card game.Card) bool) []game.Card {
	out := make([]game.Card, 0, len(cards))
	for _, card := range cards {
		if keep(card) {
			out = append(out, card)
		}
	}
	return out
}

// splitAgainstWinner partitions cards into those that lose (<winningRank) and
// those that would win (>winningRank) against the current trick winner.
func splitAgainstWinner(cards []game.Card, winningRank int) (under, over []game.Card) {
	under = make([]game.Card, 0, len(cards))
	over = make([]game.Card, 0, len(cards))
	for _, card := range cards {
		switch {
		case card.Rank < winningRank:
			under = append(under, card)
		case card.Rank > winningRank:
			over = append(over, card)
		}
	}
	return under, over
}

// highestRiskCard returns the card with the highest discard-risk score.
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

// passRisk scores how dangerous it is to keep a card (higher = pass it first).
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

// smartSuitPriority is a tiebreaker ordering for suit comparisons.
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

// detectSuitVoids scans completed tricks in playedCards (groups of 4, in play order) and
// also the in-progress currentTrick to find suits where at least one opponent has discarded
// (played off-suit when following), indicating they are void in the led suit.
// Leading a void suit risks opponents dumping penalty cards as discards.
func detectSuitVoids(playedCards []game.Card, currentTrick []game.Card) map[game.Suit]bool {
	voids := make(map[game.Suit]bool)
	// Completed tricks: playedCards is in groups of 4, first card of each group is the led card.
	for i := 0; i+3 < len(playedCards); i += 4 {
		leadSuit := playedCards[i].Suit
		for j := 1; j <= 3; j++ {
			if playedCards[i+j].Suit != leadSuit {
				voids[leadSuit] = true
				break
			}
		}
	}
	// In-progress trick: any off-suit follower is void.
	if len(currentTrick) > 1 {
		leadSuit := currentTrick[0].Suit
		for _, card := range currentTrick[1:] {
			if card.Suit != leadSuit {
				voids[leadSuit] = true
				break
			}
		}
	}
	return voids
}

// opponentCardCount estimates how many cards opponents still hold in the given
// suit: 13 minus cards in the bot's own hand in that suit minus cards already
// played in that suit.
func opponentCardCount(suit game.Suit, hand []game.Card, playedCards []game.Card) int {
	inHand := 0
	for _, c := range hand {
		if c.Suit == suit {
			inHand++
		}
	}
	played := 0
	for _, c := range playedCards {
		if c.Suit == suit {
			played++
		}
	}
	return 13 - inHand - played
}

// compareDefensiveLeadCard returns negative if a is a better defensive lead than b.
// Like compareLeadCard but prefers HIGHER rank among the filtered candidates:
// shedding the highest non-winning card first prevents it from becoming a
// guaranteed winner later as more cards get played.
func compareDefensiveLeadCard(a, b game.Card, suitCounts map[game.Suit]int, hasQueenSpades bool) int {
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
		return b.Rank - a.Rank // higher rank = better defensive lead (shed first)
	}

	return smartSuitPriority(a.Suit) - smartSuitPriority(b.Suit)
}

// compareLeadCard returns negative if a is a better lead than b.
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
