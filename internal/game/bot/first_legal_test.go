package bot

import (
	"testing"

	"github.com/JHK/hearts/internal/game"
	"github.com/stretchr/testify/require"
)

func TestFirstLegalChoosePlayUsesHandOrder(t *testing.T) {
	hand := parseCards(t, []string{"KC", "3D", "2S"})

	card, err := NewFirstLegalBot().ChoosePlay(game.TurnInput{
		Hand:         hand,
		Trick:        nil,
		HeartsBroken: true,
		FirstTrick:   false,
	})
	require.NoError(t, err)
	require.Equal(t, "KC", card.String())
}

func TestFirstLegalChoosePlayFollowsSuit(t *testing.T) {
	hand := parseCards(t, []string{"AH", "3C", "KS"})

	card, err := NewFirstLegalBot().ChoosePlay(game.TurnInput{
		Hand:         hand,
		Trick:        parsePlays(t, []string{"2S"}),
		HeartsBroken: true,
		FirstTrick:   false,
	})
	require.NoError(t, err)
	require.Equal(t, "KS", card.String())
}

func TestFirstLegalChoosePlayFirstTrickLead(t *testing.T) {
	hand := parseCards(t, []string{"3D", "2C", "AS"})

	card, err := NewFirstLegalBot().ChoosePlay(game.TurnInput{
		Hand:         hand,
		Trick:        nil,
		HeartsBroken: false,
		FirstTrick:   true,
	})
	require.NoError(t, err)
	require.Equal(t, "2C", card.String())
}

func TestFirstLegalChoosePlayNoLegalCards(t *testing.T) {
	_, err := NewFirstLegalBot().ChoosePlay(game.TurnInput{})
	require.ErrorIs(t, err, ErrNoLegalPlays)
}

func TestFirstLegalChoosePassUsesFirstThreeInOrder(t *testing.T) {
	hand := parseCards(t, []string{"KC", "3D", "2S", "AH"})

	cards, err := NewFirstLegalBot().ChoosePass(game.PassInput{Hand: hand, Direction: game.PassDirectionLeft})
	require.NoError(t, err)
	require.Len(t, cards, 3)
	require.Equal(t, "KC", cards[0].String())
	require.Equal(t, "3D", cards[1].String())
	require.Equal(t, "2S", cards[2].String())
}

func TestFirstLegalChoosePassRequiresThreeCards(t *testing.T) {
	hand := parseCards(t, []string{"KC", "3D"})

	_, err := NewFirstLegalBot().ChoosePass(game.PassInput{Hand: hand})
	require.ErrorIs(t, err, ErrNotEnoughCards)
}

func parseCards(t *testing.T, raw []string) []game.Card {
	t.Helper()

	cards, err := game.ParseCards(raw)
	require.NoError(t, err, "parse cards")

	return cards
}

// parsePlays creates plays from card strings with sequential seat assignment (0,1,2,3,0,...).
// Suitable for constructing trick and played-cards data in tests.
func parsePlays(t *testing.T, raw []string) []game.Play {
	t.Helper()
	cards := parseCards(t, raw)
	plays := make([]game.Play, len(cards))
	for i, c := range cards {
		plays[i] = game.Play{Seat: i % game.PlayersPerTable, Card: c}
	}
	return plays
}
