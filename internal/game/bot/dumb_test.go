package bot

import "testing"

func TestDumbChoosePlayLeadsShortestNonHeartSuit(t *testing.T) {
	hand := parseCards(t, []string{"3D", "9D", "2S", "KH"})

	card, err := NewDumbBot().ChoosePlay(TurnInput{
		Hand:         hand,
		Trick:        nil,
		HeartsBroken: false,
		FirstTrick:   false,
	})
	if err != nil {
		t.Fatalf("expected valid play, got %v", err)
	}

	if card.String() != "2S" {
		t.Fatalf("expected dumb lead 2S, got %s", card)
	}
}

func TestDumbChoosePlayAvoidsTakingPenaltyTrick(t *testing.T) {
	hand := parseCards(t, []string{"TS", "KS", "3C"})

	card, err := NewDumbBot().ChoosePlay(TurnInput{
		Hand:         hand,
		Trick:        parseCards(t, []string{"JS", "QS"}),
		HeartsBroken: true,
		FirstTrick:   false,
	})
	if err != nil {
		t.Fatalf("expected valid play, got %v", err)
	}

	if card.String() != "TS" {
		t.Fatalf("expected dumb follow TS, got %s", card)
	}
}

func TestDumbChoosePlayDiscardsQueenOfSpades(t *testing.T) {
	hand := parseCards(t, []string{"QS", "AH", "2D"})

	card, err := NewDumbBot().ChoosePlay(TurnInput{
		Hand:         hand,
		Trick:        parseCards(t, []string{"5C"}),
		HeartsBroken: true,
		FirstTrick:   false,
	})
	if err != nil {
		t.Fatalf("expected valid play, got %v", err)
	}

	if card.String() != "QS" {
		t.Fatalf("expected dumb discard QS, got %s", card)
	}
}

func TestDumbChoosePassPrefersDangerousCards(t *testing.T) {
	hand := parseCards(t, []string{"QS", "AS", "KH", "2C", "3C", "4D", "5D", "6H", "7H", "8S", "9S", "TC", "JD"})

	cards, err := NewDumbBot().ChoosePass(PassInput{Hand: hand, Direction: "left"})
	if err != nil {
		t.Fatalf("expected pass cards, got %v", err)
	}

	if len(cards) != 3 {
		t.Fatalf("expected 3 pass cards, got %d", len(cards))
	}

	want := map[string]struct{}{"QS": {}, "AS": {}, "KH": {}}
	for _, card := range cards {
		if _, ok := want[card.String()]; !ok {
			t.Fatalf("expected dumb pass cards QS,AS,KH got %s,%s,%s", cards[0], cards[1], cards[2])
		}
	}
}

func TestDumbChoosePassRequiresThreeCards(t *testing.T) {
	hand := parseCards(t, []string{"KC", "3D"})

	_, err := NewDumbBot().ChoosePass(PassInput{Hand: hand})
	if err == nil {
		t.Fatalf("expected not enough cards error")
	}
}
