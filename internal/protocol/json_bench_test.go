package protocol_test

import (
	"encoding/json"
	"testing"

	"github.com/JHK/hearts/internal/game"
	"github.com/JHK/hearts/internal/protocol"
)

// wsMessage mirrors the server→client envelope in internal/webui/ws.go.
type wsMessage struct {
	Type  string `json:"type"`
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

// wsCommand mirrors the client→server envelope in internal/webui/ws.go.
type wsCommand struct {
	Type     string   `json:"type"`
	Name     string   `json:"name,omitempty"`
	Token    string   `json:"token,omitempty"`
	Card     string   `json:"card,omitempty"`
	Cards    []string `json:"cards,omitempty"`
	Strategy string   `json:"strategy,omitempty"`
	Seat     *int     `json:"seat,omitempty"`
	TableID  string   `json:"table_id,omitempty"`
}

// snapshot mirrors session.Snapshot (the largest JSON payload).
type snapshot struct {
	TableID            string                                    `json:"table_id"`
	Players            []protocol.PlayerInfo                     `json:"players"`
	Started            bool                                      `json:"started"`
	Phase              string                                    `json:"phase"`
	TrickNumber        int                                       `json:"trick_number"`
	TurnPlayerID       protocol.PlayerID                         `json:"turn_player_id"`
	HeartsBroken       bool                                      `json:"hearts_broken"`
	CurrentTrick       []string                                  `json:"current_trick"`
	Hand               []string                                  `json:"hand"`
	HandSizes          map[protocol.PlayerID]int                 `json:"hand_sizes"`
	PassDirection      game.PassDirection                        `json:"pass_direction"`
	PassSubmitted      bool                                      `json:"pass_submitted"`
	PassSubmittedCount int                                       `json:"pass_submitted_count"`
	RoundPoints        map[protocol.PlayerID]game.Points         `json:"round_points"`
	RoundHistory       []map[protocol.PlayerID]game.Points       `json:"round_history"`
	TotalPoints        map[protocol.PlayerID]game.Points         `json:"total_points"`
	GameOver           bool                                      `json:"game_over"`
	Winners            []protocol.PlayerID                       `json:"winners,omitempty"`
	RematchVotes       int                                       `json:"rematch_votes,omitempty"`
	RematchTotal       int                                       `json:"rematch_total,omitempty"`
	Paused             bool                                      `json:"paused,omitempty"`
	PausedForPlayerID  protocol.PlayerID                         `json:"paused_for_player_id,omitempty"`
}

// --- fixtures ---

func fixtureWsCommand() wsCommand {
	return wsCommand{
		Type:  "play",
		Card:  "QS",
		Token: "tok_abc123",
	}
}

func fixtureCardPlayed() wsMessage {
	return wsMessage{
		Type: "card_played",
		Data: protocol.CardPlayedData{
			PlayerID:     "p1",
			Card:         "QS",
			BreaksHearts: false,
		},
	}
}

func fixtureTrickCompleted() wsMessage {
	return wsMessage{
		Type: "trick_completed",
		Data: protocol.TrickCompletedData{
			TrickNumber:    3,
			WinnerPlayerID: "p2",
			Points:         13,
		},
	}
}

func fixtureRoundCompleted() wsMessage {
	return wsMessage{
		Type: "round_completed",
		Data: protocol.RoundCompletedData{
			RoundPoints: map[protocol.PlayerID]game.Points{
				"p1": 0, "p2": 13, "p3": 10, "p4": 3,
			},
			TotalPoints: map[protocol.PlayerID]game.Points{
				"p1": 15, "p2": 40, "p3": 32, "p4": 19,
			},
		},
	}
}

func fixtureSnapshot() wsMessage {
	return wsMessage{
		Type: "table_state",
		Data: snapshot{
			TableID: "tbl_xyz",
			Players: []protocol.PlayerInfo{
				{PlayerID: "p1", Name: "Alice", Seat: 0},
				{PlayerID: "p2", Name: "Bob", Seat: 1},
				{PlayerID: "p3", Name: "Carol", Seat: 2},
				{PlayerID: "p4", Name: "Dave", Seat: 3},
			},
			Started:      true,
			Phase:        "playing",
			TrickNumber:  5,
			TurnPlayerID: "p2",
			HeartsBroken: true,
			CurrentTrick: []string{"5H", "AH"},
			Hand:         []string{"2C", "3C", "4D", "7S", "JH", "KH", "AS"},
			HandSizes: map[protocol.PlayerID]int{
				"p1": 7, "p2": 7, "p3": 7, "p4": 7,
			},
			RoundPoints: map[protocol.PlayerID]game.Points{
				"p1": 3, "p2": 0, "p3": 10, "p4": 0,
			},
			RoundHistory: []map[protocol.PlayerID]game.Points{
				{"p1": 0, "p2": 13, "p3": 10, "p4": 3},
				{"p1": 5, "p2": 8, "p3": 6, "p4": 7},
			},
			TotalPoints: map[protocol.PlayerID]game.Points{
				"p1": 8, "p2": 21, "p3": 26, "p4": 10,
			},
		},
	}
}

// --- Marshal benchmarks ---

func BenchmarkMarshal_WsCommand(b *testing.B) {
	v := fixtureWsCommand()
	b.ReportAllocs()
	for b.Loop() {
		if _, err := json.Marshal(&v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_CardPlayed(b *testing.B) {
	v := fixtureCardPlayed()
	b.ReportAllocs()
	for b.Loop() {
		if _, err := json.Marshal(&v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_TrickCompleted(b *testing.B) {
	v := fixtureTrickCompleted()
	b.ReportAllocs()
	for b.Loop() {
		if _, err := json.Marshal(&v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_RoundCompleted(b *testing.B) {
	v := fixtureRoundCompleted()
	b.ReportAllocs()
	for b.Loop() {
		if _, err := json.Marshal(&v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMarshal_Snapshot(b *testing.B) {
	v := fixtureSnapshot()
	b.ReportAllocs()
	for b.Loop() {
		if _, err := json.Marshal(&v); err != nil {
			b.Fatal(err)
		}
	}
}

// --- Unmarshal benchmarks ---

func BenchmarkUnmarshal_WsCommand(b *testing.B) {
	data, _ := json.Marshal(fixtureWsCommand())
	b.ReportAllocs()
	for b.Loop() {
		var v wsCommand
		if err := json.Unmarshal(data, &v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal_CardPlayed(b *testing.B) {
	// Unmarshal into the concrete type (mirrors what tests do after stripping envelope).
	inner, _ := json.Marshal(protocol.CardPlayedData{
		PlayerID: "p1", Card: "QS", BreaksHearts: false,
	})
	b.ReportAllocs()
	for b.Loop() {
		var v protocol.CardPlayedData
		if err := json.Unmarshal(inner, &v); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal_Snapshot(b *testing.B) {
	data, _ := json.Marshal(fixtureSnapshot().Data)
	b.ReportAllocs()
	for b.Loop() {
		var v snapshot
		if err := json.Unmarshal(data, &v); err != nil {
			b.Fatal(err)
		}
	}
}
