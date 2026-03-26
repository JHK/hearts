package bot

import (
	"testing"

	"github.com/JHK/hearts/internal/game"
	"github.com/stretchr/testify/require"
)

// c is a shorthand for constructing a game.Card from a string like "QS".
// Panics on invalid input — test-only.
func c(s string) game.Card {
	card, err := game.ParseCard(s)
	if err != nil {
		panic("bad card: " + s)
	}
	return card
}

// --- Trade strategy selection ---

func TestHardPassSelectsMoonShotStrategy(t *testing.T) {
	// 5 top hearts + 3 top clubs + 2 top diamonds = 10 guaranteed tricks → moon shot
	hand := parseCards(t, []string{
		"AH", "KH", "QH", "JH", "TH",
		"AC", "KC", "QC",
		"AD", "KD",
		"2S", "3S", "4S",
	})

	cards, err := NewHardBot().ChoosePass(game.PassInput{Hand: hand, Direction: game.PassDirectionLeft})
	require.NoError(t, err)
	require.Len(t, cards, 3)

	// Should not pass any of the run cards (the moon-shot backbone)
	moonShotKeep := map[string]struct{}{
		"AH": {}, "KH": {}, "QH": {}, "JH": {}, "TH": {},
		"AC": {}, "KC": {}, "QC": {},
		"AD": {}, "KD": {},
	}
	for _, c := range cards {
		_, ok := moonShotKeep[c.String()]
		require.False(t, ok, "moon-shot pass should not include run card %s", c)
	}
}

func TestHardPassSelectsDefensiveStrategyForShortSuit(t *testing.T) {
	// Only 1 club but no moon-shot potential — falls through to defensive pass.
	hand := parseCards(t, []string{
		"3C",
		"5D", "7D", "9D", "JD",
		"2S", "4S", "6S", "8S",
		"2H", "4H", "6H", "8H",
	})

	cards, err := NewHardBot().ChoosePass(game.PassInput{Hand: hand, Direction: "left"})
	require.NoError(t, err)
	require.Len(t, cards, 3)

	// Passing left: singleton 3C gets a void-creation boost, making it a top
	// pass candidate. Hearts outscore diamonds due to penalty points.
	passed := map[string]struct{}{}
	for _, c := range cards {
		passed[c.String()] = struct{}{}
	}
	require.Contains(t, passed, "3C", "passing left should void the singleton club")
}

func TestHardPassSelectsDefensiveStrategy(t *testing.T) {
	// Balanced hand, no moon-shot dominance, no near-void suit → defensive
	hand := parseCards(t, []string{
		"QS", "AS", "KH", "2C", "3C", "4D", "5D", "6H", "7H", "8S", "9S", "TC", "JD",
	})

	cards, err := NewHardBot().ChoosePass(game.PassInput{Hand: hand, Direction: "left"})
	require.NoError(t, err)
	require.Len(t, cards, 3)

	// Defensive: most dangerous cards (QS, AS, KH)
	want := map[string]struct{}{"QS": {}, "AS": {}, "KH": {}}
	for _, c := range cards {
		_, ok := want[c.String()]
		require.True(t, ok, "defensive pass expected QS/AS/KH, got unexpected %s", c)
	}
}

func TestHardPassDirectionAwareness(t *testing.T) {
	// Hand with J♠ (rank 11) and moderate hearts — direction should shift priorities.
	hand := parseCards(t, []string{
		"JS", "9S", "7S",
		"KH", "TH", "5H",
		"AD", "QD", "8D",
		"AC", "6C", "4C", "2C",
	})

	t.Run("left boosts high spade risk", func(t *testing.T) {
		cards, err := NewHardBot().ChoosePass(game.PassInput{Hand: hand, Direction: game.PassDirectionLeft})
		require.NoError(t, err)
		require.Len(t, cards, 3)

		passed := map[string]struct{}{}
		for _, c := range cards {
			passed[c.String()] = struct{}{}
		}
		// When passing left, J♠ gets a +25 boost — should be passed.
		require.Contains(t, passed, "JS", "passing left should shed J♠")
		require.Contains(t, passed, "KH", "passing left should shed K♥")
	})

	t.Run("right reduces J♠ urgency", func(t *testing.T) {
		cards, err := NewHardBot().ChoosePass(game.PassInput{Hand: hand, Direction: game.PassDirectionRight})
		require.NoError(t, err)
		require.Len(t, cards, 3)

		passed := map[string]struct{}{}
		for _, c := range cards {
			passed[c.String()] = struct{}{}
		}
		// When passing right, J♠ risk is reduced — may not be passed.
		// KH and TH should still be passed (hearts penalty unchanged).
		require.Contains(t, passed, "KH", "passing right should still shed K♥")
	})

	t.Run("across uses standard scoring", func(t *testing.T) {
		cards, err := NewHardBot().ChoosePass(game.PassInput{Hand: hand, Direction: game.PassDirectionAcross})
		require.NoError(t, err)
		require.Len(t, cards, 3)

		passed := map[string]struct{}{}
		for _, c := range cards {
			passed[c.String()] = struct{}{}
		}
		require.Contains(t, passed, "KH", "passing across should shed K♥")
	})
}

// --- Card tracking / safe-high-card play ---

func TestHardPlaysAceWhenOnlyLegalOption(t *testing.T) {
	// Holds A♣ as the only non-heart; K♣ and Q♣ already played → A♣ is safe and only legal lead.
	hand := parseCards(t, []string{"AC", "7H"})
	played := parsePlays(t, []string{"KC", "QC"})

	card, err := NewHardBot().ChoosePlay(game.TurnInput{
		Hand:        hand,
		Trick:       nil,
		PlayedCards: played,
	})
	require.NoError(t, err)
	require.Equal(t, "AC", card.String())
}

func TestHardDoesNotLeadUnsafeKing(t *testing.T) {
	// Holds K♣ but A♣ not yet played → K♣ is not safe (A♣ is still out)
	// The bot should prefer leading a shorter/safer suit instead.
	hand := parseCards(t, []string{"KC", "2D", "3D", "4D"})
	played := parsePlays(t, []string{"QC"}) // A♣ still outstanding

	card, err := NewHardBot().ChoosePlay(game.TurnInput{
		Hand:        hand,
		Trick:       nil,
		PlayedCards: played,
	})
	require.NoError(t, err)
	// Should prefer leading a diamond (shorter/safer) over K♣ when A♣ is still out
	require.NotEqual(t, "KC", card.String(), "should not lead unsafe K♣ when A♣ is still outstanding")
}

func TestHardDefensiveLeadPrefersLowestRankOverShortestSuit(t *testing.T) {
	// Clubs is the shortest non-heart suit (TC, QC = 2 cards), but TC rank 10
	// is far more likely to win a trick than 2S rank 2.
	// Defensive mode must prioritise rank over suit-length heuristic.
	hand := parseCards(t, []string{
		"TC", "QC",
		"3D", "7D", "8D",
		"2S", "3S", "9S", "KS", "AS",
		"4H", "KH",
	})

	card, err := NewHardBot().ChoosePlay(game.TurnInput{
		Hand:        hand,
		Trick:       nil,
		PlayedCards: nil,
	})
	require.NoError(t, err)
	require.Less(t, card.Rank, 10, "defensive lead should prefer low-rank card over short-suit heuristic, got %s (rank %d)", card, card.Rank)
}

func TestHardDefensiveLeadPrefersLowHeartOverHighSpade(t *testing.T) {
	// K♠ is the only non-heart but rank 13 will almost certainly win the trick.
	// With hearts broken and low hearts available, the bot should lead the
	// highest non-winning heart (shed borderline cards first).
	hand := parseCards(t, []string{"KS", "2H", "3H", "5H", "6H"})

	card, err := NewHardBot().ChoosePlay(game.TurnInput{
		Hand:         hand,
		Trick:        nil,
		HeartsBroken: true,
		PlayedCards:  nil,
	})
	require.NoError(t, err)
	require.NotEqual(t, game.SuitSpades, card.Suit, "should not lead %s when low hearts are available", card)
	// 6H is the highest non-winning heart — shed it first while opponents
	// still hold 7H–AH to beat it.
	want := game.Card{Suit: game.SuitHearts, Rank: 6}
	require.Equal(t, want, card, "expected highest safe heart")
}

func TestHardDefensiveLeadAvoidsGuaranteedWinner(t *testing.T) {
	// A♣ and K♣ are both guaranteed winners: bot holds the entire top run in clubs.
	// 5♦ is NOT a guaranteed winner (higher diamonds may still be in opponents' hands).
	// Defensive mode must prefer 5♦ to avoid inviting Q♠ dumps from void opponents.
	hand := parseCards(t, []string{"AC", "KC", "5D"})

	card, err := NewHardBot().ChoosePlay(game.TurnInput{
		Hand:        hand,
		Trick:       nil,
		PlayedCards: nil,
	})
	require.NoError(t, err)
	require.NotEqual(t, "AC", card.String(), "defensive lead should avoid guaranteed-winner when 5D is available")
	require.NotEqual(t, "KC", card.String(), "defensive lead should avoid guaranteed-winner when 5D is available")
}

// --- Void exploitation ---

func TestHardDiscardsQueenOfSpadesWhenVoid(t *testing.T) {
	// Can't follow clubs → should dump Q♠
	hand := parseCards(t, []string{"QS", "AH", "2D"})

	card, err := NewHardBot().ChoosePlay(game.TurnInput{
		Hand:         hand,
		Trick:        parsePlays(t, []string{"5C"}),
		HeartsBroken: true,
	})
	require.NoError(t, err)
	require.Equal(t, "QS", card.String())
}

func TestHardDiscardsHighHeartWhenVoidAndNoQueenSpades(t *testing.T) {
	// Can't follow clubs, no Q♠ → discard highest heart
	hand := parseCards(t, []string{"AH", "KH", "2D"})

	card, err := NewHardBot().ChoosePlay(game.TurnInput{
		Hand:         hand,
		Trick:        parsePlays(t, []string{"5C"}),
		HeartsBroken: true,
	})
	require.NoError(t, err)
	require.Equal(t, "AH", card.String())
}

// --- Moon-shot triggering and pursuit ---

func TestHardEvaluatesMoonShotFromHand(t *testing.T) {
	// 5 top hearts + 3 top clubs + 2 top diamonds = 10 guaranteed tricks
	hand := parseCards(t, []string{
		"AH", "KH", "QH", "JH", "TH",
		"AC", "KC", "QC",
		"AD", "KD",
		"2S", "3S", "4S",
	})
	require.True(t, evaluateMoonShot(hand), "expected moon-shot potential")
}

func TestHardDoesNotTriggerMoonShotOnWeakHand(t *testing.T) {
	hand := parseCards(t, []string{
		"2H", "3H", "4H",
		"2C", "3C", "4C", "5C", "6C",
		"2D", "3D", "4D", "5D", "6D",
	})
	require.False(t, evaluateMoonShot(hand), "weak hand should not trigger moon shot")
}

func TestHardPursueMoonShotLeadsHighCards(t *testing.T) {
	hand := parseCards(t, []string{"AH", "KH", "QH", "AS"})
	legal := hand

	card := smartChooseLead(hand, legal, nil, true)
	require.Equal(t, 14, card.Rank, "moon shot should lead highest card (A), got %s", card)
}

func TestHardPursueMoonShotWinsFollowTrick(t *testing.T) {
	// Trick: JS led; bot has QS → should play QS to win
	trick := parseCards(t, []string{"JS"})
	legal := parseCards(t, []string{"QS"}) // must follow spades

	card := smartChooseFollow(trick, legal, nil, nil, true)
	require.Equal(t, "QS", card.String(), "moon shot follow should win trick")
}

// --- Moon-shot abort on penalty leak ---

func TestHardAbortsMoonShotWhenOtherLeads(t *testing.T) {
	bot := &Hard{moonShotActive: true, moonShotAborted: false, prevPlayedCount: 4}

	// Not first trick; trick already has a card (someone else led) → abort
	// Bot has 0 round points but 8H (1 pt) was played — penalty leaked to opponent.
	hand := parseCards(t, []string{"AH", "KH"})

	_, err := bot.ChoosePlay(game.TurnInput{
		Hand:          hand,
		Trick:         parsePlays(t, []string{"3D"}), // someone else led
		HeartsBroken:  false,
		FirstTrick:    false,
		PlayedCards: parsePlays(t, []string{"2C", "5D", "7S", "8H"}), // 1 completed trick
		RoundPoints: [game.PlayersPerTable]game.Points{0, 1, 0, 0},    // bot (seat 0) didn't win the 8H
		MySeat:      0,
	})
	require.NoError(t, err)
	require.True(t, bot.moonShotAborted, "expected moon shot to be aborted when another player leads and bot lost penalty points")
}

func TestHardKeepsMoonShotWhenFollowingWithAllPenalties(t *testing.T) {
	// Bug scenario from bean: bot has 13 round points (from QS), holds strong cards,
	// but moonshot was aborted just because someone else led.
	// With the fix, the bot should keep pursuing because it holds all penalty points.
	bot := &Hard{moonShotActive: true, moonShotAborted: false, prevPlayedCount: 32}

	// 8 tricks completed (32 played cards), QS was taken by bot (13 pts).
	// Someone else leads 6H. Bot holds AC/JD/3H/KH/AH with 5 remaining tricks.
	hand := parseCards(t, []string{"AC", "JD", "3H", "KH", "AH"})
	played := parsePlays(t, []string{
		// 8 completed tricks (32 cards); QS was in one of them
		"QS", "2S", "3S", "4S",
		"2C", "3C", "4C", "5C",
		"2D", "3D", "4D", "5D",
		"6C", "7C", "8C", "9C",
		"6D", "7D", "8D", "9D",
		"6S", "7S", "8S", "9S",
		"TC", "TD", "TS", "KC",
		"KD", "KS", "QC", "QD",
	})

	_, err := bot.ChoosePlay(game.TurnInput{
		Hand:          hand,
		Trick:         parsePlays(t, []string{"6H"}), // someone else led
		HeartsBroken:  true,
		FirstTrick:    false,
		PlayedCards:   played,
		RoundPoints: [game.PlayersPerTable]game.Points{13, 0, 0, 0}, // bot (seat 0) captured QS
		MySeat:      0,
	})
	require.NoError(t, err)
	require.False(t, bot.moonShotAborted, "moon shot should NOT abort when bot holds all penalty points and follows")
	require.True(t, bot.moonShotActive, "moon shot should remain active")
}

func TestHardContinuesMoonShotWhenLeadingEveryTrick(t *testing.T) {
	bot := &Hard{moonShotActive: true, moonShotAborted: false, prevPlayedCount: 4}

	// Bot is leading (trick is empty), not first trick → should NOT abort
	hand := parseCards(t, []string{"AH", "KH", "QH", "JH"})

	_, err := bot.ChoosePlay(game.TurnInput{
		Hand:         hand,
		Trick:        nil, // bot is leading
		HeartsBroken: true,
		FirstTrick:   false,
		PlayedCards:  parsePlays(t, []string{"2C", "5D", "7S", "8H"}),
	})
	require.NoError(t, err)
	require.False(t, bot.moonShotAborted, "moon shot should not abort when bot is leading every trick")
}

// --- Hard-specific moonshot evaluation ---

func TestHardRelaxedMoonShotThreshold(t *testing.T) {
	// A♥-K♥ + A♣-K♣-Q♣-J♣-T♣ = 2 hearts + 5 clubs = 7 guaranteed tricks.
	// Medium wouldn't trigger (needs 3 hearts or 8 total); hard should.
	hand := parseCards(t, []string{
		"AH", "KH",
		"AC", "KC", "QC", "JC", "TC",
		"2D", "3D", "4D", "5D", "6D", "7D",
	})
	require.True(t, hardEvaluateMoonShot(hand), "hard should trigger moonshot with 2 hearts + 7 total")
	require.False(t, evaluateMoonShot(hand), "medium threshold should NOT trigger for this hand")
}

// --- Hard moonshot lead prefers non-hearts ---

func TestHardMoonShotLeadPrefersNonHearts(t *testing.T) {
	// Bot has safe leads in both hearts and clubs. Should prefer clubs first.
	hand := parseCards(t, []string{"AH", "KH", "AC", "KC"})
	played := parseCards(t, []string{}) // empty, both A's are top of suit
	legal := hand                       // all can be led (hearts broken implied)

	card := hardMoonShotLead(hand, legal, played)
	require.Equal(t, game.SuitClubs, card.Suit,
		"moonshot lead should prefer non-heart suit, got %s", card)
}

// --- Hard moonshot follow plays highest when penalty at stake ---

func TestHardMoonShotFollowPlaysHighestWhenPenaltyAndNotLast(t *testing.T) {
	// Following a heart trick (penalty in trick), not last to play.
	// Should play highest to maximize chance of winning the penalty trick.
	trick := parseCards(t, []string{"5H"})            // heart led, 1 card in trick
	legal := parseCards(t, []string{"7H", "TH", "KH"}) // must follow hearts

	card := hardMoonShotFollow(trick, legal)
	require.Equal(t, "KH", card.String(),
		"should play highest when penalty in trick and not last")
}

func TestHardMoonShotFollowPlaysLowestWhenLastToPlay(t *testing.T) {
	// Following a heart trick, last to play (3 cards in trick already).
	// Should play lowest winning card (guaranteed to take trick).
	trick := parseCards(t, []string{"5H", "3H", "2H"}) // 3 cards, we're last
	legal := parseCards(t, []string{"7H", "TH", "KH"})

	card := hardMoonShotFollow(trick, legal)
	require.Equal(t, "7H", card.String(),
		"should play lowest over when last to play (guaranteed win)")
}

// --- Hard moonshot pass voids short off-suits ---

func TestHardMoonShotPassVoidsShortSuit(t *testing.T) {
	// Strong moonshot hand with 2 low spades (not in any top run).
	// Should pass the spades to create a void.
	hand := parseCards(t, []string{
		"AH", "KH", "QH", "JH", "TH",
		"AC", "KC", "QC",
		"AD", "KD",
		"2S", "3S", "4S",
	})

	cards, err := NewHardBot().ChoosePass(game.PassInput{Hand: hand, Direction: game.PassDirectionLeft})
	require.NoError(t, err)
	require.Len(t, cards, 3)

	// All 3 spades should be passed (they're non-run and ≤3 cards).
	for _, c := range cards {
		require.Equal(t, game.SuitSpades, c.Suit,
			"moonshot pass should void the short spade suit, got %s", c)
	}
}

// --- Near-safe card counting ---

func TestCountNearSafeCards(t *testing.T) {
	// K♣ with A♣ played → safe (not near-safe)
	// K♦ with A♦ not played and not in hand → near-safe (1 gap: A♦)
	// Q♦ with K♦ in hand, A♦ missing → near-safe (1 gap: A♦)
	// J♦ with Q♦ in hand, K♦ in hand, A♦ missing → near-safe (1 gap: A♦)
	// 5♠ with many higher cards missing → NOT near-safe (>1 gap)
	hand := parseCards(t, []string{"KC", "KD", "QD", "JD", "5S"})
	played := parseCards(t, []string{"AC"})

	count := countNearSafeCards(hand, played)
	// KC is safe (AC played). KD, QD, JD each have 1 gap (AD). 5S has many gaps.
	require.Equal(t, 3, count, "expected 3 near-safe cards (KD, QD, JD)")
}

// --- Moon-shot prevention (blocking) ---

func TestDetectMoonShooterFindsOpponentWithAllPenalties(t *testing.T) {
	// 5 completed tricks, 14 penalty points, all held by seat 2.
	played := []game.Play{
		{Seat: 1, Card: c("QS")}, {Seat: 2, Card: c("AS")}, {Seat: 3, Card: c("3S")}, {Seat: 0, Card: c("4S")}, // seat 2 wins (AS)
		{Seat: 2, Card: c("2H")}, {Seat: 3, Card: c("5C")}, {Seat: 0, Card: c("6C")}, {Seat: 1, Card: c("7C")}, // seat 2 wins (only heart)
		{Seat: 2, Card: c("AC")}, {Seat: 3, Card: c("3C")}, {Seat: 0, Card: c("4C")}, {Seat: 1, Card: c("5D")}, // no penalty
		{Seat: 2, Card: c("AD")}, {Seat: 3, Card: c("7D")}, {Seat: 0, Card: c("8D")}, {Seat: 1, Card: c("9D")}, // no penalty
		{Seat: 2, Card: c("KS")}, {Seat: 3, Card: c("JS")}, {Seat: 0, Card: c("TS")}, {Seat: 1, Card: c("9S")}, // no penalty
	}
	roundPoints := [game.PlayersPerTable]game.Points{0, 0, 14, 0}

	target := detectMoonShooter(roundPoints, played, 0)
	require.Equal(t, 2, target, "should detect seat 2 as shooter")
}

func TestDetectMoonShooterReturnsNegativeWhenNoShooter(t *testing.T) {
	// 5 completed tricks, penalties split between two players.
	played := []game.Play{
		{Seat: 0, Card: c("QS")}, {Seat: 1, Card: c("AS")}, {Seat: 2, Card: c("3S")}, {Seat: 3, Card: c("4S")}, // seat 1 wins
		{Seat: 2, Card: c("2H")}, {Seat: 3, Card: c("AH")}, {Seat: 0, Card: c("5C")}, {Seat: 1, Card: c("6C")}, // seat 3 wins (AH)
		{Seat: 2, Card: c("3H")}, {Seat: 3, Card: c("KH")}, {Seat: 0, Card: c("8C")}, {Seat: 1, Card: c("9C")}, // seat 3 wins
		{Seat: 0, Card: c("2C")}, {Seat: 1, Card: c("3C")}, {Seat: 2, Card: c("4C")}, {Seat: 3, Card: c("5D")}, // no penalty
		{Seat: 0, Card: c("6D")}, {Seat: 1, Card: c("7D")}, {Seat: 2, Card: c("8D")}, {Seat: 3, Card: c("9D")}, // no penalty
	}
	roundPoints := [game.PlayersPerTable]game.Points{0, 13, 2, 0} // split

	target := detectMoonShooter(roundPoints, played, 1)
	require.Equal(t, -1, target, "should not detect shooter when penalties are split")
}

func TestDetectMoonShooterIgnoresSelf(t *testing.T) {
	// Bot (seat 0) has all penalties — should not return self as target.
	played := []game.Play{
		{Seat: 0, Card: c("AS")}, {Seat: 1, Card: c("QS")}, {Seat: 2, Card: c("3S")}, {Seat: 3, Card: c("4S")}, // seat 0 wins
		{Seat: 0, Card: c("AH")}, {Seat: 1, Card: c("5C")}, {Seat: 2, Card: c("6C")}, {Seat: 3, Card: c("7C")}, // seat 0 wins
		{Seat: 0, Card: c("AC")}, {Seat: 1, Card: c("3C")}, {Seat: 2, Card: c("4C")}, {Seat: 3, Card: c("5D")}, // no penalty
		{Seat: 0, Card: c("AD")}, {Seat: 1, Card: c("7D")}, {Seat: 2, Card: c("8D")}, {Seat: 3, Card: c("9D")}, // no penalty
		{Seat: 0, Card: c("KS")}, {Seat: 1, Card: c("JS")}, {Seat: 2, Card: c("TS")}, {Seat: 3, Card: c("9S")}, // no penalty
	}
	roundPoints := [game.PlayersPerTable]game.Points{14, 0, 0, 0}

	target := detectMoonShooter(roundPoints, played, 0)
	require.Equal(t, -1, target, "should not detect self as shooter")
}

func TestDetectMoonShooterRequiresMinTricks(t *testing.T) {
	// Only 2 completed tricks — too early (need 3+).
	played := []game.Play{
		{Seat: 1, Card: c("QS")}, {Seat: 2, Card: c("AS")}, {Seat: 3, Card: c("3S")}, {Seat: 0, Card: c("4S")},
		{Seat: 2, Card: c("2H")}, {Seat: 3, Card: c("5C")}, {Seat: 0, Card: c("6C")}, {Seat: 1, Card: c("7C")},
	}
	roundPoints := [game.PlayersPerTable]game.Points{0, 0, 14, 0}

	target := detectMoonShooter(roundPoints, played, 0)
	require.Equal(t, -1, target, "should not detect shooter before trick 3")
}

func TestDetectMoonShooterRequiresMinPenalty(t *testing.T) {
	// 4 completed tricks but only 2 penalty points — too low (< 3).
	played := []game.Play{
		{Seat: 1, Card: c("2H")}, {Seat: 2, Card: c("AH")}, {Seat: 3, Card: c("5C")}, {Seat: 0, Card: c("6C")},
		{Seat: 2, Card: c("3H")}, {Seat: 3, Card: c("KH")}, {Seat: 0, Card: c("7C")}, {Seat: 1, Card: c("8C")},
		{Seat: 0, Card: c("2C")}, {Seat: 1, Card: c("3C")}, {Seat: 2, Card: c("4C")}, {Seat: 3, Card: c("5D")},
		{Seat: 0, Card: c("6D")}, {Seat: 1, Card: c("7D")}, {Seat: 2, Card: c("8D")}, {Seat: 3, Card: c("9D")},
	}
	roundPoints := [game.PlayersPerTable]game.Points{0, 0, 2, 0}

	target := detectMoonShooter(roundPoints, played, 0)
	require.Equal(t, -1, target, "should not detect shooter with < 3 penalty points")
}

func TestDetectMoonShooterEarlyDetectionBeforeQueenSpades(t *testing.T) {
	// 3 completed tricks, seat 1 won all penalty tricks (5 heart points).
	// Q♠ not yet taken, but the pattern is consistent — detect early.
	played := []game.Play{
		{Seat: 0, Card: c("2H")}, {Seat: 1, Card: c("AH")}, {Seat: 2, Card: c("5C")}, {Seat: 3, Card: c("6C")}, // seat 1 wins (AH)
		{Seat: 1, Card: c("KH")}, {Seat: 2, Card: c("3H")}, {Seat: 3, Card: c("4H")}, {Seat: 0, Card: c("7C")}, // seat 1 wins (KH)
		{Seat: 1, Card: c("AC")}, {Seat: 2, Card: c("3C")}, {Seat: 3, Card: c("4C")}, {Seat: 0, Card: c("5D")}, // no penalty
	}
	roundPoints := [game.PlayersPerTable]game.Points{0, 5, 0, 0}

	target := detectMoonShooter(roundPoints, played, 0)
	require.Equal(t, 1, target, "should detect seat 1 as early shooter via trick-winner pattern")
}

func TestDetectMoonShooterEarlyRequiresConsistentWinner(t *testing.T) {
	// 3 completed tricks with hearts, but penalty tricks won by different players.
	played := []game.Play{
		{Seat: 0, Card: c("2H")}, {Seat: 1, Card: c("AH")}, {Seat: 2, Card: c("5C")}, {Seat: 3, Card: c("6C")}, // seat 1 wins
		{Seat: 2, Card: c("3H")}, {Seat: 3, Card: c("KH")}, {Seat: 0, Card: c("7C")}, {Seat: 1, Card: c("8C")}, // seat 3 wins
		{Seat: 0, Card: c("2C")}, {Seat: 1, Card: c("3C")}, {Seat: 2, Card: c("4C")}, {Seat: 3, Card: c("5D")},
	}
	roundPoints := [game.PlayersPerTable]game.Points{0, 1, 0, 2} // split

	target := detectMoonShooter(roundPoints, played, 0)
	require.Equal(t, -1, target, "should not detect shooter when penalty tricks are split")
}

func TestHardBlockMoonLeadPrefersSafeHeart(t *testing.T) {
	// Bot holds AH (safe — no higher heart exists) and low non-hearts.
	hand := parseCards(t, []string{"AH", "2C", "3D"})
	legal := hand // all legal to lead (hearts broken)
	played := parseCards(t, []string{})

	card := hardBlockMoonLead(hand, legal, played)
	require.Equal(t, "AH", card.String(), "block lead should play safe high heart")
}

func TestHardBlockMoonLeadFallsBackWhenNoSafeHeart(t *testing.T) {
	// Bot holds KH but AH not played — KH is not safe. Should fall back to defensive.
	hand := parseCards(t, []string{"KH", "2C", "3D"})
	legal := hand
	played := parseCards(t, []string{})

	card := hardBlockMoonLead(hand, legal, played)
	// Should NOT lead KH (not safe), should lead 2C or 3D defensively.
	require.NotEqual(t, "KH", card.String(), "block lead should not lead unsafe heart")
}

// --- Block discard: avoid feeding shooter ---

func TestHardBlockMoonDiscardHoldsPenaltiesWhenShooterWins(t *testing.T) {
	// Shooter (seat 1) has played and is winning. Don't feed penalties.
	trick := []game.Play{
		{Seat: 1, Card: c("AC")},
		{Seat: 2, Card: c("3C")},
	}
	legal := parseCards(t, []string{"QS", "AH", "5D"})

	card := hardBlockMoonDiscard(trick, legal, 1)
	require.False(t, game.IsPenaltyCard(card),
		"should not dump penalty card when shooter is winning, got %s", card)
}

func TestHardBlockMoonDiscardDumpsPenaltiesWhenNonShooterWins(t *testing.T) {
	// Non-shooter (seat 3) is winning. Dump penalties to break the shoot.
	trick := []game.Play{
		{Seat: 3, Card: c("AC")},
		{Seat: 1, Card: c("3C")}, // shooter played but is losing
	}
	legal := parseCards(t, []string{"QS", "AH", "5D"})

	card := hardBlockMoonDiscard(trick, legal, 1)
	// Normal defensive discard should dump QS.
	require.Equal(t, "QS", card.String(), "should dump QS when non-shooter is winning")
}

func TestHardBlockMoonDiscardDumpsWhenShooterHasNotPlayed(t *testing.T) {
	// Shooter (seat 3) hasn't played yet — uncertain, dump normally.
	trick := []game.Play{
		{Seat: 0, Card: c("AC")},
		{Seat: 1, Card: c("3C")},
	}
	legal := parseCards(t, []string{"QS", "AH", "5D"})

	card := hardBlockMoonDiscard(trick, legal, 3)
	// Normal defensive discard — dump QS.
	require.Equal(t, "QS", card.String(), "should dump penalties when shooter hasn't played")
}

// --- Q♠-aware discard ---

func TestHardDiscardDumpsAceSpadeBeforeHeartsWhenQueenOut(t *testing.T) {
	// Q♠ at large — A♠ risks winning Q♠ (13 pts) later. Dump it before hearts.
	legal := parseCards(t, []string{"AS", "KH", "2H"})
	played := parseCards(t, []string{}) // Q♠ not played

	card := hardChooseDiscard(legal, played, false)
	require.Equal(t, "AS", card.String(), "should dump A♠ before hearts when Q♠ is at large")
}

func TestHardDiscardDumpsKingSpadeWhenQueenOut(t *testing.T) {
	// Q♠ at large — K♠ also risks winning Q♠.
	legal := parseCards(t, []string{"KS", "AH", "2H"})
	played := parseCards(t, []string{})

	card := hardChooseDiscard(legal, played, false)
	require.Equal(t, "KS", card.String(), "should dump K♠ before hearts when Q♠ is at large")
}

func TestHardDiscardPrefersAceOverKingSpade(t *testing.T) {
	// Both A♠ and K♠ available — dump A♠ first (higher risk).
	legal := parseCards(t, []string{"AS", "KS", "2H"})
	played := parseCards(t, []string{})

	card := hardChooseDiscard(legal, played, false)
	require.Equal(t, "AS", card.String(), "should dump A♠ before K♠")
}

func TestHardDiscardDumpsQueenSpadeFirst(t *testing.T) {
	// Q♠ in hand — always dump it first regardless of other cards.
	legal := parseCards(t, []string{"QS", "AS", "AH"})
	played := parseCards(t, []string{})

	card := hardChooseDiscard(legal, played, false)
	require.Equal(t, "QS", card.String(), "should dump Q♠ first")
}

func TestHardDiscardNormalAfterQueenPlayed(t *testing.T) {
	// Q♠ already played — A♠/K♠ are safe. Standard discard (hearts first).
	legal := parseCards(t, []string{"AS", "AH", "2D"})
	played := parseCards(t, []string{"QS"})

	card := hardChooseDiscard(legal, played, false)
	require.Equal(t, "AH", card.String(), "should dump heart when Q♠ is already played")
}

func TestHardDiscardIntegration(t *testing.T) {
	// Full ChoosePlay: void in clubs, Q♠ at large, holds A♠ and hearts.
	hand := parseCards(t, []string{"AS", "KH", "3H", "5D"})

	card, err := NewHardBot().ChoosePlay(game.TurnInput{
		Hand:         hand,
		Trick:        parsePlays(t, []string{"5C"}), // clubs led, bot void
		HeartsBroken: true,
		PlayedCards:  nil,
	})
	require.NoError(t, err)
	require.Equal(t, "AS", card.String(), "should dump A♠ when void in clubs and Q♠ at large")
}

// --- Pass requires three cards ---

func TestHardChoosePassRequiresThreeCards(t *testing.T) {
	hand := parseCards(t, []string{"KC", "3D"})

	_, err := NewHardBot().ChoosePass(game.PassInput{Hand: hand})
	require.ErrorIs(t, err, ErrNotEnoughCards)
}
