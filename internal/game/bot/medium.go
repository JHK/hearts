package bot

import (
	"slices"

	"github.com/JHK/hearts/internal/game"
)

type roundStrategy int

const (
	strategyDefensive    roundStrategy = iota
	strategyVoidCreation               // pass remaining cards in short suit to go void
	strategyMoonShot                   // attempt to win all penalty cards
)

// Medium is a stateful bot that selects a per-round strategy during the pass
// phase and maintains moon-shot state across tricks.
type Medium struct {
	moonShotActive   bool
	moonShotAborted  bool
	winningAllTricks bool // true while bot has won every trick this round
	prevPlayedCount  int
}

var mediumBotNames = []string{"Ada", "Grace", "Alan", "Radia", "Margaret", "Barbara", "Edsger", "Claude"}

func (s *Medium) Kind() StrategyKind { return StrategyMedium }

// MoonShotActive reports whether the bot is currently pursuing a shoot-the-moon strategy.
func (s *Medium) MoonShotActive() bool { return s.moonShotActive }

// MoonShotAborted reports whether the bot abandoned a moon-shot attempt this round.
func (s *Medium) MoonShotAborted() bool { return s.moonShotAborted }

// NewMediumBot creates a medium bot for testing.
func NewMediumBot() *Medium {
	return &Medium{}
}

func (s *Medium) ChoosePass(input game.PassInput) ([]game.Card, error) {
	if len(input.Hand) < 3 {
		return nil, ErrNotEnoughCards
	}

	strategy := selectRoundStrategy(input.Hand)
	s.moonShotActive = strategy == strategyMoonShot
	s.moonShotAborted = false
	s.winningAllTricks = true
	s.prevPlayedCount = 0

	switch strategy {
	case strategyMoonShot:
		return chooseMoonShotPass(input.Hand), nil
	default:
		return chooseDefensivePass(input.Hand), nil
	}
}

func (s *Medium) ChoosePlay(input game.TurnInput) (game.Card, error) {
	legal := game.LegalPlays(input.Hand, input.Trick, input.HeartsBroken, input.FirstTrick)
	if len(legal) == 0 {
		return game.Card{}, ErrNoLegalPlays
	}

	// Detect hold round: PlayedCards just reset to 0 but choosePass wasn't
	// called (prevPlayedCount > 0 means we were mid-round last call, not a
	// pass-phase reset). Re-evaluate hand once at round start.
	if len(input.PlayedCards) == 0 && s.prevPlayedCount > 0 {
		s.moonShotActive = evaluateMoonShot(input.Hand)
		s.moonShotAborted = false
		s.winningAllTricks = true
	}

	// Re-evaluate at trick 1 of a passing round with the actual post-pass hand.
	// choosePass saw the pre-pass hand; received cards can change viability —
	// e.g., receiving A♥ might complete a qualifying run that wasn't there before.
	if len(input.PlayedCards) == 0 && s.prevPlayedCount == 0 {
		s.moonShotActive = evaluateMoonShot(input.Hand)
		s.winningAllTricks = true
	}

	s.prevPlayedCount = len(input.PlayedCards)

	// Abort moon shot if someone else is leading this trick (they won the last).
	// FirstTrick is exempt: the player with 2♣ leads, not necessarily us.
	if !input.FirstTrick && len(input.Trick) > 0 {
		s.winningAllTricks = false
		if s.moonShotActive && !s.moonShotAborted {
			s.moonShotAborted = true
		}
	}

	pursuing := s.moonShotActive && !s.moonShotAborted

	if len(input.Trick) == 0 {
		// Count safe high cards across the entire hand (not just legal plays).
		// Hearts may not be legal leads yet (hearts not broken), but they're still
		// valid future winners — don't self-abort just because they're temporarily illegal.
		allSafeHighCards := filterCards(input.Hand, func(c game.Card) bool {
			return isSafeHighCard(c, input.Hand, input.PlayedCards)
		})
		remainingTricks := len(input.Hand)

		// Dynamic activation: if we hold guaranteed wins for ALL remaining tricks,
		// pursue moon shot regardless of prior history. This covers:
		//   - Normal mid-game detection (won all tricks so far)
		//   - End-game where penalties may have been distributed but we control
		//     the remaining tricks entirely (late-game "take control" mode).
		if len(allSafeHighCards) >= remainingTricks && remainingTricks > 0 {
			s.moonShotActive = true
			s.moonShotAborted = false
			pursuing = true
		}

		// Self-abort only when no safe high cards exist anywhere in hand.
		if pursuing && len(allSafeHighCards) == 0 {
			s.moonShotAborted = true
			pursuing = false
		}

		return smartChooseLead(input.Hand, legal, input.PlayedCards, pursuing), nil
	}

	leadSuit := input.Trick[0].Suit
	if game.HasSuit(input.Hand, leadSuit) {
		return smartChooseFollow(input.Trick, legal, input.Hand, input.PlayedCards, pursuing), nil
	}

	return smartChooseDiscard(legal, pursuing), nil
}

// --- Strategy selection ---

// selectRoundStrategy chooses a round strategy based on the initial hand.
func selectRoundStrategy(hand []game.Card) roundStrategy {
	if evaluateMoonShot(hand) {
		return strategyMoonShot
	}

	suitCounts := make(map[game.Suit]int, 4)
	for _, card := range hand {
		suitCounts[card.Suit]++
	}

	// Void creation: if 1–2 cards remain in a non-heart suit, pass them to go void.
	for _, suit := range []game.Suit{game.SuitClubs, game.SuitDiamonds, game.SuitSpades} {
		if c := suitCounts[suit]; c == 1 || c == 2 {
			return strategyVoidCreation
		}
	}

	return strategyDefensive
}

// evaluateMoonShot returns true when the hand has enough guaranteed winning
// tricks to plausibly sweep all penalties. Requires BOTH:
//   - At least 3 guaranteed consecutive heart tricks (A♥ must be held, then K♥, etc.)
//   - At least 7 guaranteed consecutive non-heart tricks (total across clubs, diamonds, spades)
//
// The balance constraint ensures early-game non-heart control (before hearts
// break) AND late-game heart control. Total minimum is 10.
func evaluateMoonShot(hand []game.Card) bool {
	return guaranteedHeartTricks(hand) >= 3 && guaranteedTricks(hand) >= 8
}

// guaranteedTricks counts how many tricks the hand can guarantee by running
// consecutive top-card sequences across all suits.
func guaranteedTricks(hand []game.Card) int {
	return guaranteedHeartTricks(hand) + guaranteedNonHeartTricks(hand)
}

// guaranteedHeartTricks counts consecutive top-card sequences in hearts only.
func guaranteedHeartTricks(hand []game.Card) int {
	count := 0
	for rank := 14; rank >= 2; rank-- {
		if game.ContainsCard(hand, game.Card{Suit: game.SuitHearts, Rank: rank}) {
			count++
		} else {
			break
		}
	}
	return count
}

// guaranteedNonHeartTricks counts consecutive top-card sequences across clubs, diamonds, and spades.
func guaranteedNonHeartTricks(hand []game.Card) int {
	total := 0
	for _, suit := range []game.Suit{game.SuitClubs, game.SuitDiamonds, game.SuitSpades} {
		for rank := 14; rank >= 2; rank-- {
			if game.ContainsCard(hand, game.Card{Suit: suit, Rank: rank}) {
				total++
			} else {
				break
			}
		}
	}
	return total
}

// --- Pass strategies ---

// chooseMoonShotPass passes cards that don't support a moon-shot attempt.
// Keeps: high hearts, Q♠, A♠/K♠; passes low off-suit cards.
func chooseMoonShotPass(hand []game.Card) []game.Card {
	type scored struct {
		card  game.Card
		score int // lower = less useful for moon shot = pass first
	}

	var candidates []scored
	for _, card := range hand {
		score := moonShotSupport(card, hand)
		candidates = append(candidates, scored{card, score})
	}

	slices.SortFunc(candidates, func(a, b scored) int {
		if a.score != b.score {
			return a.score - b.score
		}
		return a.card.Rank - b.card.Rank
	})

	out := make([]game.Card, 3)
	for i := range out {
		out[i] = candidates[i].card
	}
	return out
}

// moonShotSupport returns how valuable a card is for moon-shot pursuit.
// Cards that are part of a consecutive run from the top in their suit are
// the backbone of the moon-shot attempt and must be kept.
// Higher = keep. Lower = pass.
func moonShotSupport(card game.Card, hand []game.Card) int {
	score := card.Rank * 2

	// Check if this card belongs to the consecutive top-card run in its suit.
	inRun := true
	for rank := 14; rank > card.Rank; rank-- {
		if !game.ContainsCard(hand, game.Card{Suit: card.Suit, Rank: rank}) {
			inRun = false
			break
		}
	}
	if inRun {
		score += 50
	}

	return score
}

// chooseVoidCreationPass passes cards in the shortest non-heart suit to go void,
// filling any remaining slots with the most dangerous cards.
func chooseVoidCreationPass(hand []game.Card) []game.Card {
	suitCounts := make(map[game.Suit]int, 4)
	for _, card := range hand {
		suitCounts[card.Suit]++
	}

	// Find shortest non-heart suit with 1–2 cards.
	shortSuit := game.Suit("")
	shortCount := 3
	for _, suit := range []game.Suit{game.SuitClubs, game.SuitDiamonds, game.SuitSpades} {
		c := suitCounts[suit]
		if c > 0 && c < shortCount {
			shortSuit = suit
			shortCount = c
		}
	}

	var toPass []game.Card
	if shortSuit != "" {
		for _, card := range hand {
			if card.Suit == shortSuit {
				toPass = append(toPass, card)
			}
		}
	}

	if len(toPass) >= 3 {
		return toPass[:3]
	}

	// Fill remaining slots with most dangerous cards not already chosen.
	remaining := make([]game.Card, 0, len(hand))
	for _, card := range hand {
		if !game.ContainsCard(toPass, card) {
			remaining = append(remaining, card)
		}
	}
	slices.SortFunc(remaining, func(a, b game.Card) int {
		return passRisk(b, suitCounts) - passRisk(a, suitCounts)
	})

	for _, card := range remaining {
		if len(toPass) == 3 {
			break
		}
		toPass = append(toPass, card)
	}

	return toPass
}

// chooseDefensivePass passes the three most dangerous cards.
func chooseDefensivePass(hand []game.Card) []game.Card {
	suitCounts := make(map[game.Suit]int, 4)
	for _, card := range hand {
		suitCounts[card.Suit]++
	}

	candidates := append([]game.Card(nil), hand...)
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

	return append([]game.Card(nil), candidates[:3]...)
}

// --- Play helpers ---

// smartChooseLead picks the best card to lead a trick.
func smartChooseLead(hand []game.Card, legal []game.Card, playedCards []game.Card, pursuing bool) game.Card {
	if pursuing {
		// Lead the highest guaranteed-winner (all higher cards gone/held).
		// The self-abort in choosePlay ensures safeLeads is non-empty here.
		safeLeads := filterCards(legal, func(c game.Card) bool {
			return isSafeHighCard(c, hand, playedCards)
		})
		if len(safeLeads) > 0 {
			return highestRankedCard(safeLeads)
		}
		return highestRankedCard(legal)
	}

	// Prefer non-hearts.
	nonHearts := filterCards(legal, func(c game.Card) bool { return c.Suit != game.SuitHearts })
	pool := legal
	if len(nonHearts) > 0 {
		pool = nonHearts
	}

	// Avoid leading suits where a void has been observed (opponent will dump penalties).
	suitVoids := detectSuitVoids(playedCards, nil)
	noVoidPool := filterCards(pool, func(c game.Card) bool {
		return !suitVoids[c.Suit]
	})
	if len(noVoidPool) > 0 {
		pool = noVoidPool
	}

	// Prefer suits where opponents still hold cards (opponentCardCount >= 2),
	// reducing the chance of penalty discards from void opponents.
	crowdedSuits := filterCards(pool, func(c game.Card) bool {
		return opponentCardCount(c.Suit, hand, playedCards) >= 2
	})
	if len(crowdedSuits) > 0 {
		pool = crowdedSuits
	}

	// Prefer low cards that are NOT guaranteed trick-winners.
	// - Guaranteed winners (all higher ranks gone or held) invite void opponents to
	//   dump Q♠ onto the trick.
	// - High uncertain cards (rank ≥ 11 that can still be beaten) risk ceding control.
	// Only low non-winning cards are genuinely safe defensive leads.
	safePool := filterCards(pool, func(c game.Card) bool {
		return c.Rank < 11 && !isSafeHighCard(c, hand, playedCards)
	})
	if len(safePool) > 0 {
		pool = safePool
	}

	// Within the safe pool, further prefer genuinely low cards (rank < 9).
	// This prevents choosing a borderline card like T (rank 10) when a 2 or 3 is
	// available in another suit, while keeping the suit-length tiebreaker below.
	lowPool := filterCards(pool, func(c game.Card) bool { return c.Rank < 9 })
	// If the non-heart / void filters left no low options, widen to the full
	// legal set so a low heart beats leading a high non-heart (e.g. K♠).
	// Apply !isSafeHighCard to avoid leading a guaranteed-winner heart.
	if len(lowPool) == 0 {
		lowPool = filterCards(legal, func(c game.Card) bool {
			return c.Rank < 9 && !isSafeHighCard(c, hand, playedCards)
		})
	}
	if len(lowPool) > 0 {
		pool = lowPool
	}

	// Final tiebreaker: shortest suit, then highest rank (shed borderline cards
	// first while opponents still hold higher ones), avoiding Q♠ danger.
	hasQS := game.ContainsCard(hand, game.Card{Suit: game.SuitSpades, Rank: 12})
	suitCounts := make(map[game.Suit]int, 4)
	for _, c := range hand {
		suitCounts[c.Suit]++
	}

	best := pool[0]
	for _, c := range pool[1:] {
		if compareDefensiveLeadCard(c, best, suitCounts, hasQS) < 0 {
			best = c
		}
	}
	return best
}

// smartChooseFollow picks the best card when following suit.
func smartChooseFollow(trick []game.Card, legal []game.Card, hand []game.Card, playedCards []game.Card, pursuing bool) game.Card {
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

	if pursuing {
		// Moon shot: always try to win cheaply, save big cards.
		if len(over) > 0 {
			return lowestRankedCard(over)
		}
		// Can't win — preserve high cards for future guaranteed leads.
		if len(under) > 0 {
			return lowestRankedCard(under)
		}
		return legal[0]
	}

	// highestUnsafeUnder returns the highest under-card that is NOT a safe high card
	// (i.e. a higher card exists in opponents' hands). When all under-cards are "safe"
	// (all higher cards are in our hand or already played), shed the highest one — safe
	// high cards are still future trick-winners and therefore liabilities in defensive play.
	// Q♠ is always prioritised among unsafe cards regardless of rank: shedding it onto
	// the trick winner (who takes 13 pts) beats keeping it at the cost of a lower card.
	queenSpades := game.Card{Suit: game.SuitSpades, Rank: 12}
	highestUnsafeUnder := func(cards []game.Card) game.Card {
		unsafe := filterCards(cards, func(c game.Card) bool {
			return !isSafeHighCard(c, hand, playedCards)
		})
		if len(unsafe) > 0 {
			// Prefer Q♠ over higher-ranked non-penalty cards: it scores 13 pts
			// when taken, so offloading it here is always better than a face card.
			if game.ContainsCard(unsafe, queenSpades) {
				return queenSpades
			}
			return highestRankedCard(unsafe)
		}
		if game.ContainsCard(cards, queenSpades) {
			return queenSpades
		}
		return highestRankedCard(cards)
	}

	// Defensive: avoid taking penalty tricks.
	if penaltyInTrick {
		if len(under) > 0 {
			return highestUnsafeUnder(under)
		}
		// Forced to take; prefer non-penalty card so Q♠ isn't added to the trick.
		nonPenaltyLegal := filterCards(legal, func(c game.Card) bool { return !game.IsPenaltyCard(c) })
		if len(nonPenaltyLegal) > 0 {
			return lowestRankedCard(nonPenaltyLegal)
		}
		return lowestRankedCard(legal)
	}

	// Last to play with no penalty in trick: win cheaply.
	// Prefer non-penalty overs so Q♠ isn't played as a "cheap win" when K♠/A♠ is available.
	if len(trick) == game.PlayersPerTable-1 && len(over) > 0 {
		nonPenaltyOver := filterCards(over, func(c game.Card) bool { return !game.IsPenaltyCard(c) })
		if len(nonPenaltyOver) > 0 {
			return lowestRankedCard(nonPenaltyOver)
		}
		return lowestRankedCard(over)
	}

	// Not last, no penalty: shed highest unsafe card, or highest if all safe.
	if len(under) > 0 {
		return highestUnsafeUnder(under)
	}
	// All our cards beat the current winner; play lowest to avoid wasting high cards.
	// Prefer non-penalty so Q♠ isn't played when higher non-penalty cards (K♠/A♠) exist.
	// If we hold both Q♠ and A♠, playing Q♠ guarantees we win it (no one can override),
	// taking 13 pts unnecessarily. K♠ or A♠ cost nothing and keep Q♠ for a clean shed.
	nonPenaltyOver := filterCards(legal, func(c game.Card) bool { return !game.IsPenaltyCard(c) })
	if len(nonPenaltyOver) > 0 {
		return lowestRankedCard(nonPenaltyOver)
	}
	return lowestRankedCard(legal)
}

// smartChooseDiscard picks the best card to discard when void in the led suit.
func smartChooseDiscard(legal []game.Card, pursuing bool) game.Card {
	if pursuing {
		// Moon shot: discard lowest non-penalty card to preserve winning cards.
		nonPenalty := filterCards(legal, func(c game.Card) bool { return !game.IsPenaltyCard(c) })
		if len(nonPenalty) > 0 {
			return lowestRankedCard(nonPenalty)
		}
		return lowestRankedCard(legal)
	}

	// Defensive: shed the most dangerous card.
	queenSpades := game.Card{Suit: game.SuitSpades, Rank: 12}
	if game.ContainsCard(legal, queenSpades) {
		return queenSpades
	}

	penalties := filterCards(legal, game.IsPenaltyCard)
	if len(penalties) > 0 {
		return highestRiskCard(penalties)
	}

	spades := filterCards(legal, func(c game.Card) bool { return c.Suit == game.SuitSpades })
	if len(spades) > 0 {
		return highestRiskCard(spades)
	}

	return highestRiskCard(legal)
}

// --- Card analysis helpers ---

// isSafeHighCard returns true if all cards of higher rank in the same suit
// have already been played or are in the bot's own hand.
func isSafeHighCard(card game.Card, hand []game.Card, playedCards []game.Card) bool {
	for rank := card.Rank + 1; rank <= 14; rank++ {
		higher := game.Card{Suit: card.Suit, Rank: rank}
		if !game.ContainsCard(playedCards, higher) && !game.ContainsCard(hand, higher) {
			return false
		}
	}
	return true
}

func highestRankedCard(cards []game.Card) game.Card {
	best := cards[0]
	for _, c := range cards[1:] {
		if c.Rank > best.Rank {
			best = c
		}
	}
	return best
}

// penaltyPointsInCards returns the total penalty points across the given cards.
func penaltyPointsInCards(cards []game.Card) game.Points {
	var total game.Points
	for _, c := range cards {
		total += game.PenaltyPoints(c)
	}
	return total
}

func lowestRankedCard(cards []game.Card) game.Card {
	best := cards[0]
	for _, c := range cards[1:] {
		if c.Rank < best.Rank {
			best = c
		}
	}
	return best
}
