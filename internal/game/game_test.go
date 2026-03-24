package game

import (
	"testing"
)

func TestNewGame(t *testing.T) {
	g := NewGame()
	if g.RoundsPlayed() != 0 {
		t.Fatalf("expected 0 rounds played, got %d", g.RoundsPlayed())
	}
	if g.IsOver() {
		t.Fatal("new game should not be over")
	}
}

func TestAddRoundScores(t *testing.T) {
	g := NewGame()

	if over := g.AddRoundScores([PlayersPerTable]Points{10, 5, 8, 3}); over {
		t.Fatal("game should not be over after first round")
	}
	if g.RoundsPlayed() != 1 {
		t.Fatalf("expected 1 round played, got %d", g.RoundsPlayed())
	}
	if g.Score(0) != 10 {
		t.Fatalf("expected seat 0 to have 10 points, got %d", g.Score(0))
	}
	if g.Score(1) != 5 {
		t.Fatalf("expected seat 1 to have 5 points, got %d", g.Score(1))
	}

	if over := g.AddRoundScores([PlayersPerTable]Points{20, 0, 2, 4}); over {
		t.Fatal("game should not be over after second round")
	}
	if g.Score(0) != 30 {
		t.Fatalf("expected seat 0 to have 30 points, got %d", g.Score(0))
	}
}

func TestGameOverThresholdReached(t *testing.T) {
	g := NewGame()
	g.AddRoundScores([PlayersPerTable]Points{90, 10, 20, 30})
	if g.IsOver() {
		t.Fatal("game should not be over at 90 points")
	}

	if over := g.AddRoundScores([PlayersPerTable]Points{10, 5, 5, 6}); !over {
		t.Fatal("game should be over when a player reaches 100")
	}
}

func TestGameOverExactThreshold(t *testing.T) {
	g := NewGame()
	if over := g.AddRoundScores([PlayersPerTable]Points{100, 0, 0, 0}); !over {
		t.Fatal("game should be over at exactly 100 points")
	}
}

func TestWinnersSingleWinner(t *testing.T) {
	g := NewGame()
	g.AddRoundScores([PlayersPerTable]Points{26, 0, 0, 0})
	g.AddRoundScores([PlayersPerTable]Points{26, 0, 0, 0})
	g.AddRoundScores([PlayersPerTable]Points{26, 0, 0, 0})
	g.AddRoundScores([PlayersPerTable]Points{26, 10, 5, 8})

	if !g.IsOver() {
		t.Fatal("game should be over")
	}
	winners := g.Winners()
	if len(winners) != 1 || winners[0] != 2 {
		t.Fatalf("expected winner [2], got %v", winners)
	}
}

func TestWinnersTied(t *testing.T) {
	g := NewGame()
	g.AddRoundScores([PlayersPerTable]Points{100, 50, 20, 20})

	if !g.IsOver() {
		t.Fatal("game should be over")
	}
	winners := g.Winners()
	if len(winners) != 2 || winners[0] != 2 || winners[1] != 3 {
		t.Fatalf("expected winners [2, 3], got %v", winners)
	}
}

func TestResult(t *testing.T) {
	g := NewGame()
	g.AddRoundScores([PlayersPerTable]Points{100, 30, 20, 50})

	result := g.Result()
	expected := [PlayersPerTable]Points{100, 30, 20, 50}
	if result.Scores != expected {
		t.Fatalf("expected scores %v, got %v", expected, result.Scores)
	}
	if len(result.Winners) != 1 || result.Winners[0] != 2 {
		t.Fatalf("expected winners [2], got %v", result.Winners)
	}
}

func TestNextPassDirection(t *testing.T) {
	g := NewGame()
	expected := []PassDirection{PassDirectionLeft, PassDirectionRight, PassDirectionAcross, PassDirectionHold, PassDirectionLeft}
	for i, want := range expected {
		if got := g.NextPassDirection(); got != want {
			t.Fatalf("round %d: expected %s, got %s", i, want, got)
		}
		g.AddRoundScores([PlayersPerTable]Points{0, 0, 0, 0})
	}
}
