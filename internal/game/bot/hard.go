package bot

import (
	"slices"

	"github.com/JHK/hearts/internal/game"
)

// Hard is a stateful bot that selects a per-round strategy during the pass
// phase and maintains moon-shot state across tricks.
type Hard struct {
	moonShotActive   bool
	moonShotAborted  bool
	winningAllTricks bool // true while bot has won every trick this round
	prevPlayedCount  int
}

var hardBotNames = []string{"Kasparov", "Carlsen", "Neumann", "Turing", "Lovelace", "Hopper", "Knuth", "Dijkstra"}

func (s *Hard) Kind() StrategyKind { return StrategyHard }

// MoonShotActive reports whether the bot is currently pursuing a shoot-the-moon strategy.
func (s *Hard) MoonShotActive() bool { return s.moonShotActive }

// MoonShotAborted reports whether the bot abandoned a moon-shot attempt this round.
func (s *Hard) MoonShotAborted() bool { return s.moonShotAborted }

// NewHardBot creates a hard bot for testing.
func NewHardBot() *Hard {
	return &Hard{}
}

func (s *Hard) ChoosePass(input game.PassInput) ([]game.Card, error) {
	if len(input.Hand) < 3 {
		return nil, ErrNotEnoughCards
	}

	strategy := hardSelectRoundStrategy(input.Hand)
	s.moonShotActive = strategy == strategyMoonShot
	s.moonShotAborted = false
	s.winningAllTricks = true
	s.prevPlayedCount = 0

	switch strategy {
	case strategyMoonShot:
		return hardChooseMoonShotPass(input.Hand), nil
	default:
		return chooseDefensivePass(input.Hand), nil
	}
}

func (s *Hard) ChoosePlay(input game.TurnInput) (game.Card, error) {
	legal := game.LegalPlays(input.Hand, input.Trick, input.HeartsBroken, input.FirstTrick)
	if len(legal) == 0 {
		return game.Card{}, ErrNoLegalPlays
	}

	// Detect hold round: PlayedCards just reset to 0 but choosePass wasn't
	// called (prevPlayedCount > 0 means we were mid-round last call, not a
	// pass-phase reset). Re-evaluate hand once at round start.
	if len(input.PlayedCards) == 0 && s.prevPlayedCount > 0 {
		s.moonShotActive = hardEvaluateMoonShot(input.Hand)
		s.moonShotAborted = false
		s.winningAllTricks = true
	}

	// Re-evaluate at trick 1 of a passing round with the actual post-pass hand.
	// choosePass saw the pre-pass hand; received cards can change viability —
	// e.g., receiving A♥ might complete a qualifying run that wasn't there before.
	if len(input.PlayedCards) == 0 && s.prevPlayedCount == 0 {
		s.moonShotActive = hardEvaluateMoonShot(input.Hand)
		s.winningAllTricks = true
	}

	s.prevPlayedCount = len(input.PlayedCards)

	// Abort moon shot if someone else is leading this trick (they won the last).
	// FirstTrick is exempt: the player with 2♣ leads, not necessarily us.
	if !input.FirstTrick && len(input.Trick) > 0 {
		s.winningAllTricks = false
		if s.moonShotActive && !s.moonShotAborted {
			// Only abort if we've lost penalty points to other players.
			// If we hold all penalty points scored so far, keep pursuing —
			// we may still control remaining tricks.
			totalPenalty := penaltyPointsInCards(input.PlayedCards)
			if input.RoundPoints[input.MySeat] < totalPenalty {
				s.moonShotAborted = true
			}
		}
	}

	pursuing := s.moonShotActive && !s.moonShotAborted

	// Count safe high cards across the entire hand (not just legal plays).
	// Hearts may not be legal leads yet (hearts not broken), but they're still
	// valid future winners — don't self-abort just because they're temporarily illegal.
	allSafeHighCards := filterCards(input.Hand, func(c game.Card) bool {
		return isSafeHighCard(c, input.Hand, input.PlayedCards)
	})
	remainingTricks := len(input.Hand)

	// Dynamic activation: if we hold guaranteed wins for ALL remaining tricks
	// AND we haven't lost penalty points to other players, pursue moon shot
	// regardless of prior history.
	totalPenaltyPlayed := penaltyPointsInCards(input.PlayedCards)
	ownsAllPenalties := totalPenaltyPlayed == 0 || input.RoundPoints[input.MySeat] >= totalPenaltyPlayed
	if ownsAllPenalties && len(allSafeHighCards) >= remainingTricks && remainingTricks > 0 {
		s.moonShotActive = true
		s.moonShotAborted = false
		pursuing = true
	}

	// Soft re-activation: if we're one safe card short but have a near-safe card
	// (only one higher card unaccounted for) to fill the gap, still pursue.
	if !pursuing && ownsAllPenalties && remainingTricks > 0 &&
		len(allSafeHighCards) == remainingTricks-1 {
		nearSafe := countNearSafeCards(input.Hand, input.PlayedCards)
		if nearSafe > 0 {
			s.moonShotActive = true
			s.moonShotAborted = false
			pursuing = true
		}
	}

	// Self-abort only when no safe high cards exist anywhere in hand.
	if pursuing && len(allSafeHighCards) == 0 {
		s.moonShotAborted = true
		pursuing = false
	}

	if len(input.Trick) == 0 {
		if pursuing {
			return hardMoonShotLead(input.Hand, legal, input.PlayedCards), nil
		}
		return smartChooseLead(input.Hand, legal, input.PlayedCards, false), nil
	}

	leadSuit := input.Trick[0].Suit
	if game.HasSuit(input.Hand, leadSuit) {
		if pursuing {
			return hardMoonShotFollow(input.Trick, legal), nil
		}
		return smartChooseFollow(input.Trick, legal, input.Hand, input.PlayedCards, false), nil
	}

	return smartChooseDiscard(legal, pursuing), nil
}

// --- Hard-specific moonshot evaluation ---

// hardSelectRoundStrategy uses the hard bot's relaxed moonshot threshold.
func hardSelectRoundStrategy(hand []game.Card) roundStrategy {
	if hardEvaluateMoonShot(hand) {
		return strategyMoonShot
	}

	suitCounts := make(map[game.Suit]int, 4)
	for _, card := range hand {
		suitCounts[card.Suit]++
	}
	for _, suit := range []game.Suit{game.SuitClubs, game.SuitDiamonds, game.SuitSpades} {
		if c := suitCounts[suit]; c == 1 || c == 2 {
			return strategyVoidCreation
		}
	}
	return strategyDefensive
}

// hardEvaluateMoonShot uses a relaxed threshold compared to medium.
// EV analysis shows moonshot is profitable at ~31% success rate, so we
// trigger more aggressively. The improved pass/lead/follow strategies
// compensate for the lower threshold.
func hardEvaluateMoonShot(hand []game.Card) bool {
	hearts := guaranteedHeartTricks(hand)
	total := guaranteedTricks(hand)

	// Standard medium threshold.
	if hearts >= 3 && total >= 8 {
		return true
	}
	// Relaxed: 2+ guaranteed hearts with strong overall control.
	if hearts >= 2 && total >= 7 {
		return true
	}
	return false
}

// countNearSafeCards returns the number of cards in hand where exactly one
// higher card in their suit is unaccounted for (not played and not in hand).
// These are "probably winning" cards — only one outstanding card can beat them.
func countNearSafeCards(hand, playedCards []game.Card) int {
	count := 0
	for _, c := range hand {
		if isSafeHighCard(c, hand, playedCards) {
			continue
		}
		gaps := 0
		for rank := c.Rank + 1; rank <= 14; rank++ {
			higher := game.Card{Suit: c.Suit, Rank: rank}
			if !game.ContainsCard(playedCards, higher) && !game.ContainsCard(hand, higher) {
				gaps++
			}
		}
		if gaps == 1 {
			count++
		}
	}
	return count
}

// --- Hard-specific moonshot play strategies ---

// hardChooseMoonShotPass passes cards that don't support moonshot, preferring
// to void short off-suits for better trick control during play.
func hardChooseMoonShotPass(hand []game.Card) []game.Card {
	type scored struct {
		card  game.Card
		score int
	}

	var candidates []scored
	for _, card := range hand {
		score := moonShotSupport(card, hand)
		candidates = append(candidates, scored{card, score})
	}

	// Find a voidable off-suit: all cards in it are non-run AND count ≤ 3.
	// Passing these creates a void, letting us discard freely when opponents lead it.
	for _, suit := range []game.Suit{game.SuitClubs, game.SuitDiamonds, game.SuitSpades} {
		var suitIndices []int
		allNonRun := true
		for i, sc := range candidates {
			if sc.card.Suit == suit {
				suitIndices = append(suitIndices, i)
				if isSafeHighCard(sc.card, hand, nil) {
					allNonRun = false
				}
			}
		}
		if allNonRun && len(suitIndices) > 0 && len(suitIndices) <= 3 {
			for _, idx := range suitIndices {
				candidates[idx].score -= 40
			}
		}
	}

	slices.SortFunc(candidates, func(a, b scored) int {
		if a.score != b.score {
			return a.score - b.score
		}
		return a.card.Rank - b.card.Rank
	})

	return []game.Card{candidates[0].card, candidates[1].card, candidates[2].card}
}

// hardMoonShotLead picks the best card to lead during moonshot pursuit.
// Prefers non-heart safe leads to exhaust opponents' off-suit cards first;
// opponents who become void in off-suits discard hearts into our winning tricks.
func hardMoonShotLead(hand, legal, playedCards []game.Card) game.Card {
	safeLeads := filterCards(legal, func(c game.Card) bool {
		return isSafeHighCard(c, hand, playedCards)
	})
	if len(safeLeads) == 0 {
		return highestRankedCard(legal)
	}

	nonHeartSafe := filterCards(safeLeads, func(c game.Card) bool {
		return c.Suit != game.SuitHearts
	})
	if len(nonHeartSafe) > 0 {
		// Among non-heart safe leads, prefer suit where we hold the MOST cards
		// in hand (run the longest suit first to drain opponents' holdings).
		suitCounts := make(map[game.Suit]int)
		for _, c := range hand {
			suitCounts[c.Suit]++
		}
		best := nonHeartSafe[0]
		for _, c := range nonHeartSafe[1:] {
			cb, bb := suitCounts[c.Suit], suitCounts[best.Suit]
			if cb > bb || (cb == bb && c.Rank > best.Rank) {
				best = c
			}
		}
		return best
	}

	return highestRankedCard(safeLeads)
}

// hardMoonShotFollow picks the best card when following suit during moonshot.
// When penalty cards are at stake and we're not last to play, plays highest
// to maximize the chance of winning the trick. Otherwise wins cheaply.
func hardMoonShotFollow(trick, legal []game.Card) game.Card {
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

	_, over := splitAgainstWinner(legal, winningRank)

	if len(over) > 0 {
		// When penalty cards are in the trick and players remain after us,
		// play our highest card to maximise the chance of winning this trick.
		// A lost penalty trick kills the moonshot.
		notLast := len(trick) < game.PlayersPerTable-1
		if penaltyInTrick && notLast {
			return highestRankedCard(over)
		}
		return lowestRankedCard(over)
	}

	// Can't beat the current winner — preserve high cards.
	return lowestRankedCard(legal)
}
