package bot

import (
	"github.com/JHK/hearts/internal/game"
)

// Hard is a stateful bot that selects a per-round strategy during the pass
// phase and maintains moon-shot state across tricks. Currently identical to
// Medium; future improvements will land here.
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

func (s *Hard) ChoosePlay(input game.TurnInput) (game.Card, error) {
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
