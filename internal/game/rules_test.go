package game

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidatePlayFirstTrickLeadMustBeTwoClubs(t *testing.T) {
	hand := []Card{{Suit: SuitClubs, Rank: 2}, {Suit: SuitHearts, Rank: 5}}

	err := ValidatePlay(ValidatePlayInput{
		Hand:       hand,
		Card:       Card{Suit: SuitHearts, Rank: 5},
		FirstTrick: true,
	})
	require.ErrorIs(t, err, ErrMustLeadTwoClubs)

	err = ValidatePlay(ValidatePlayInput{
		Hand:       hand,
		Card:       Card{Suit: SuitClubs, Rank: 2},
		FirstTrick: true,
	})
	require.NoError(t, err)
}

func TestValidatePlayMustFollowSuit(t *testing.T) {
	hand := []Card{{Suit: SuitClubs, Rank: 10}, {Suit: SuitHearts, Rank: 4}}

	err := ValidatePlay(ValidatePlayInput{
		Hand:       hand,
		Card:       Card{Suit: SuitHearts, Rank: 4},
		Trick:      []Card{{Suit: SuitClubs, Rank: 2}},
		FirstTrick: false,
	})
	require.ErrorIs(t, err, ErrMustFollowSuit)
}

func TestTrickWinnerAndPoints(t *testing.T) {
	winner, points, err := TrickWinner([]Play{
		{Seat: 0, Card: Card{Suit: SuitSpades, Rank: 10}},
		{Seat: 1, Card: Card{Suit: SuitSpades, Rank: 14}},
		{Seat: 2, Card: Card{Suit: SuitHearts, Rank: 2}},
		{Seat: 3, Card: Card{Suit: SuitSpades, Rank: 12}},
	})
	require.NoError(t, err)
	require.Equal(t, 1, winner)
	require.Equal(t, Points(14), points)
}

func TestApplyShootTheMoon(t *testing.T) {
	adjusted := ApplyShootTheMoon([PlayersPerTable]Points{26, 0, 0, 0})

	require.Equal(t, Points(0), adjusted[0])
	require.Equal(t, Points(26), adjusted[1])
	require.Equal(t, Points(26), adjusted[2])
	require.Equal(t, Points(26), adjusted[3])
}
