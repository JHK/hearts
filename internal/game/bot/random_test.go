package bot

import (
	"math/rand"
	"testing"

	"github.com/JHK/hearts/internal/game"
)

func TestRandomChoosePassReturnsThreeDistinctCardsFromHand(t *testing.T) {
	hand := mustParseCards(t, []string{"KC", "3D", "2S", "AH", "7C"})
	strategy := NewRandomBot(rand.New(rand.NewSource(7)))

	cards, err := strategy.ChoosePass(PassInput{Hand: hand, Direction: game.PassDirectionLeft})
	if err != nil {
		t.Fatalf("expected pass cards, got %v", err)
	}

	if len(cards) != 3 {
		t.Fatalf("expected 3 pass cards, got %d", len(cards))
	}

	seen := map[game.Card]struct{}{}
	for _, card := range cards {
		if _, exists := seen[card]; exists {
			t.Fatalf("expected distinct cards, got duplicate %s", card)
		}
		seen[card] = struct{}{}

		if !game.ContainsCard(hand, card) {
			t.Fatalf("expected selected card %s to be in original hand", card)
		}
	}
}

func TestRandomChoosePassRequiresThreeCards(t *testing.T) {
	hand := mustParseCards(t, []string{"KC", "3D"})
	strategy := NewRandomBot(rand.New(rand.NewSource(7)))

	_, err := strategy.ChoosePass(PassInput{Hand: hand})
	if err == nil {
		t.Fatalf("expected not enough cards error")
	}
}

func mustParseCards(t *testing.T, raw []string) []game.Card {
	t.Helper()

	cards, err := game.ParseCards(raw)
	if err != nil {
		t.Fatalf("parse cards: %v", err)
	}

	return cards
}
