package table

import (
	"testing"

	"github.com/JHK/hearts/internal/game"
)

func TestLeaveRemovesPlayerBeforeRoundStart(t *testing.T) {
	runtime := NewRuntime("leave-before-start")
	defer runtime.Close()

	alice, err := runtime.Join("Alice", "alice-token")
	if err != nil {
		t.Fatalf("join alice: %v", err)
	}
	bob, err := runtime.Join("Bob", "bob-token")
	if err != nil {
		t.Fatalf("join bob: %v", err)
	}

	runtime.Leave(alice.PlayerID)

	snapshot := runtime.Snapshot("")
	if len(snapshot.Players) != 1 {
		t.Fatalf("expected one player after leave, got %d", len(snapshot.Players))
	}
	if snapshot.Players[0].PlayerID != bob.PlayerID {
		t.Fatalf("expected bob to remain, got %s", snapshot.Players[0].PlayerID)
	}
	if snapshot.Players[0].Seat != 0 {
		t.Fatalf("expected remaining player to be reseated to 0, got %d", snapshot.Players[0].Seat)
	}
}

func TestLeaveConvertsHumanToBotDuringRound(t *testing.T) {
	runtime := NewRuntime("leave-during-round")
	defer runtime.Close()

	alice, err := runtime.Join("Alice", "alice-token")
	if err != nil {
		t.Fatalf("join alice: %v", err)
	}
	bob, err := runtime.Join("Bob", "bob-token")
	if err != nil {
		t.Fatalf("join bob: %v", err)
	}
	if _, err := runtime.AddBot(""); err != nil {
		t.Fatalf("add bot 1: %v", err)
	}
	if _, err := runtime.AddBot(""); err != nil {
		t.Fatalf("add bot 2: %v", err)
	}

	start := runtime.Start(alice.PlayerID)
	if !start.Accepted {
		t.Fatalf("expected start accepted, got %s", start.Reason)
	}

	runtime.Leave(bob.PlayerID)

	snapshot := runtime.Snapshot("")
	if len(snapshot.Players) != 4 {
		t.Fatalf("expected 4 players after converting to bot, got %d", len(snapshot.Players))
	}

	for _, player := range snapshot.Players {
		if player.PlayerID != bob.PlayerID {
			continue
		}
		if !player.IsBot {
			t.Fatalf("expected leaving player to become bot")
		}
		return
	}

	t.Fatalf("expected to find leaving player in snapshot")
}

func TestStartBeginsInPassingPhaseAndBlocksPlay(t *testing.T) {
	runtime := NewRuntime("passing-phase")
	defer runtime.Close()

	alice, err := runtime.Join("Alice", "alice-token")
	if err != nil {
		t.Fatalf("join alice: %v", err)
	}
	bob, err := runtime.Join("Bob", "bob-token")
	if err != nil {
		t.Fatalf("join bob: %v", err)
	}
	carol, err := runtime.Join("Carol", "carol-token")
	if err != nil {
		t.Fatalf("join carol: %v", err)
	}
	if _, err := runtime.Join("Dave", "dave-token"); err != nil {
		t.Fatalf("join dave: %v", err)
	}

	start := runtime.Start(alice.PlayerID)
	if !start.Accepted {
		t.Fatalf("expected start accepted, got %s", start.Reason)
	}

	snapshot := runtime.Snapshot(alice.PlayerID)
	if snapshot.Phase != "passing" {
		t.Fatalf("expected passing phase, got %q", snapshot.Phase)
	}

	play := runtime.Play(bob.PlayerID, snapshot.Hand[0])
	if play.Accepted {
		t.Fatalf("expected play to be rejected during passing phase")
	}

	play = runtime.Play(carol.PlayerID, snapshot.Hand[0])
	if play.Accepted {
		t.Fatalf("expected play to be rejected during passing phase")
	}
}

func TestPassingAndReviewFlowTransitionsToPlay(t *testing.T) {
	runtime := NewRuntime("pass-review")
	defer runtime.Close()

	players := make([]game.PlayerID, 0, 4)
	for _, entry := range []struct {
		name  string
		token string
	}{{"Alice", "alice-token"}, {"Bob", "bob-token"}, {"Carol", "carol-token"}, {"Dave", "dave-token"}} {
		join, err := runtime.Join(entry.name, entry.token)
		if err != nil {
			t.Fatalf("join %s: %v", entry.name, err)
		}
		players = append(players, join.PlayerID)
	}

	start := runtime.Start(players[0])
	if !start.Accepted {
		t.Fatalf("expected start accepted, got %s", start.Reason)
	}

	passChoices := make(map[game.PlayerID][]string, len(players))
	for _, playerID := range players {
		hand := runtime.Snapshot(playerID).Hand
		if len(hand) != 13 {
			t.Fatalf("expected full hand before passing, got %d", len(hand))
		}
		passChoices[playerID] = append([]string(nil), hand[:3]...)
	}

	for _, playerID := range players {
		pass := runtime.Pass(playerID, passChoices[playerID])
		if !pass.Accepted {
			t.Fatalf("expected pass accepted for %s, got %s", playerID, pass.Reason)
		}
	}

	for _, playerID := range players {
		snapshot := runtime.Snapshot(playerID)
		if snapshot.Phase != "pass_review" {
			t.Fatalf("expected pass review phase for %s, got %q", playerID, snapshot.Phase)
		}
		if len(snapshot.Hand) != 13 {
			t.Fatalf("expected 13 cards after passing for %s, got %d", playerID, len(snapshot.Hand))
		}
		if len(snapshot.PassSent) != 3 {
			t.Fatalf("expected 3 sent cards for %s, got %d", playerID, len(snapshot.PassSent))
		}
		if len(snapshot.PassReceived) != 3 {
			t.Fatalf("expected 3 received cards for %s, got %d", playerID, len(snapshot.PassReceived))
		}
	}

	for i := 0; i < len(players)-1; i++ {
		ready := runtime.ReadyAfterPass(players[i])
		if !ready.Accepted {
			t.Fatalf("expected ready accepted for %s, got %s", players[i], ready.Reason)
		}
	}

	if phase := runtime.Snapshot(players[0]).Phase; phase != "pass_review" {
		t.Fatalf("expected pass review until all ready, got %q", phase)
	}

	ready := runtime.ReadyAfterPass(players[len(players)-1])
	if !ready.Accepted {
		t.Fatalf("expected last ready accepted, got %s", ready.Reason)
	}

	playing := runtime.Snapshot(players[0])
	if playing.Phase != "playing" {
		t.Fatalf("expected play phase after all ready, got %q", playing.Phase)
	}
	if playing.TurnPlayerID == "" {
		t.Fatalf("expected turn player after transitioning to play")
	}
}

func TestPassDirectionCycle(t *testing.T) {
	tests := []struct {
		roundIndex int
		expected   string
	}{
		{roundIndex: 0, expected: passDirectionLeft},
		{roundIndex: 1, expected: passDirectionRight},
		{roundIndex: 2, expected: passDirectionAcross},
		{roundIndex: 3, expected: passDirectionHold},
		{roundIndex: 4, expected: passDirectionLeft},
	}

	for _, tc := range tests {
		if actual := passDirectionForRound(tc.roundIndex); actual != tc.expected {
			t.Fatalf("round %d: expected %q, got %q", tc.roundIndex, tc.expected, actual)
		}
	}
}

func TestCompleteRoundUpdatesTotalsAndHistory(t *testing.T) {
	runtime := &Runtime{}
	state := &tableState{
		totals: map[game.PlayerID]game.Points{
			"p1": 10,
			"p2": 15,
			"p3": 2,
			"p4": 8,
		},
		round: &roundState{
			RoundPoints: map[game.PlayerID]game.Points{
				"p1": 5,
				"p2": 7,
				"p3": 3,
				"p4": 11,
			},
		},
	}

	completed := runtime.completeRound(state)

	if len(state.roundHistory) != 1 {
		t.Fatalf("expected one round in history, got %d", len(state.roundHistory))
	}

	historyEntry := state.roundHistory[0]
	if historyEntry["p1"] != 5 || historyEntry["p2"] != 7 || historyEntry["p3"] != 3 || historyEntry["p4"] != 11 {
		t.Fatalf("unexpected history entry: %#v", historyEntry)
	}

	if state.totals["p1"] != 15 || state.totals["p2"] != 22 || state.totals["p3"] != 5 || state.totals["p4"] != 19 {
		t.Fatalf("unexpected totals: %#v", state.totals)
	}

	if completed.RoundPoints["p1"] != 5 || completed.RoundPoints["p2"] != 7 || completed.RoundPoints["p3"] != 3 || completed.RoundPoints["p4"] != 11 {
		t.Fatalf("unexpected round completed points: %#v", completed.RoundPoints)
	}

	if completed.TotalPoints["p1"] != 15 || completed.TotalPoints["p2"] != 22 || completed.TotalPoints["p3"] != 5 || completed.TotalPoints["p4"] != 19 {
		t.Fatalf("unexpected completed totals: %#v", completed.TotalPoints)
	}
}

func TestBuildSnapshotCopiesRoundHistory(t *testing.T) {
	runtime := &Runtime{tableID: "history-copy"}
	state := &tableState{
		roundHistory: []map[game.PlayerID]game.Points{
			{"p1": 1, "p2": 2, "p3": 3, "p4": 4},
		},
		totals: map[game.PlayerID]game.Points{"p1": 1},
	}

	snapshot := runtime.buildSnapshot(state, "")
	if len(snapshot.RoundHistory) != 1 {
		t.Fatalf("expected one round history entry, got %d", len(snapshot.RoundHistory))
	}
	if snapshot.RoundHistory[0]["p3"] != 3 {
		t.Fatalf("unexpected snapshot round history: %#v", snapshot.RoundHistory)
	}

	state.roundHistory[0]["p3"] = 99
	if snapshot.RoundHistory[0]["p3"] != 3 {
		t.Fatalf("expected snapshot round history to be copied, got %#v", snapshot.RoundHistory)
	}
}
