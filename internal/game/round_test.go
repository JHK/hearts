package game

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoundDealAndHand(t *testing.T) {
	hand0 := []Card{{Suit: SuitClubs, Rank: 2}, {Suit: SuitHearts, Rank: 5}}
	hand1 := []Card{{Suit: SuitDiamonds, Rank: 3}}
	hand2 := []Card{{Suit: SuitSpades, Rank: 7}}
	hand3 := []Card{{Suit: SuitHearts, Rank: 10}}

	r := NewRound([PlayersPerTable][]Card{hand0, hand1, hand2, hand3}, PassDirectionLeft)

	require.Len(t, r.Hand(0), 2)
	require.Len(t, r.Hand(1), 1)
}

func TestRoundSubmitAndApplyPasses(t *testing.T) {
	hands := [PlayersPerTable][]Card{
		{{Suit: SuitClubs, Rank: 2}, {Suit: SuitClubs, Rank: 3}, {Suit: SuitClubs, Rank: 4}, {Suit: SuitHearts, Rank: 5}},
		{{Suit: SuitDiamonds, Rank: 2}, {Suit: SuitDiamonds, Rank: 3}, {Suit: SuitDiamonds, Rank: 4}, {Suit: SuitSpades, Rank: 5}},
		{{Suit: SuitHearts, Rank: 2}, {Suit: SuitHearts, Rank: 3}, {Suit: SuitHearts, Rank: 4}, {Suit: SuitClubs, Rank: 5}},
		{{Suit: SuitSpades, Rank: 2}, {Suit: SuitSpades, Rank: 3}, {Suit: SuitSpades, Rank: 4}, {Suit: SuitDiamonds, Rank: 5}},
	}
	r := NewRound(hands, PassDirectionLeft)

	for seat := 0; seat < PlayersPerTable; seat++ {
		require.NoError(t, r.SubmitPass(seat, r.Hand(seat)[:3]), "submit pass seat %d", seat)
	}

	require.True(t, r.HasSubmittedPass(0))
	require.NoError(t, r.ApplyPasses())
	require.Equal(t, PhasePassReview, r.Phase())

	// After left pass, seat 0 sent to seat 1; seat 0 received from seat 3.
	require.Len(t, r.PassReceived(0), 3)
	require.Len(t, r.Hand(0), 4)
}

func TestRoundMarkReadyAndStartPlaying(t *testing.T) {
	hands := dealTestHands()
	r := NewRound(hands, PassDirectionHold)

	require.NoError(t, r.StartPlaying())
	require.Equal(t, PhasePlaying, r.Phase())
}

func TestRoundPlayValidation(t *testing.T) {
	hands := dealTestHands()
	r := NewRound(hands, PassDirectionHold)
	_ = r.StartPlaying()

	// Wrong seat should fail.
	wrongSeat := (r.TurnSeat() + 1) % PlayersPerTable
	_, err := r.Play(wrongSeat, r.Hand(wrongSeat)[0])
	require.ErrorIs(t, err, ErrNotYourTurn)
}

func TestRoundPlayTrickCompletion(t *testing.T) {
	// Build hands where seat 0 has 2C (leads first trick).
	hands := [PlayersPerTable][]Card{
		{{Suit: SuitClubs, Rank: 2}},
		{{Suit: SuitClubs, Rank: 3}},
		{{Suit: SuitClubs, Rank: 4}},
		{{Suit: SuitClubs, Rank: 5}},
	}
	r := NewRound(hands, PassDirectionHold)
	_ = r.StartPlaying()

	require.Equal(t, 0, r.TurnSeat(), "expected seat 0 to lead (has 2C)")

	// Play all 4 cards.
	for seat := 0; seat < PlayersPerTable; seat++ {
		result, err := r.Play(r.TurnSeat(), r.Hand(r.TurnSeat())[0])
		require.NoError(t, err, "play seat %d", seat)
		if seat < 3 {
			require.Nil(t, result, "expected nil result for card %d", seat+1)
		}
		if seat == 3 {
			require.NotNil(t, result, "expected trick result after 4th card")
			require.Equal(t, 3, result.WinnerSeat, "expected seat 3 (5C) to win")
			require.Equal(t, Points(0), result.Points, "expected 0 points (all clubs)")
		}
	}

	require.Equal(t, PhaseComplete, r.Phase(), "expected PhaseComplete after last trick")
}

func TestRoundScores(t *testing.T) {
	r := NewTestRound([PlayersPerTable]Points{5, 7, 3, 11})
	scores := r.Scores()

	require.Equal(t, Points(5), scores.Raw[0])
	require.Equal(t, Points(11), scores.Raw[3])
	require.Equal(t, Points(5), scores.Adjusted[0], "no shoot the moon")
}

func TestRoundScoresShootTheMoon(t *testing.T) {
	r := NewTestRound([PlayersPerTable]Points{26, 0, 0, 0})
	scores := r.Scores()

	require.Equal(t, Points(0), scores.Adjusted[0])
	require.Equal(t, Points(26), scores.Adjusted[1])
}

func dealTestHands() [PlayersPerTable][]Card {
	return [PlayersPerTable][]Card{
		{{Suit: SuitClubs, Rank: 2}},
		{{Suit: SuitClubs, Rank: 3}},
		{{Suit: SuitClubs, Rank: 4}},
		{{Suit: SuitClubs, Rank: 5}},
	}
}
