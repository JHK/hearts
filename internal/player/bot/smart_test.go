package bot

import (
	"testing"

	"github.com/JHK/hearts/internal/game"
)

// --- Trade strategy selection ---

func TestSmartPassSelectsMoonShotStrategy(t *testing.T) {
	// 5 top hearts + 3 top clubs + 2 top diamonds = 10 guaranteed tricks → moon shot
	hand := parseCards(t, []string{
		"AH", "KH", "QH", "JH", "TH",
		"AC", "KC", "QC",
		"AD", "KD",
		"2S", "3S", "4S",
	})

	cards, err := NewSmartBot().ChoosePass(PassInput{Hand: hand, Direction: game.PassDirectionLeft})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cards) != 3 {
		t.Fatalf("expected 3 cards, got %d", len(cards))
	}

	// Should not pass any of the run cards (the moon-shot backbone)
	moonShotKeep := map[string]struct{}{
		"AH": {}, "KH": {}, "QH": {}, "JH": {}, "TH": {},
		"AC": {}, "KC": {}, "QC": {},
		"AD": {}, "KD": {},
	}
	for _, c := range cards {
		if _, ok := moonShotKeep[c.String()]; ok {
			t.Fatalf("moon-shot pass should not include run card %s", c)
		}
	}
}

func TestSmartPassSelectsDefensiveStrategyForShortSuit(t *testing.T) {
	// Only 1 club but no moon-shot potential — falls through to defensive pass.
	// Defensive pass should send the three highest-risk cards.
	hand := parseCards(t, []string{
		"3C",
		"5D", "7D", "9D", "JD",
		"2S", "4S", "6S", "8S",
		"2H", "4H", "6H", "8H",
	})

	cards, err := NewSmartBot().ChoosePass(PassInput{Hand: hand, Direction: "left"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cards) != 3 {
		t.Fatalf("expected 3 cards, got %d", len(cards))
	}

	// Defensive pass: highest-risk cards. JD (rank 11) and 8S/8H are the riskiest non-penalty cards.
	// Just verify 3 cards are returned and none is an obviously wrong choice (e.g. a 2).
	for _, c := range cards {
		if c.Rank == 2 || c.Rank == 3 {
			t.Fatalf("defensive pass should not include very low card %s", c)
		}
	}
}

func TestSmartPassSelectsDefensiveStrategy(t *testing.T) {
	// Balanced hand, no moon-shot dominance, no near-void suit → defensive
	hand := parseCards(t, []string{
		"QS", "AS", "KH", "2C", "3C", "4D", "5D", "6H", "7H", "8S", "9S", "TC", "JD",
	})

	cards, err := NewSmartBot().ChoosePass(PassInput{Hand: hand, Direction: "left"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cards) != 3 {
		t.Fatalf("expected 3 cards, got %d", len(cards))
	}

	// Defensive: most dangerous cards (QS, AS, KH)
	want := map[string]struct{}{"QS": {}, "AS": {}, "KH": {}}
	for _, c := range cards {
		if _, ok := want[c.String()]; !ok {
			t.Fatalf("defensive pass expected QS/AS/KH, got unexpected %s", c)
		}
	}
}

// --- Card tracking / safe-high-card play ---

func TestSmartPlaysAceWhenOnlyLegalOption(t *testing.T) {
	// Holds A♣ as the only non-heart; K♣ and Q♣ already played → A♣ is safe and only legal lead.
	hand := parseCards(t, []string{"AC", "7H"})
	played := parseCards(t, []string{"KC", "QC"})

	card, err := NewSmartBot().ChoosePlay(TurnInput{
		Hand:        hand,
		Trick:       nil,
		PlayedCards: played,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if card.String() != "AC" {
		t.Fatalf("expected safe lead AC (only legal non-heart), got %s", card)
	}
}

func TestSmartDoesNotLeadUnsafeKing(t *testing.T) {
	// Holds K♣ but A♣ not yet played → K♣ is not safe (A♣ is still out)
	// The bot should prefer leading a shorter/safer suit instead.
	hand := parseCards(t, []string{"KC", "2D", "3D", "4D"})
	played := parseCards(t, []string{"QC"}) // A♣ still outstanding

	card, err := NewSmartBot().ChoosePlay(TurnInput{
		Hand:        hand,
		Trick:       nil,
		PlayedCards: played,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should prefer leading a diamond (shorter/safer) over K♣ when A♣ is still out
	if card.String() == "KC" {
		t.Fatalf("should not lead unsafe K♣ when A♣ is still outstanding")
	}
}

// --- Void exploitation ---

func TestSmartDiscardsQueenOfSpadesWhenVoid(t *testing.T) {
	// Can't follow clubs → should dump Q♠
	hand := parseCards(t, []string{"QS", "AH", "2D"})

	card, err := NewSmartBot().ChoosePlay(TurnInput{
		Hand:        hand,
		Trick:       parseCards(t, []string{"5C"}),
		HeartsBroken: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if card.String() != "QS" {
		t.Fatalf("expected void exploitation QS discard, got %s", card)
	}
}

func TestSmartDiscardsHighHeartWhenVoidAndNoQueenSpades(t *testing.T) {
	// Can't follow clubs, no Q♠ → discard highest heart
	hand := parseCards(t, []string{"AH", "KH", "2D"})

	card, err := NewSmartBot().ChoosePlay(TurnInput{
		Hand:        hand,
		Trick:       parseCards(t, []string{"5C"}),
		HeartsBroken: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if card.String() != "AH" {
		t.Fatalf("expected discard AH, got %s", card)
	}
}

// --- Moon-shot triggering and pursuit ---

func TestSmartEvaluatesMoonShotFromHand(t *testing.T) {
	// 5 top hearts + 3 top clubs + 2 top diamonds = 10 guaranteed tricks
	hand := parseCards(t, []string{
		"AH", "KH", "QH", "JH", "TH",
		"AC", "KC", "QC",
		"AD", "KD",
		"2S", "3S", "4S",
	})
	if !evaluateMoonShot(hand) {
		t.Fatal("expected moon-shot potential, got false")
	}
}

func TestSmartDoesNotTriggerMoonShotOnWeakHand(t *testing.T) {
	hand := parseCards(t, []string{
		"2H", "3H", "4H",
		"2C", "3C", "4C", "5C", "6C",
		"2D", "3D", "4D", "5D", "6D",
	})
	if evaluateMoonShot(hand) {
		t.Fatal("weak hand should not trigger moon shot")
	}
}

func TestSmartPursueMoonShotLeadsHighCards(t *testing.T) {
	hand := parseCards(t, []string{"AH", "KH", "QH", "AS"})
	legal := hand

	card := smartChooseLead(hand, legal, nil, true)
	if card.Rank != 14 {
		t.Fatalf("moon shot should lead highest card (A), got %s", card)
	}
}

func TestSmartPursueMoonShotWinsFollowTrick(t *testing.T) {
	// Trick: JS led; bot has QS → should play QS to win
	trick := parseCards(t, []string{"JS"})
	legal := parseCards(t, []string{"QS"}) // must follow spades

	card := smartChooseFollow(trick, legal, nil, nil, true)
	if card.String() != "QS" {
		t.Fatalf("moon shot follow should win trick, got %s", card)
	}
}

// --- Moon-shot abort on penalty leak ---

func TestSmartAbortsMoonShotWhenOtherLeads(t *testing.T) {
	bot := &Smart{moonShotActive: true, moonShotAborted: false, prevPlayedCount: 4}

	// Not first trick; trick already has a card (someone else led) → abort
	hand := parseCards(t, []string{"AH", "KH"})

	_, err := bot.ChoosePlay(TurnInput{
		Hand:        hand,
		Trick:       parseCards(t, []string{"3D"}), // someone else led
		HeartsBroken: false,
		FirstTrick:  false,
		PlayedCards: parseCards(t, []string{"2C", "5D", "7S", "8H"}), // 1 completed trick
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bot.moonShotAborted {
		t.Fatal("expected moon shot to be aborted when another player leads")
	}
}

func TestSmartContinuesMoonShotWhenLeadingEveryTrick(t *testing.T) {
	bot := &Smart{moonShotActive: true, moonShotAborted: false, prevPlayedCount: 4}

	// Bot is leading (trick is empty), not first trick → should NOT abort
	hand := parseCards(t, []string{"AH", "KH", "QH", "JH"})

	_, err := bot.ChoosePlay(TurnInput{
		Hand:        hand,
		Trick:       nil, // bot is leading
		HeartsBroken: true,
		FirstTrick:  false,
		PlayedCards: parseCards(t, []string{"2C", "5D", "7S", "8H"}),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if bot.moonShotAborted {
		t.Fatal("moon shot should not abort when bot is leading every trick")
	}
}

// --- Pass requires three cards ---

func TestSmartChoosePassRequiresThreeCards(t *testing.T) {
	hand := parseCards(t, []string{"KC", "3D"})

	_, err := NewSmartBot().ChoosePass(PassInput{Hand: hand})
	if err == nil {
		t.Fatal("expected not enough cards error")
	}
}
