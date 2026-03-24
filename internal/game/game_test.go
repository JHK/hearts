package game

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewGame(t *testing.T) {
	g := NewGame()
	require.Equal(t, 0, g.RoundsPlayed())
	require.False(t, g.IsOver())
}

func TestAddRoundScores(t *testing.T) {
	g := NewGame()

	require.False(t, g.AddRoundScores([PlayersPerTable]Points{10, 5, 8, 3}), "game should not be over after first round")
	require.Equal(t, 1, g.RoundsPlayed())
	require.Equal(t, Points(10), g.Score(0))
	require.Equal(t, Points(5), g.Score(1))

	require.False(t, g.AddRoundScores([PlayersPerTable]Points{20, 0, 2, 4}), "game should not be over after second round")
	require.Equal(t, Points(30), g.Score(0))
}

func TestGameOverThresholdReached(t *testing.T) {
	g := NewGame()
	g.AddRoundScores([PlayersPerTable]Points{90, 10, 20, 30})
	require.False(t, g.IsOver(), "game should not be over at 90 points")

	require.True(t, g.AddRoundScores([PlayersPerTable]Points{10, 5, 5, 6}), "game should be over when a player reaches 100")
}

func TestGameOverExactThreshold(t *testing.T) {
	g := NewGame()
	require.True(t, g.AddRoundScores([PlayersPerTable]Points{100, 0, 0, 0}), "game should be over at exactly 100 points")
}

func TestWinnersSingleWinner(t *testing.T) {
	g := NewGame()
	g.AddRoundScores([PlayersPerTable]Points{26, 0, 0, 0})
	g.AddRoundScores([PlayersPerTable]Points{26, 0, 0, 0})
	g.AddRoundScores([PlayersPerTable]Points{26, 0, 0, 0})
	g.AddRoundScores([PlayersPerTable]Points{26, 10, 5, 8})

	require.True(t, g.IsOver())
	require.Equal(t, []int{2}, g.Winners())
}

func TestWinnersTied(t *testing.T) {
	g := NewGame()
	g.AddRoundScores([PlayersPerTable]Points{100, 50, 20, 20})

	require.True(t, g.IsOver())
	require.Equal(t, []int{2, 3}, g.Winners())
}

func TestResult(t *testing.T) {
	g := NewGame()
	g.AddRoundScores([PlayersPerTable]Points{100, 30, 20, 50})

	result := g.Result()
	require.Equal(t, [PlayersPerTable]Points{100, 30, 20, 50}, result.Scores)
	require.Equal(t, []int{2}, result.Winners)
}

func TestNextPassDirection(t *testing.T) {
	g := NewGame()
	expected := []PassDirection{PassDirectionLeft, PassDirectionRight, PassDirectionAcross, PassDirectionHold, PassDirectionLeft}
	for i, want := range expected {
		require.Equal(t, want, g.NextPassDirection(), "round %d", i)
		g.AddRoundScores([PlayersPerTable]Points{0, 0, 0, 0})
	}
}
