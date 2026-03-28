package session

import (
	"testing"

	"github.com/JHK/hearts/internal/game"
	"github.com/JHK/hearts/internal/game/bot"
	"github.com/JHK/hearts/internal/protocol"
	"github.com/stretchr/testify/require"
)

func TestLeaveRemovesPlayerBeforeRoundStart(t *testing.T) {
	runtime := NewTable("leave-before-start", nil)
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

func TestLeavePausesGameDuringRound(t *testing.T) {
	runtime := NewTable("leave-during-round", nil)
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
	require.Len(t, snapshot.Players, 4, "all players remain seated")
	require.True(t, snapshot.Paused, "game should be paused after disconnect")

	for _, player := range snapshot.Players {
		if player.PlayerID == bob.PlayerID {
			require.True(t, player.IsBot, "disconnected player should be converted to bot")
		}
	}

	// Game commands should be rejected while paused.
	aliceSnap := runtime.Snapshot(alice.PlayerID)
	if len(aliceSnap.Hand) > 0 {
		play := runtime.Play(alice.PlayerID, aliceSnap.Hand[0])
		require.False(t, play.Accepted, "play should be rejected while paused")
		require.Equal(t, "game is paused", play.Reason)
	}
}

func TestResumeGameAfterDisconnect(t *testing.T) {
	runtime := NewTable("resume-game", nil)
	defer runtime.Close()

	alice, err := runtime.Join("Alice", "alice-token")
	require.NoError(t, err, "join alice")
	_, err = runtime.Join("Bob", "bob-token")
	require.NoError(t, err, "join bob")
	_, err = runtime.AddBot("")
	require.NoError(t, err, "add bot 1")
	_, err = runtime.AddBot("")
	require.NoError(t, err, "add bot 2")

	start := runtime.Start(alice.PlayerID)
	require.True(t, start.Accepted, "expected start accepted, got %s", start.Reason)

	runtime.Leave(alice.PlayerID)
	require.True(t, runtime.Snapshot("").Paused, "game should be paused")

	rejoin, err := runtime.Join("Alice", "alice-token")
	require.NoError(t, err, "rejoin alice")
	require.True(t, rejoin.Accepted, "rejoin should be accepted")

	resp := runtime.ResumeGame(rejoin.PlayerID)
	// Game already resumed on reconnect, so this should say "not paused".
	require.False(t, resp.Accepted, "game already resumed on reconnect")

	snapshot := runtime.Snapshot("")
	require.False(t, snapshot.Paused, "game should not be paused")
}

func TestResumeGameWithoutReconnect(t *testing.T) {
	runtime := NewTable("resume-no-reconnect", nil)
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
	require.True(t, snapshot.Paused, "game should be paused")
	require.Equal(t, bob.PlayerID, snapshot.PausedForPlayerID)

	resp := runtime.ResumeGame(alice.PlayerID)
	require.True(t, resp.Accepted, "expected resume accepted, got %s", resp.Reason)

	snapshot = runtime.Snapshot("")
	require.False(t, snapshot.Paused, "game should resume")
}

func TestReconnectResumesGame(t *testing.T) {
	runtime := NewTable("reconnect", nil)
	defer runtime.Close()

	alice, err := runtime.Join("Alice", "alice-token")
	require.NoError(t, err, "join alice")
	_, err = runtime.Join("Bob", "bob-token")
	require.NoError(t, err, "join bob")
	_, err = runtime.AddBot("")
	require.NoError(t, err, "add bot 1")
	_, err = runtime.AddBot("")
	require.NoError(t, err, "add bot 2")

	start := runtime.Start(alice.PlayerID)
	require.True(t, start.Accepted, "expected start accepted, got %s", start.Reason)

	runtime.Leave(alice.PlayerID)
	require.True(t, runtime.Snapshot("").Paused, "game should be paused")

	rejoin, err := runtime.Join("Alice", "alice-token")
	require.NoError(t, err, "rejoin alice")
	require.True(t, rejoin.Accepted, "expected rejoin accepted")
	require.Equal(t, alice.PlayerID, rejoin.PlayerID, "should reclaim same player ID")

	snapshot := runtime.Snapshot("")
	require.False(t, snapshot.Paused, "game should resume after reconnection")

	for _, player := range snapshot.Players {
		if player.PlayerID == alice.PlayerID {
			require.False(t, player.IsBot, "reconnected player should not be a bot")
		}
	}
}

func TestMultipleDisconnectsOverwritePausedPlayer(t *testing.T) {
	runtime := NewTable("multi-disconnect", nil)
	defer runtime.Close()

	alice, err := runtime.Join("Alice", "alice-token")
	require.NoError(t, err)
	bob, err := runtime.Join("Bob", "bob-token")
	require.NoError(t, err)
	carol, err := runtime.Join("Carol", "carol-token")
	require.NoError(t, err)
	_, err = runtime.Join("Dave", "dave-token")
	require.NoError(t, err)

	start := runtime.Start(alice.PlayerID)
	require.True(t, start.Accepted, "start: %s", start.Reason)

	runtime.Leave(bob.PlayerID)
	snapshot := runtime.Snapshot("")
	require.True(t, snapshot.Paused)
	require.Equal(t, bob.PlayerID, snapshot.PausedForPlayerID)

	// Second disconnect overwrites the paused-for player but stays paused.
	runtime.Leave(carol.PlayerID)
	snapshot = runtime.Snapshot("")
	require.True(t, snapshot.Paused)
	require.Equal(t, carol.PlayerID, snapshot.PausedForPlayerID)

	// Resume clears pause even though two players disconnected.
	resp := runtime.ResumeGame(alice.PlayerID)
	require.True(t, resp.Accepted, "resume: %s", resp.Reason)
	require.False(t, runtime.Snapshot("").Paused)
}

func TestResumeGameRejectsBot(t *testing.T) {
	runtime := NewTable("resume-rejects-bot", nil)
	defer runtime.Close()

	alice, err := runtime.Join("Alice", "alice-token")
	require.NoError(t, err)
	_, err = runtime.Join("Bob", "bob-token")
	require.NoError(t, err)
	bot1, err := runtime.AddBot("")
	require.NoError(t, err)
	_, err = runtime.AddBot("")
	require.NoError(t, err)

	start := runtime.Start(alice.PlayerID)
	require.True(t, start.Accepted, "start: %s", start.Reason)

	runtime.Leave(alice.PlayerID)
	require.True(t, runtime.Snapshot("").Paused)

	// A bot should not be able to resume.
	resp := runtime.ResumeGame(bot1.JoinResponse.PlayerID)
	require.False(t, resp.Accepted)
	require.Equal(t, "only seated human players can resume", resp.Reason)
}

func TestStartBeginsInPassingPhaseAndBlocksPlay(t *testing.T) {
	runtime := NewTable("passing-phase", nil)
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
	runtime := NewTable("pass-review", nil)
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

func TestRematchResetsGameAfterAllHumansVote(t *testing.T) {
	// Test the rematch handler directly with fabricated game-over state.
	rt := &Table{tableID: "rematch-test"}
	g := game.NewGame()
	g.AddRoundScores([game.PlayersPerTable]game.Points{50, 30, 20, 110})

	alice := &playerState{id: "alice", Name: "Alice", position: 0, Token: "alice-token"}
	bob := &playerState{id: "bob", Name: "Bob", position: 1, Token: "bob-token"}
	bot1 := &playerState{id: "bot1", Name: "Bot 1", position: 2, bot: bot.StrategyRandom.NewBot()}
	bot2 := &playerState{id: "bot2", Name: "Bot 2", position: 3, bot: bot.StrategyRandom.NewBot()}

	state := &tableState{
		players:     []*playerState{alice, bob, bot1, bot2},
		playersByID: map[protocol.PlayerID]*playerState{"alice": alice, "bob": bob, "bot1": bot1, "bot2": bot2},
		playersByToken: map[string]*playerState{
			"alice-token": alice,
			"bob-token":   bob,
		},
		departedTokens: make(map[string]protocol.PlayerID),
		game:           g,
		gameOver:       true,
		roundHistory:   []map[protocol.PlayerID]game.Points{{"alice": 50, "bob": 30, "bot1": 20, "bot2": 110}},
	}

	// Rematch before game over should fail.
	state.gameOver = false
	resp := rt.handleRematch(state, "alice")
	require.False(t, resp.Accepted, "rematch should fail when game not over")
	state.gameOver = true

	// First vote.
	resp = rt.handleRematch(state, "alice")
	require.True(t, resp.Accepted, "alice rematch vote: %s", resp.Reason)
	require.True(t, state.gameOver, "game still over after 1 vote")
	require.Len(t, state.rematchVotes, 1)

	snap := rt.buildSnapshot(state, "alice")
	require.Equal(t, 1, snap.RematchVotes)
	require.Equal(t, 2, snap.RematchTotal)
	require.True(t, snap.RematchVoted)

	bobSnap := rt.buildSnapshot(state, "bob")
	require.False(t, bobSnap.RematchVoted)

	// Second vote triggers rematch.
	resp = rt.handleRematch(state, "bob")
	require.True(t, resp.Accepted, "bob rematch vote: %s", resp.Reason)

	require.False(t, state.gameOver, "game should be reset after rematch")
	require.Len(t, state.players, 4, "table should be full (2 humans + 2 auto-bots)")
	require.Equal(t, "Alice", state.players[0].Name)
	require.Equal(t, "Bob", state.players[1].Name)
	require.Nil(t, state.players[0].bot, "Alice should be human")
	require.Nil(t, state.players[1].bot, "Bob should be human")
	require.NotNil(t, state.players[2].bot, "seat 2 should be bot")
	require.NotNil(t, state.players[3].bot, "seat 3 should be bot")
	for i, p := range state.players {
		require.Equal(t, i, p.position, "position should match index")
	}
	require.Nil(t, state.roundHistory, "history should be cleared")
	require.Nil(t, state.rematchVotes)
	require.False(t, state.game.IsOver(), "new game should not be over")
}

func TestRematchRejectsBot(t *testing.T) {
	rt := &Table{tableID: "rematch-bot"}
	g := game.NewGame()
	g.AddRoundScores([game.PlayersPerTable]game.Points{0, 0, 0, 110})

	bot1 := &playerState{id: "bot1", Name: "Bot 1", position: 0, bot: bot.StrategyRandom.NewBot()}
	state := &tableState{
		players:        []*playerState{bot1},
		playersByID:    map[protocol.PlayerID]*playerState{"bot1": bot1},
		playersByToken: make(map[string]*playerState),
		departedTokens: make(map[string]protocol.PlayerID),
		game:           g,
		gameOver:       true,
	}

	resp := rt.handleRematch(state, "bot1")
	require.False(t, resp.Accepted)
	require.Equal(t, "only seated human players can vote for rematch", resp.Reason)
}

func TestRematchLeaveClearsVote(t *testing.T) {
	rt := &Table{tableID: "rematch-leave"}
	g := game.NewGame()
	g.AddRoundScores([game.PlayersPerTable]game.Points{50, 30, 20, 110})

	alice := &playerState{id: "alice", Name: "Alice", position: 0, Token: "alice-token"}
	bob := &playerState{id: "bob", Name: "Bob", position: 1, Token: "bob-token"}

	state := &tableState{
		players:     []*playerState{alice, bob},
		playersByID: map[protocol.PlayerID]*playerState{"alice": alice, "bob": bob},
		playersByToken: map[string]*playerState{
			"alice-token": alice,
			"bob-token":   bob,
		},
		departedTokens: make(map[string]protocol.PlayerID),
		game:           g,
		gameOver:       true,
		rematchVotes:   map[protocol.PlayerID]bool{"alice": true},
		roundHistory:   []map[protocol.PlayerID]game.Points{{"alice": 50, "bob": 110}},
	}

	require.Len(t, state.rematchVotes, 1)

	// Alice leaves — game-over state should fully reset.
	rt.handleLeave(state, "alice")

	require.Nil(t, state.rematchVotes)
	require.Nil(t, state.roundHistory)
	require.False(t, state.gameOver, "game-over should reset when player leaves")
	require.False(t, state.game.IsOver(), "game scores should reset when player leaves")
}

func TestClaimSeatReplacesBot(t *testing.T) {
	runtime := NewTable("claim-seat", nil)
	defer runtime.Close()

	alice, err := runtime.Join("Alice", "alice-token")
	require.NoError(t, err)
	_, err = runtime.AddBot("")
	require.NoError(t, err)
	_, err = runtime.AddBot("")
	require.NoError(t, err)
	_, err = runtime.AddBot("")
	require.NoError(t, err)

	start := runtime.Start(alice.PlayerID)
	require.True(t, start.Accepted, "start: %s", start.Reason)

	// Observer claims seat 2 (a bot seat).
	claim, err := runtime.ClaimSeat(2, "Observer", "observer-token")
	require.NoError(t, err)
	require.True(t, claim.Accepted, "claim: %s", claim.Reason)
	require.Equal(t, 2, claim.Seat)

	snapshot := runtime.Snapshot(claim.PlayerID)
	require.NotEmpty(t, snapshot.Hand, "claimed player should see their hand")

	for _, p := range snapshot.Players {
		if p.PlayerID == claim.PlayerID {
			require.False(t, p.IsBot, "claimed seat should no longer be a bot")
			require.Equal(t, "Observer", p.Name)
		}
	}
}

func TestClaimSeatRejectsHumanSeat(t *testing.T) {
	runtime := NewTable("claim-human", nil)
	defer runtime.Close()

	alice, err := runtime.Join("Alice", "alice-token")
	require.NoError(t, err)
	_, err = runtime.AddBot("")
	require.NoError(t, err)
	_, err = runtime.AddBot("")
	require.NoError(t, err)
	_, err = runtime.AddBot("")
	require.NoError(t, err)

	start := runtime.Start(alice.PlayerID)
	require.True(t, start.Accepted)

	claim, err := runtime.ClaimSeat(0, "Observer", "observer-token")
	require.NoError(t, err)
	require.False(t, claim.Accepted)
	require.Equal(t, "seat is not bot-controlled", claim.Reason)
}

func TestClaimSeatRejectsAlreadySeated(t *testing.T) {
	runtime := NewTable("claim-already-seated", nil)
	defer runtime.Close()

	_, err := runtime.Join("Alice", "alice-token")
	require.NoError(t, err)
	_, err = runtime.AddBot("")
	require.NoError(t, err)
	_, err = runtime.AddBot("")
	require.NoError(t, err)
	_, err = runtime.AddBot("")
	require.NoError(t, err)

	// Alice tries to claim a bot seat with her existing token.
	claim, err := runtime.ClaimSeat(1, "Alice", "alice-token")
	require.NoError(t, err)
	require.False(t, claim.Accepted)
	require.Equal(t, "already seated", claim.Reason)
}

func TestClaimSeatRaceExactlyOneSucceeds(t *testing.T) {
	runtime := NewTable("claim-race", nil)
	defer runtime.Close()

	_, err := runtime.Join("Alice", "alice-token")
	require.NoError(t, err)
	_, err = runtime.AddBot("")
	require.NoError(t, err)
	_, err = runtime.AddBot("")
	require.NoError(t, err)
	_, err = runtime.AddBot("")
	require.NoError(t, err)

	// Two observers try to claim the same seat.
	claim1, err := runtime.ClaimSeat(1, "Obs1", "obs1-token")
	require.NoError(t, err)
	require.True(t, claim1.Accepted, "first claim should succeed")

	claim2, err := runtime.ClaimSeat(1, "Obs2", "obs2-token")
	require.NoError(t, err)
	require.False(t, claim2.Accepted, "second claim should fail")
	require.Equal(t, "seat is not bot-controlled", claim2.Reason)
}

func TestClaimSeatResumesPausedGame(t *testing.T) {
	runtime := NewTable("claim-resume", nil)
	defer runtime.Close()

	alice, err := runtime.Join("Alice", "alice-token")
	require.NoError(t, err)
	bob, err := runtime.Join("Bob", "bob-token")
	require.NoError(t, err)
	_, err = runtime.AddBot("")
	require.NoError(t, err)
	_, err = runtime.AddBot("")
	require.NoError(t, err)

	start := runtime.Start(alice.PlayerID)
	require.True(t, start.Accepted)

	// Bob disconnects — game pauses and seat becomes bot.
	runtime.Leave(bob.PlayerID)
	snapshot := runtime.Snapshot("")
	require.True(t, snapshot.Paused)
	require.Equal(t, bob.PlayerID, snapshot.PausedForPlayerID)

	// Observer claims Bob's (now bot) seat — should unpause.
	claim, err := runtime.ClaimSeat(bob.Seat, "Observer", "observer-token")
	require.NoError(t, err)
	require.True(t, claim.Accepted)
	require.Equal(t, bob.PlayerID, claim.PlayerID, "should inherit the original player ID")

	snapshot = runtime.Snapshot("")
	require.False(t, snapshot.Paused, "game should resume after observer claims paused seat")
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

func TestRenameSeatedHumanPlayer(t *testing.T) {
	runtime := NewTable("rename-happy", nil)
	defer runtime.Close()

	alice, err := runtime.Join("Alice", "alice-token")
	require.NoError(t, err)

	resp := runtime.Rename(alice.PlayerID, "Alicia")
	require.True(t, resp.Accepted)

	snapshot := runtime.Snapshot("")
	require.Equal(t, "Alicia", snapshot.Players[0].Name)
}

func TestRenameEmptyNameRejected(t *testing.T) {
	runtime := NewTable("rename-empty", nil)
	defer runtime.Close()

	alice, err := runtime.Join("Alice", "alice-token")
	require.NoError(t, err)

	resp := runtime.Rename(alice.PlayerID, "   ")
	require.False(t, resp.Accepted)
	require.Equal(t, "name must not be empty", resp.Reason)

	snapshot := runtime.Snapshot("")
	require.Equal(t, "Alice", snapshot.Players[0].Name)
}

func TestRenameSameNameNoOp(t *testing.T) {
	runtime := NewTable("rename-noop", nil)
	defer runtime.Close()

	alice, err := runtime.Join("Alice", "alice-token")
	require.NoError(t, err)

	resp := runtime.Rename(alice.PlayerID, "Alice")
	require.True(t, resp.Accepted)

	snapshot := runtime.Snapshot("")
	require.Equal(t, "Alice", snapshot.Players[0].Name)
}

func TestRenameBotRejected(t *testing.T) {
	runtime := NewTable("rename-bot", nil)
	defer runtime.Close()

	_, err := runtime.Join("Alice", "alice-token")
	require.NoError(t, err)
	added, err := runtime.AddBot("")
	require.NoError(t, err)

	resp := runtime.Rename(added.JoinResponse.PlayerID, "NewBotName")
	require.False(t, resp.Accepted)
	require.Equal(t, "only seated human players can rename", resp.Reason)
}

func TestRenameUnknownPlayerRejected(t *testing.T) {
	runtime := NewTable("rename-unknown", nil)
	defer runtime.Close()

	resp := runtime.Rename("nonexistent", "Ghost")
	require.False(t, resp.Accepted)
	require.Equal(t, "only seated human players can rename", resp.Reason)
}

func TestDisconnectDuringPassingDoesNotAutoSubmit(t *testing.T) {
	runtime := NewTable("pass-disconnect", nil)
	defer runtime.Close()

	alice, err := runtime.Join("Alice", "alice-token")
	require.NoError(t, err)
	bob, err := runtime.Join("Bob", "bob-token")
	require.NoError(t, err)
	_, err = runtime.AddBot("")
	require.NoError(t, err)
	_, err = runtime.AddBot("")
	require.NoError(t, err)

	start := runtime.Start(alice.PlayerID)
	require.True(t, start.Accepted, "start: %s", start.Reason)

	snap := runtime.Snapshot(bob.PlayerID)
	require.Equal(t, "passing", snap.Phase, "should be in passing phase")
	require.False(t, snap.PassSubmitted, "bob hasn't submitted yet")

	// Bob disconnects during passing — his pass should NOT be auto-submitted.
	runtime.Leave(bob.PlayerID)

	snap = runtime.Snapshot(alice.PlayerID)
	require.True(t, snap.Paused, "game should be paused")

	// Bob's pass specifically must not have been submitted.
	bobSnap := runtime.Snapshot(bob.PlayerID)
	require.False(t, bobSnap.PassSubmitted, "bob's pass should not be auto-submitted on disconnect")

	// Bob reconnects — should still be able to submit their own pass.
	bob2, err := runtime.Join("Bob", "bob-token")
	require.NoError(t, err)
	require.Equal(t, bob.PlayerID, bob2.PlayerID, "same player ID on reconnect")

	snap = runtime.Snapshot(bob2.PlayerID)
	require.False(t, snap.Paused, "game should be unpaused after reconnect")
	require.Equal(t, "passing", snap.Phase, "still in passing phase")
	require.False(t, snap.PassSubmitted, "bob's pass should still be unsubmitted")

	// Bob can submit their own pass.
	passResp := runtime.Pass(bob2.PlayerID, snap.Hand[:3])
	require.True(t, passResp.Accepted, "pass: %s", passResp.Reason)

	snap = runtime.Snapshot(bob2.PlayerID)
	require.True(t, snap.PassSubmitted, "bob's pass should now be submitted")
}

func TestDisconnectDuringPassingBotSubmitsOnResume(t *testing.T) {
	runtime := NewTable("pass-resume-bot", nil)
	defer runtime.Close()

	alice, err := runtime.Join("Alice", "alice-token")
	require.NoError(t, err)
	bob, err := runtime.Join("Bob", "bob-token")
	require.NoError(t, err)
	_, err = runtime.AddBot("")
	require.NoError(t, err)
	_, err = runtime.AddBot("")
	require.NoError(t, err)

	start := runtime.Start(alice.PlayerID)
	require.True(t, start.Accepted, "start: %s", start.Reason)

	snap := runtime.Snapshot(bob.PlayerID)
	require.Equal(t, "passing", snap.Phase)

	// Bob disconnects — pass not auto-submitted.
	runtime.Leave(bob.PlayerID)

	snap = runtime.Snapshot(alice.PlayerID)
	require.True(t, snap.Paused)

	// Bob's pass must not have been auto-submitted.
	bobSnap := runtime.Snapshot(bob.PlayerID)
	require.False(t, bobSnap.PassSubmitted, "bob's pass should not be auto-submitted")

	// Alice resumes without Bob reconnecting — bot should submit on resume.
	resumeResp := runtime.ResumeGame(alice.PlayerID)
	require.True(t, resumeResp.Accepted, "resume: %s", resumeResp.Reason)

	// Submit Alice's pass so we can check all passes complete.
	aliceSnap := runtime.Snapshot(alice.PlayerID)
	passResp := runtime.Pass(alice.PlayerID, aliceSnap.Hand[:3])
	require.True(t, passResp.Accepted, "alice pass: %s", passResp.Reason)

	// After resume, bot passes are scheduled asynchronously. Give the table
	// goroutine a moment to process them, then check all passes are in.
	runtime.Drain()
	snap = runtime.Snapshot(alice.PlayerID)
	require.Equal(t, 4, snap.PassSubmittedCount, "all 4 passes should be submitted after resume")
}
