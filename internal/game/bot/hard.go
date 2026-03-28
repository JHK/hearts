package bot

import (
	"slices"

	"github.com/JHK/hearts/internal/game"
)

// Hard is a stateful bot that selects a per-round strategy during the pass
// phase and maintains moon-shot state across tricks. For defensive play
// decisions it uses Monte Carlo hand sampling to evaluate outcomes.
type Hard struct {
	moonShotActive   bool
	moonShotAborted  bool
	winningAllTricks bool // true while bot has won every trick this round
	prevPlayedCount  int
	blockMoonTarget  int // seat index of suspected shooter, or -1
	mc               *mcEvaluator
}

var hardBotNames = []string{"Kasparov", "Carlsen", "Neumann", "Turing", "Lovelace", "Hopper", "Knuth", "Dijkstra"}

func (s *Hard) Kind() StrategyKind { return StrategyHard }

// MoonShotActive reports whether the bot is currently pursuing a shoot-the-moon strategy.
func (s *Hard) MoonShotActive() bool { return s.moonShotActive }

// MoonShotAborted reports whether the bot abandoned a moon-shot attempt this round.
func (s *Hard) MoonShotAborted() bool { return s.moonShotAborted }

// NewHardBot creates a hard bot for testing.
func NewHardBot() *Hard {
	return &Hard{blockMoonTarget: -1}
}

func (s *Hard) ChoosePass(input game.PassInput) ([]game.Card, error) {
	if len(input.Hand) < 3 {
		return nil, ErrNotEnoughCards
	}

	strategy := hardSelectRoundStrategy(input.Hand, input.GameScores, input.MySeat)
	s.moonShotActive = strategy == strategyMoonShot
	s.moonShotAborted = false
	s.winningAllTricks = true
	s.prevPlayedCount = 0
	s.blockMoonTarget = -1

	switch strategy {
	case strategyMoonShot:
		return hardChooseMoonShotPass(input.Hand), nil
	default:
		return hardChooseDefensivePass(input.Hand, input.Direction), nil
	}
}

func (s *Hard) ChoosePlay(input game.TurnInput) (game.Card, error) {
	trick := input.TrickCards()
	playedCards := input.PlayedCardsList()
	legal := game.LegalPlays(input.Hand, trick, input.HeartsBroken, input.FirstTrick)
	if len(legal) == 0 {
		return game.Card{}, ErrNoLegalPlays
	}

	// Detect hold round: PlayedCards just reset to 0 but choosePass wasn't
	// called (prevPlayedCount > 0 means we were mid-round last call, not a
	// pass-phase reset). Re-evaluate hand once at round start.
	if len(input.PlayedCards) == 0 && s.prevPlayedCount > 0 {
		s.moonShotActive = hardEvaluateMoonShot(input.Hand, input.GameScores, input.MySeat)
		s.moonShotAborted = false
		s.winningAllTricks = true
		s.blockMoonTarget = -1
	}

	// Re-evaluate at trick 1 of a passing round with the actual post-pass hand.
	// choosePass saw the pre-pass hand; received cards can change viability —
	// e.g., receiving A♥ might complete a qualifying run that wasn't there before.
	if len(input.PlayedCards) == 0 && s.prevPlayedCount == 0 {
		s.moonShotActive = hardEvaluateMoonShot(input.Hand, input.GameScores, input.MySeat)
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
			totalPenalty := penaltyPointsInCards(playedCards)
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
		return isSafeHighCard(c, input.Hand, playedCards)
	})
	remainingTricks := len(input.Hand)

	// Dynamic activation: if we hold guaranteed wins for ALL remaining tricks
	// AND we haven't lost penalty points to other players, pursue moon shot
	// regardless of prior history.
	totalPenaltyPlayed := penaltyPointsInCards(playedCards)
	ownsAllPenalties := totalPenaltyPlayed == 0 || input.RoundPoints[input.MySeat] >= totalPenaltyPlayed
	if ownsAllPenalties && len(allSafeHighCards) >= remainingTricks && remainingTricks > 0 {
		s.moonShotActive = true
		s.moonShotAborted = false
		pursuing = true
	}

	// Soft re-activation: if we're one safe card short but have a near-safe card
	// (only one higher card unaccounted for) to fill the gap, still pursue.
	// Suppress when near game-over — the risk of a failed moon-shot is too high.
	if !pursuing && ownsAllPenalties && remainingTricks > 0 &&
		len(allSafeHighCards) == remainingTricks-1 &&
		!nearGameOver(input.GameScores, input.MySeat) {
		nearSafe := countNearSafeCards(input.Hand, playedCards)
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

	// --- Moon-shot prevention: detect opponent shooting ---
	if !pursuing {
		s.blockMoonTarget = detectMoonShooter(input.RoundPoints, input.PlayedCards, input.MySeat)
		// Score-aware: skip blocking if the shooter is the sole last-place
		// player. Their moon shot hurts all opponents equally, so spending
		// our own points to block benefits competitors as much as us.
		if s.blockMoonTarget >= 0 && !shouldBlockShooter(s.blockMoonTarget, input.GameScores) {
			s.blockMoonTarget = -1
		}
	} else {
		s.blockMoonTarget = -1
	}
	blocking := s.blockMoonTarget >= 0

	// Monte Carlo evaluation for defensive play (non-moon-shot, non-blocking).
	// Normally gates to late game (hand ≤ 7, trick 6+). When near game-over
	// (score 85+), activate earlier (hand ≤ 9, trick 4+) — the cost of a
	// bad heuristic decision is much higher when close to elimination.
	if s.mc != nil && !pursuing && !blocking && len(legal) > 1 {
		mcThreshold := 7
		if nearGameOver(input.GameScores, input.MySeat) {
			mcThreshold = 9
		}
		if len(input.Hand) <= mcThreshold {
			return s.mc.evaluate(input, legal), nil
		}
	}

	if len(input.Trick) == 0 {
		if pursuing {
			return hardMoonShotLead(input.Hand, legal, playedCards), nil
		}
		if blocking {
			return hardBlockMoonLead(input.Hand, legal, playedCards), nil
		}
		return smartChooseLead(input.Hand, legal, playedCards, false), nil
	}

	leadSuit := input.Trick[0].Card.Suit
	if game.HasSuit(input.Hand, leadSuit) {
		if pursuing {
			return hardMoonShotFollow(trick, legal), nil
		}
		return smartChooseFollow(trick, legal, input.Hand, playedCards, false), nil
	}

	if blocking {
		return hardBlockMoonDiscard(input.Trick, legal, s.blockMoonTarget), nil
	}
	return hardChooseDiscard(legal, playedCards, pursuing), nil
}

// --- Score-aware standing analysis ---

// scoreDelta returns (myScore - lowestOpponentScore). Positive means trailing.
func scoreDelta(gameScores [game.PlayersPerTable]game.Points, mySeat int) game.Points {
	myScore := gameScores[mySeat]
	minOpp := game.Points(999)
	for seat, score := range gameScores {
		if seat != mySeat && score < minOpp {
			minOpp = score
		}
	}
	return myScore - minOpp
}

// nearGameOver returns true if the bot's score is within dangerZone points
// of the game-over threshold.
func nearGameOver(gameScores [game.PlayersPerTable]game.Points, mySeat int) bool {
	return gameScores[mySeat] >= game.GameOverThreshold-15
}

// --- Hard-specific moonshot evaluation ---

// hardSelectRoundStrategy uses the hard bot's relaxed moonshot threshold,
// adjusted by game standings.
func hardSelectRoundStrategy(hand []game.Card, gameScores [game.PlayersPerTable]game.Points, mySeat int) roundStrategy {
	if hardEvaluateMoonShot(hand, gameScores, mySeat) {
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

// hardEvaluateMoonShot uses a relaxed threshold compared to medium, adjusted
// by game standings. When trailing by 30+, accept riskier moon-shot hands.
// When leading or near game-over, require stronger hands.
func hardEvaluateMoonShot(hand []game.Card, gameScores [game.PlayersPerTable]game.Points, mySeat int) bool {
	hearts := guaranteedHeartTricks(hand)
	total := guaranteedTricks(hand)

	delta := scoreDelta(gameScores, mySeat)

	// Near game-over: only attempt with very strong hands.
	if nearGameOver(gameScores, mySeat) {
		return hearts >= 4 && total >= 9
	}

	// Leading (score is at least 15 below the leader): raise threshold.
	if delta <= -15 {
		return hearts >= 3 && total >= 8
	}

	// Standard medium threshold.
	if hearts >= 3 && total >= 8 {
		return true
	}

	// Trailing by 30+: accept riskier hands.
	if delta >= 30 {
		if hearts >= 1 && total >= 6 {
			return true
		}
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

// hardChooseDefensivePass adjusts pass selection based on direction.
// Passing left is most dangerous (recipient plays right after you), so we
// increase the urgency to shed dangerous cards. Passing right is safer,
// so we can afford to keep slightly riskier cards.
func hardChooseDefensivePass(hand []game.Card, dir game.PassDirection) []game.Card {
	suitCounts := make(map[game.Suit]int, 4)
	for _, card := range hand {
		suitCounts[card.Suit]++
	}

	type scored struct {
		card game.Card
		risk int
	}

	candidates := make([]scored, len(hand))
	for i, card := range hand {
		risk := passRisk(card, suitCounts)
		candidates[i] = scored{card, risk}
	}

	// Keep low hearts (≤6) — they let us follow heart leads and take
	// a penalty trick to break opponent moon shots.
	for i, sc := range candidates {
		if sc.card.Suit == game.SuitHearts && sc.card.Rank <= 6 {
			candidates[i].risk -= 30
		}
	}

	// Direction adjustments: boost/penalise risk to shift pass priorities.
	switch dir {
	case game.PassDirectionLeft:
		// Passing left — recipient plays after us and can weaponize our
		// passed cards immediately. Shed dangerous cards more aggressively,
		// and value void creation higher (short suits become exploitable).
		for i, sc := range candidates {
			if sc.card.Suit == game.SuitSpades && sc.card.Rank >= 11 {
				candidates[i].risk += 25
			}
			if sc.card.Suit == game.SuitHearts && sc.card.Rank >= 10 {
				candidates[i].risk += 15
			}
			// Boost singleton/doubleton value: voiding a suit protects us
			// from the left neighbor leading it and trapping us.
			// Exclude hearts — voiding hearts is undesirable defensively.
			if sc.card.Suit != game.SuitHearts {
				if suitCounts[sc.card.Suit] == 1 {
					candidates[i].risk += 15
				} else if suitCounts[sc.card.Suit] == 2 {
					candidates[i].risk += 8
				}
			}
		}
	case game.PassDirectionRight:
		// Passing right — safer; recipient acts before us. Keep high cards
		// we can use reactively. Reduce urgency to shed mid-range spades.
		for i, sc := range candidates {
			if sc.card.Suit == game.SuitSpades && sc.card.Rank == 11 {
				candidates[i].risk -= 15
			}
			// Singletons are less dangerous when passing right — we play
			// after the recipient so void exploitation is weaker.
			if suitCounts[sc.card.Suit] == 1 && sc.card.Rank <= 8 {
				candidates[i].risk -= 15
			}
		}
	}

	slices.SortFunc(candidates, func(a, b scored) int {
		if a.risk != b.risk {
			return b.risk - a.risk
		}
		if a.card.Rank != b.card.Rank {
			return b.card.Rank - a.card.Rank
		}
		return smartSuitPriority(b.card.Suit) - smartSuitPriority(a.card.Suit)
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

// --- Moon-shot prevention ---

// shouldBlockShooter returns true if it's worth spending points to block
// the given shooter. Returns false if the shooter has the sole highest game
// score (clearly in last place) — their moon shot hurts all opponents
// equally, so the blocker gains no relative advantage. Always blocks if a
// successful moon-shot would push any non-shooter over the game-over threshold.
func shouldBlockShooter(shooterSeat int, gameScores [game.PlayersPerTable]game.Points) bool {
	// Always block if a moon-shot would end the game for us or another non-shooter.
	for seat, score := range gameScores {
		if seat != shooterSeat && score+game.ShootTheMoonPoints >= game.GameOverThreshold {
			return true
		}
	}

	shooterScore := gameScores[shooterSeat]
	for seat, score := range gameScores {
		if seat != shooterSeat && score >= shooterScore {
			return true // someone else is as bad or worse — worth blocking
		}
	}
	return false
}

// detectMoonShooter returns the seat of an opponent who appears to be
// shooting the moon, or -1 if no shooter is detected. Triggers when 3+
// tricks are complete, 3+ penalty points have been scored, and one opponent
// holds all of them.
func detectMoonShooter(roundPoints [game.PlayersPerTable]game.Points, plays []game.Play, mySeat int) int {
	completedTricks := len(plays) / 4
	if completedTricks < 3 {
		return -1
	}
	playedCards := game.CardsFrom(plays)
	totalPenalty := penaltyPointsInCards(playedCards)

	// Detect shooter: one opponent holds all scored penalties (3+ points).
	if totalPenalty >= 3 {
		for seat := range game.PlayersPerTable {
			if seat != mySeat && roundPoints[seat] >= totalPenalty {
				return seat
			}
		}
	}

	return -1
}

// hardBlockMoonLead picks a card to lead when trying to block an opponent's
// moon shot. Leads a safe high heart (guaranteed to win the trick) to capture
// a heart and break the shoot. Falls back to normal defensive lead otherwise
// to avoid pointlessly sacrificing points.
func hardBlockMoonLead(hand, legal, playedCards []game.Card) game.Card {
	// Lead a safe high heart — guaranteed to win the trick and capture
	// at least one heart, breaking the shoot.
	safeHearts := filterCards(legal, func(c game.Card) bool {
		return c.Suit == game.SuitHearts && isSafeHighCard(c, hand, playedCards)
	})
	if len(safeHearts) > 0 {
		return highestRankedCard(safeHearts)
	}
	// No guaranteed heart lead — play defensively.
	return smartChooseLead(hand, legal, playedCards, false)
}

// hardBlockMoonDiscard picks the best discard when void in the led suit while
// blocking a moon shot. If the shooter is winning the current trick, we hold
// penalty cards to avoid feeding them. If a non-shooter is winning, we dump
// penalties normally to split the shoot.
func hardBlockMoonDiscard(trick []game.Play, legal []game.Card, shooterSeat int) game.Card {
	leadSuit := trick[0].Card.Suit
	winningRank := 0
	winnerSeat := trick[0].Seat
	shooterPlayed := false

	for _, p := range trick {
		if p.Seat == shooterSeat {
			shooterPlayed = true
		}
		if p.Card.Suit == leadSuit && p.Card.Rank > winningRank {
			winningRank = p.Card.Rank
			winnerSeat = p.Seat
		}
	}

	// Shooter has played and is winning — don't feed them penalty cards.
	if shooterPlayed && winnerSeat == shooterSeat {
		nonPenalty := filterCards(legal, func(c game.Card) bool { return !game.IsPenaltyCard(c) })
		if len(nonPenalty) > 0 {
			return highestRiskCard(nonPenalty)
		}
		// Only penalty cards left — dump lowest to minimize damage.
		return lowestRankedCard(legal)
	}

	// Non-shooter is winning, or shooter hasn't played yet — dump normally.
	return smartChooseDiscard(legal, false)
}

// hardChooseDiscard picks the best discard when void in the led suit, with
// Q♠-aware high-spade prioritization. When Q♠ is still at large, dumping
// A♠ or K♠ takes priority over dumping hearts: A♠/K♠ risk winning Q♠ (13 pts)
// when forced to follow spades later, while a heart costs only 1 pt.
func hardChooseDiscard(legal, playedCards []game.Card, pursuing bool) game.Card {
	if pursuing {
		return smartChooseDiscard(legal, true)
	}

	queenSpades := game.Card{Suit: game.SuitSpades, Rank: 12}

	// Always dump Q♠ first.
	if game.ContainsCard(legal, queenSpades) {
		return queenSpades
	}

	// When Q♠ is at large, dump A♠/K♠ before hearts — they risk winning
	// Q♠ (13 pts) when forced to follow spades later.
	if !game.ContainsCard(playedCards, queenSpades) {
		highSpades := filterCards(legal, func(c game.Card) bool {
			return c.Suit == game.SuitSpades && c.Rank >= 13
		})
		if len(highSpades) > 0 {
			return highestRankedCard(highSpades) // A♠ before K♠
		}
	}

	// Fall through to standard defensive discard.
	return smartChooseDiscard(legal, false)
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
