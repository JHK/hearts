package bot

import (
	"testing"

	"github.com/JHK/hearts/internal/game"
)

func TestFirstLegalChoosePlayUsesHandOrder(t *testing.T) {
	hand := parseCards(t, []string{"KC", "3D", "2S"})

	card, err := NewFirstLegalBot().ChoosePlay(game.TurnInput{
		Hand:         hand,
		Trick:        nil,
		HeartsBroken: true,
		FirstTrick:   false,
	})
	if err != nil {
		t.Fatalf("expected valid play, got %v", err)
	}

	if card.String() != "KC" {
		t.Fatalf("expected first legal card KC, got %s", card)
	}
}

func TestFirstLegalChoosePlayFollowsSuit(t *testing.T) {
	hand := parseCards(t, []string{"AH", "3C", "KS"})

	card, err := NewFirstLegalBot().ChoosePlay(game.TurnInput{
		Hand:         hand,
		Trick:        parseCards(t, []string{"2S"}),
		HeartsBroken: true,
		FirstTrick:   false,
	})
	if err != nil {
		t.Fatalf("expected valid play, got %v", err)
	}

	if card.String() != "KS" {
		t.Fatalf("expected first legal card KS, got %s", card)
	}
}

func TestFirstLegalChoosePlayFirstTrickLead(t *testing.T) {
	hand := parseCards(t, []string{"3D", "2C", "AS"})

	card, err := NewFirstLegalBot().ChoosePlay(game.TurnInput{
		Hand:         hand,
		Trick:        nil,
		HeartsBroken: false,
		FirstTrick:   true,
	})
	if err != nil {
		t.Fatalf("expected valid play, got %v", err)
	}

	if card.String() != "2C" {
		t.Fatalf("expected first legal card 2C, got %s", card)
	}
}

func TestFirstLegalChoosePlayNoLegalCards(t *testing.T) {
	_, err := NewFirstLegalBot().ChoosePlay(game.TurnInput{})
	if err == nil {
		t.Fatalf("expected no legal plays error")
	}
}

func TestFirstLegalChoosePassUsesFirstThreeInOrder(t *testing.T) {
	hand := parseCards(t, []string{"KC", "3D", "2S", "AH"})

	cards, err := NewFirstLegalBot().ChoosePass(game.PassInput{Hand: hand, Direction: game.PassDirectionLeft})
	if err != nil {
		t.Fatalf("expected pass cards, got %v", err)
	}

	if len(cards) != 3 {
		t.Fatalf("expected 3 pass cards, got %d", len(cards))
	}

	if cards[0].String() != "KC" || cards[1].String() != "3D" || cards[2].String() != "2S" {
		t.Fatalf("expected first three cards KC,3D,2S got %s,%s,%s", cards[0], cards[1], cards[2])
	}
}

func TestFirstLegalChoosePassRequiresThreeCards(t *testing.T) {
	hand := parseCards(t, []string{"KC", "3D"})

	_, err := NewFirstLegalBot().ChoosePass(game.PassInput{Hand: hand})
	if err == nil {
		t.Fatalf("expected not enough cards error")
	}
}

func parseCards(t *testing.T, raw []string) []game.Card {
	t.Helper()

	cards, err := game.ParseCards(raw)
	if err != nil {
		t.Fatalf("parse cards: %v", err)
	}

	return cards
}
