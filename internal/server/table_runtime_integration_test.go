package server

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JHK/hearts/internal/game"
	"github.com/JHK/hearts/internal/player/bot"
	"github.com/JHK/hearts/internal/protocol"
	natswire "github.com/JHK/hearts/internal/transport/nats"
	"github.com/nats-io/nats.go"
)

func TestRoundCompletesWithFirstLegalBotsAndDeterministicDeck(t *testing.T) {
	local, err := Open("127.0.0.1", 0)
	if err != nil {
		t.Fatalf("open local server: %v", err)
	}
	t.Cleanup(local.Close)

	tableID := "integration"
	deterministicDeck := game.BuildDeck()
	table := newTableRuntimeWithDeck(tableID, local.nc, func() []game.Card {
		return append([]game.Card(nil), deterministicDeck...)
	})
	if err := table.register(); err != nil {
		t.Fatalf("register table: %v", err)
	}
	local.tables[tableID] = table

	observerConn, err := nats.Connect(local.URL())
	if err != nil {
		t.Fatalf("connect observer: %v", err)
	}
	t.Cleanup(observerConn.Close)

	roundCompleted := make(chan protocol.RoundCompletedData, 1)
	decodeErrors := make(chan error, 1)
	var trickCount atomic.Int32

	observer := natswire.NewParticipantClient(observerConn, tableID, "")
	if err := observer.Start(natswire.ParticipantEventHandlers{
		OnTrickCompleted: func(protocol.TrickCompletedData) {
			trickCount.Add(1)
		},
		OnRoundCompleted: func(data protocol.RoundCompletedData) {
			select {
			case roundCompleted <- data:
			default:
			}
		},
		OnDecodeError: func(err error) {
			select {
			case decodeErrors <- err:
			default:
			}
		},
	}); err != nil {
		t.Fatalf("start observer: %v", err)
	}
	t.Cleanup(observer.Stop)

	type botAgent struct {
		conn     *nats.Conn
		runtime  *bot.Runtime
		playerID game.PlayerID
	}

	agents := make([]botAgent, 0, game.PlayersPerTable)
	for i := 0; i < game.PlayersPerTable; i++ {
		conn, err := nats.Connect(local.URL())
		if err != nil {
			t.Fatalf("connect bot %d: %v", i+1, err)
		}

		joinClient := natswire.NewParticipantClient(conn, tableID, "")
		joinResp, err := joinClient.Join(fmt.Sprintf("bot-%d", i+1))
		if err != nil {
			conn.Close()
			t.Fatalf("join bot %d: %v", i+1, err)
		}
		if !joinResp.Accepted {
			conn.Close()
			t.Fatalf("join bot %d rejected: %s", i+1, joinResp.Reason)
		}

		runtime := bot.NewRuntime(conn, tableID, joinResp.PlayerID, bot.NewFirstLegalBot())
		if err := runtime.Start(); err != nil {
			conn.Close()
			t.Fatalf("start bot %d runtime: %v", i+1, err)
		}

		agents = append(agents, botAgent{conn: conn, runtime: runtime, playerID: joinResp.PlayerID})
	}
	t.Cleanup(func() {
		for _, agent := range agents {
			agent.runtime.Stop()
			agent.conn.Close()
		}
	})

	starter := natswire.NewParticipantClient(agents[0].conn, tableID, agents[0].playerID)
	if err := starter.StartRound(); err != nil {
		t.Fatalf("start round: %v", err)
	}

	select {
	case err := <-decodeErrors:
		t.Fatalf("decode error while observing events: %v", err)
	case <-time.After(8 * time.Second):
		t.Fatalf("timed out waiting for round completion")
	case round := <-roundCompleted:
		if trickCount.Load() != 13 {
			t.Fatalf("expected 13 completed tricks, got %d", trickCount.Load())
		}
		if len(round.RoundPoints) != game.PlayersPerTable {
			t.Fatalf("expected round points for %d players, got %d", game.PlayersPerTable, len(round.RoundPoints))
		}
		if len(round.TotalPoints) != game.PlayersPerTable {
			t.Fatalf("expected total points for %d players, got %d", game.PlayersPerTable, len(round.TotalPoints))
		}
	}
}
