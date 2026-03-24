package bot

import (
	"math/rand"
	"testing"

	"github.com/JHK/hearts/internal/game"
	"github.com/stretchr/testify/require"
)

func TestRandomChoosePassReturnsThreeDistinctCardsFromHand(t *testing.T) {
	hand := mustParseCards(t, []string{"KC", "3D", "2S", "AH", "7C"})
	strategy := NewRandomBot(rand.New(rand.NewSource(7)))

	cards, err := strategy.ChoosePass(game.PassInput{Hand: hand, Direction: game.PassDirectionLeft})
	require.NoError(t, err)
	require.Len(t, cards, 3)

	seen := map[game.Card]struct{}{}
	for _, card := range cards {
		_, exists := seen[card]
		require.False(t, exists, "expected distinct cards, got duplicate %s", card)
		seen[card] = struct{}{}
		require.True(t, game.ContainsCard(hand, card), "expected selected card %s to be in original hand", card)
	}
}

func TestRandomChoosePassRequiresThreeCards(t *testing.T) {
	hand := mustParseCards(t, []string{"KC", "3D"})
	strategy := NewRandomBot(rand.New(rand.NewSource(7)))

	_, err := strategy.ChoosePass(game.PassInput{Hand: hand})
	require.ErrorIs(t, err, ErrNotEnoughCards)
}

func mustParseCards(t *testing.T, raw []string) []game.Card {
	t.Helper()

	cards, err := game.ParseCards(raw)
	require.NoError(t, err, "parse cards")

	return cards
}
