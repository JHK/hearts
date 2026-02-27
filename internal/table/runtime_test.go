package table

import "testing"

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
