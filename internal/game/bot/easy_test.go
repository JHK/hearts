package bot

import (
	"testing"

	"github.com/JHK/hearts/internal/game"
	"github.com/stretchr/testify/require"
)

func TestEasyChoosePlayLeadsShortestNonHeartSuit(t *testing.T) {
	hand := parseCards(t, []string{"3D", "9D", "2S", "KH"})

	card, err := NewEasyBot().ChoosePlay(game.TurnInput{
		Hand:         hand,
		Trick:        nil,
		HeartsBroken: false,
		FirstTrick:   false,
	})
	require.NoError(t, err)
	require.Equal(t, "2S", card.String())
}

func TestEasyChoosePlayAvoidsTakingPenaltyTrick(t *testing.T) {
	hand := parseCards(t, []string{"TS", "KS", "3C"})

	card, err := NewEasyBot().ChoosePlay(game.TurnInput{
		Hand:         hand,
		Trick:        parseCards(t, []string{"JS", "QS"}),
		HeartsBroken: true,
		FirstTrick:   false,
	})
	require.NoError(t, err)
	require.Equal(t, "TS", card.String())
}

func TestEasyChoosePlayDiscardsQueenOfSpades(t *testing.T) {
	hand := parseCards(t, []string{"QS", "AH", "2D"})

	card, err := NewEasyBot().ChoosePlay(game.TurnInput{
		Hand:         hand,
		Trick:        parseCards(t, []string{"5C"}),
		HeartsBroken: true,
		FirstTrick:   false,
	})
	require.NoError(t, err)
	require.Equal(t, "QS", card.String())
}

func TestEasyChoosePassPrefersDangerousCards(t *testing.T) {
	hand := parseCards(t, []string{"QS", "AS", "KH", "2C", "3C", "4D", "5D", "6H", "7H", "8S", "9S", "TC", "JD"})

	cards, err := NewEasyBot().ChoosePass(game.PassInput{Hand: hand, Direction: "left"})
	require.NoError(t, err)
	require.Len(t, cards, 3)

	want := map[string]struct{}{"QS": {}, "AS": {}, "KH": {}}
	for _, card := range cards {
		_, ok := want[card.String()]
		require.True(t, ok, "expected easy pass cards QS,AS,KH got %s,%s,%s", cards[0], cards[1], cards[2])
	}
}

func TestEasyChoosePassRequiresThreeCards(t *testing.T) {
	hand := parseCards(t, []string{"KC", "3D"})

	_, err := NewEasyBot().ChoosePass(game.PassInput{Hand: hand})
	require.ErrorIs(t, err, ErrNotEnoughCards)
}
