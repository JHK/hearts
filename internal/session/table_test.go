package session

import (
	"testing"

	"github.com/JHK/hearts/internal/game"
	"github.com/JHK/hearts/internal/protocol"
	"github.com/stretchr/testify/require"
)

func TestLeaveRemovesPlayerBeforeRoundStart(t *testing.T) {
	runtime := NewTable("leave-before-start")
	defer runtime.Close()

	alice, err := runtime.Join("Alice", "alice-token")
	require.NoError(t, err, "join alice")
	bob, err := runtime.Join("Bob", "bob-token")
	require.NoError(t, err, "join bob")

	runtime.Leave(alice.PlayerID)

	snapshot := runtime.Snapshot("")
	require.Len(t, snapshot.Players, 1)
	require.Equal(t, bob.PlayerID, snapshot.Players[0].PlayerID)
	require.Equal(t, 0, snapshot.Players[0].Seat)
}

func TestLeaveConvertsHumanToBotDuringRound(t *testing.T) {
	runtime := NewTable("leave-during-round")
	defer runtime.Close()

	alice, err := runtime.Join("Alice", "alice-token")
	require.NoError(t, err, "join alice")
	bob, err := runtime.Join("Bob", "bob-token")
	require.NoError(t, err, "join bob")
	_, err = runtime.AddBot("")
	require.NoError(t, err, "add bot 1")
	_, err = runtime.AddBot("")
	require.NoError(t, err, "add bot 2")

	start := runtime.Start(alice.PlayerID)
	require.True(t, start.Accepted, "expected start accepted, got %s", start.Reason)

	runtime.Leave(bob.PlayerID)

	snapshot := runtime.Snapshot("")
	require.Len(t, snapshot.Players, 4)

	for _, player := range snapshot.Players {
		if player.PlayerID != bob.PlayerID {
			continue
		}
		require.True(t, player.IsBot, "expected leaving player to become bot")
		return
	}

	t.Fatal("expected to find leaving player in snapshot")
}

func TestStartBeginsInPassingPhaseAndBlocksPlay(t *testing.T) {
	runtime := NewTable("passing-phase")
	defer runtime.Close()

	alice, err := runtime.Join("Alice", "alice-token")
	require.NoError(t, err, "join alice")
	bob, err := runtime.Join("Bob", "bob-token")
	require.NoError(t, err, "join bob")
	carol, err := runtime.Join("Carol", "carol-token")
	require.NoError(t, err, "join carol")
	_, err = runtime.Join("Dave", "dave-token")
	require.NoError(t, err, "join dave")

	start := runtime.Start(alice.PlayerID)
	require.True(t, start.Accepted, "expected start accepted, got %s", start.Reason)

	snapshot := runtime.Snapshot(alice.PlayerID)
	require.Equal(t, "passing", snapshot.Phase)

	play := runtime.Play(bob.PlayerID, snapshot.Hand[0])
	require.False(t, play.Accepted, "expected play to be rejected during passing phase")

	play = runtime.Play(carol.PlayerID, snapshot.Hand[0])
	require.False(t, play.Accepted, "expected play to be rejected during passing phase")
}

func TestPassingAndReviewFlowTransitionsToPlay(t *testing.T) {
	runtime := NewTable("pass-review")
	defer runtime.Close()

	players := make([]protocol.PlayerID, 0, 4)
	for _, entry := range []struct {
		name  string
		token string
	}{{"Alice", "alice-token"}, {"Bob", "bob-token"}, {"Carol", "carol-token"}, {"Dave", "dave-token"}} {
		join, err := runtime.Join(entry.name, entry.token)
		require.NoError(t, err, "join %s", entry.name)
		players = append(players, join.PlayerID)
	}

	start := runtime.Start(players[0])
	require.True(t, start.Accepted, "expected start accepted, got %s", start.Reason)

	passChoices := make(map[protocol.PlayerID][]string, len(players))
	for _, playerID := range players {
		hand := runtime.Snapshot(playerID).Hand
		require.Len(t, hand, 13, "expected full hand before passing")
		passChoices[playerID] = append([]string(nil), hand[:3]...)
	}

	for _, playerID := range players {
		pass := runtime.Pass(playerID, passChoices[playerID])
		require.True(t, pass.Accepted, "expected pass accepted for %s, got %s", playerID, pass.Reason)
	}

	for _, playerID := range players {
		snapshot := runtime.Snapshot(playerID)
		require.Equal(t, "pass_review", snapshot.Phase, "phase for %s", playerID)
		require.Len(t, snapshot.Hand, 13, "cards after passing for %s", playerID)
		require.Len(t, snapshot.PassSent, 3, "sent cards for %s", playerID)
		require.Len(t, snapshot.PassReceived, 3, "received cards for %s", playerID)
	}

	for i := 0; i < len(players)-1; i++ {
		ready := runtime.ReadyAfterPass(players[i])
		require.True(t, ready.Accepted, "expected ready accepted for %s, got %s", players[i], ready.Reason)
	}

	require.Equal(t, "pass_review", runtime.Snapshot(players[0]).Phase, "expected pass review until all ready")

	ready := runtime.ReadyAfterPass(players[len(players)-1])
	require.True(t, ready.Accepted, "expected last ready accepted, got %s", ready.Reason)

	playing := runtime.Snapshot(players[0])
	require.Equal(t, "playing", playing.Phase)
	require.NotEmpty(t, playing.TurnPlayerID, "expected turn player after transitioning to play")
}

func TestPassDirectionCycle(t *testing.T) {
	tests := []struct {
		roundIndex int
		expected   game.PassDirection
	}{
		{roundIndex: 0, expected: game.PassDirectionLeft},
		{roundIndex: 1, expected: game.PassDirectionRight},
		{roundIndex: 2, expected: game.PassDirectionAcross},
		{roundIndex: 3, expected: game.PassDirectionHold},
		{roundIndex: 4, expected: game.PassDirectionLeft},
	}

	for _, tc := range tests {
		require.Equal(t, tc.expected, game.PassDirectionForRound(tc.roundIndex), "round %d", tc.roundIndex)
	}
}

func TestCompleteRoundUpdatesTotalsAndHistory(t *testing.T) {
	runtime := &Table{}
	players := []*playerState{
		{id: "p1", position: 0},
		{id: "p2", position: 1},
		{id: "p3", position: 2},
		{id: "p4", position: 3},
	}
	g := game.NewGame()
	g.AddRoundScores([game.PlayersPerTable]game.Points{10, 15, 2, 8})
	state := &tableState{
		players: players,
		round:   game.NewTestRound([game.PlayersPerTable]game.Points{5, 7, 3, 11}),
		game:    g,
	}

	completed := runtime.completeRound(state)

	require.Len(t, state.roundHistory, 1)

	historyEntry := state.roundHistory[0]
	require.Equal(t, game.Points(5), historyEntry["p1"])
	require.Equal(t, game.Points(7), historyEntry["p2"])
	require.Equal(t, game.Points(3), historyEntry["p3"])
	require.Equal(t, game.Points(11), historyEntry["p4"])

	scores := state.game.Scores()
	require.Equal(t, game.Points(15), scores[0])
	require.Equal(t, game.Points(22), scores[1])
	require.Equal(t, game.Points(5), scores[2])
	require.Equal(t, game.Points(19), scores[3])

	require.Equal(t, game.Points(5), completed.RoundPoints["p1"])
	require.Equal(t, game.Points(7), completed.RoundPoints["p2"])
	require.Equal(t, game.Points(3), completed.RoundPoints["p3"])
	require.Equal(t, game.Points(11), completed.RoundPoints["p4"])

	require.Equal(t, game.Points(15), completed.TotalPoints["p1"])
	require.Equal(t, game.Points(22), completed.TotalPoints["p2"])
	require.Equal(t, game.Points(5), completed.TotalPoints["p3"])
	require.Equal(t, game.Points(19), completed.TotalPoints["p4"])
}

func TestBuildSnapshotCopiesRoundHistory(t *testing.T) {
	runtime := &Table{tableID: "history-copy"}
	g := game.NewGame()
	g.AddRoundScores([game.PlayersPerTable]game.Points{1, 0, 0, 0})
	state := &tableState{
		players: []*playerState{
			{id: "p1", position: 0},
		},
		game: g,
		roundHistory: []map[protocol.PlayerID]game.Points{
			{"p1": 1, "p2": 2, "p3": 3, "p4": 4},
		},
	}

	snapshot := runtime.buildSnapshot(state, "")
	require.Len(t, snapshot.RoundHistory, 1)
	require.Equal(t, game.Points(3), snapshot.RoundHistory[0]["p3"])

	state.roundHistory[0]["p3"] = 99
	require.Equal(t, game.Points(3), snapshot.RoundHistory[0]["p3"], "expected snapshot round history to be copied")
}
