package bot

import (
	"testing"

	"github.com/JHK/hearts/internal/game"
	"github.com/stretchr/testify/require"
)

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
	// Defensive pass should send the three highest-risk cards.
	hand := parseCards(t, []string{
		"3C",
		"5D", "7D", "9D", "JD",
		"2S", "4S", "6S", "8S",
		"2H", "4H", "6H", "8H",
	})

	cards, err := NewHardBot().ChoosePass(game.PassInput{Hand: hand, Direction: "left"})
	require.NoError(t, err)
	require.Len(t, cards, 3)

	// Defensive pass: highest-risk cards. JD (rank 11) and 8S/8H are the riskiest non-penalty cards.
	// Just verify 3 cards are returned and none is an obviously wrong choice (e.g. a 2).
	for _, c := range cards {
		require.False(t, c.Rank == 2 || c.Rank == 3, "defensive pass should not include very low card %s", c)
	}
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

// --- Card tracking / safe-high-card play ---

func TestHardPlaysAceWhenOnlyLegalOption(t *testing.T) {
	// Holds A♣ as the only non-heart; K♣ and Q♣ already played → A♣ is safe and only legal lead.
	hand := parseCards(t, []string{"AC", "7H"})
	played := parseCards(t, []string{"KC", "QC"})

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
	played := parseCards(t, []string{"QC"}) // A♣ still outstanding

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
		Trick:        parseCards(t, []string{"5C"}),
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
		Trick:        parseCards(t, []string{"5C"}),
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
		Trick:         parseCards(t, []string{"3D"}), // someone else led
		HeartsBroken:  false,
		FirstTrick:    false,
		PlayedCards: parseCards(t, []string{"2C", "5D", "7S", "8H"}), // 1 completed trick
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
	played := parseCards(t, []string{
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
		Trick:         parseCards(t, []string{"6H"}), // someone else led
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
		PlayedCards:  parseCards(t, []string{"2C", "5D", "7S", "8H"}),
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

// --- Pass requires three cards ---

func TestHardChoosePassRequiresThreeCards(t *testing.T) {
	hand := parseCards(t, []string{"KC", "3D"})

	_, err := NewHardBot().ChoosePass(game.PassInput{Hand: hand})
	require.ErrorIs(t, err, ErrNotEnoughCards)
}
