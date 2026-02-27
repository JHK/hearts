package webui

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JHK/hearts/internal/protocol"
	"github.com/JHK/hearts/internal/table"
	"github.com/gorilla/websocket"
)

type testWSMessage struct {
	Type  string          `json:"type"`
	Data  json.RawMessage `json:"data,omitempty"`
	Error string          `json:"error,omitempty"`
}

func TestWebSocketJoinAndStateFlow(t *testing.T) {
	manager := table.NewManager()
	defer manager.Close()

	handler, err := NewHandler(manager)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	srv := httptest.NewServer(handler)
	defer srv.Close()

	createTable(t, srv.URL, "demo")

	ws := mustDialTableSocket(t, srv.URL, "demo")
	defer ws.Close()

	_ = readMessage(t, ws)

	if err := ws.WriteJSON(wsCommand{Type: "join", Name: "Alice", Token: "token-a"}); err != nil {
		t.Fatalf("join write: %v", err)
	}

	joinMsg := readMessageType(t, ws, "join_result")
	var joinResp protocol.JoinResponse
	if err := json.Unmarshal(joinMsg.Data, &joinResp); err != nil {
		t.Fatalf("decode join result: %v", err)
	}
	if !joinResp.Accepted {
		t.Fatalf("expected join accepted, got rejected: %s", joinResp.Reason)
	}
	if joinResp.PlayerID == "" {
		t.Fatalf("expected player id")
	}

	if err := ws.WriteJSON(wsCommand{Type: "state"}); err != nil {
		t.Fatalf("state write: %v", err)
	}

	stateMsg := readMessageType(t, ws, "table_state")
	var snapshot table.Snapshot
	if err := json.Unmarshal(stateMsg.Data, &snapshot); err != nil {
		t.Fatalf("decode table state: %v", err)
	}
	if snapshot.TableID != "demo" {
		t.Fatalf("unexpected table id: %s", snapshot.TableID)
	}
	if len(snapshot.Players) != 1 {
		t.Fatalf("expected 1 player, got %d", len(snapshot.Players))
	}

	if err := ws.WriteJSON(wsCommand{Type: "start"}); err != nil {
		t.Fatalf("start write: %v", err)
	}

	startMsg := readMessageType(t, ws, "start_result")
	var startResp protocol.CommandResponse
	if err := json.Unmarshal(startMsg.Data, &startResp); err != nil {
		t.Fatalf("decode start response: %v", err)
	}
	if startResp.Accepted {
		t.Fatalf("expected start rejected with insufficient players")
	}

	if err := ws.WriteJSON(wsCommand{Type: "unknown"}); err != nil {
		t.Fatalf("unknown write: %v", err)
	}

	errMsg := readMessageType(t, ws, "error")
	if strings.TrimSpace(errMsg.Error) == "" {
		t.Fatalf("expected error message for unknown command")
	}
}

func TestWebSocketJoinReusesPlayerByToken(t *testing.T) {
	manager := table.NewManager()
	defer manager.Close()

	handler, err := NewHandler(manager)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	srv := httptest.NewServer(handler)
	defer srv.Close()

	createTable(t, srv.URL, "rejoin")

	first := mustDialTableSocket(t, srv.URL, "rejoin")
	_ = readMessage(t, first)
	if err := first.WriteJSON(wsCommand{Type: "join", Name: "Alice", Token: "stable-token"}); err != nil {
		t.Fatalf("first join write: %v", err)
	}
	firstJoin := readMessageType(t, first, "join_result")
	var firstResp protocol.JoinResponse
	if err := json.Unmarshal(firstJoin.Data, &firstResp); err != nil {
		t.Fatalf("decode first join: %v", err)
	}
	if !firstResp.Accepted {
		t.Fatalf("first join rejected: %s", firstResp.Reason)
	}
	_ = first.Close()

	second := mustDialTableSocket(t, srv.URL, "rejoin")
	defer second.Close()
	_ = readMessage(t, second)
	if err := second.WriteJSON(wsCommand{Type: "join", Name: "Renamed", Token: "stable-token"}); err != nil {
		t.Fatalf("second join write: %v", err)
	}
	secondJoin := readMessageType(t, second, "join_result")
	var secondResp protocol.JoinResponse
	if err := json.Unmarshal(secondJoin.Data, &secondResp); err != nil {
		t.Fatalf("decode second join: %v", err)
	}

	if !secondResp.Accepted {
		t.Fatalf("second join rejected: %s", secondResp.Reason)
	}
	if secondResp.PlayerID != firstResp.PlayerID {
		t.Fatalf("expected same player id on reconnect, got %s and %s", firstResp.PlayerID, secondResp.PlayerID)
	}
	if secondResp.Seat != firstResp.Seat {
		t.Fatalf("expected same seat on reconnect, got %d and %d", firstResp.Seat, secondResp.Seat)
	}
}

func createTable(t *testing.T, baseURL, tableID string) {
	t.Helper()

	body, err := json.Marshal(map[string]string{"table_id": tableID})
	if err != nil {
		t.Fatalf("marshal table create payload: %v", err)
	}

	resp, err := http.Post(baseURL+"/api/tables", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("create table request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected create table status: %d", resp.StatusCode)
	}
}

func mustDialTableSocket(t *testing.T, baseURL, tableID string) *websocket.Conn {
	t.Helper()

	wsURL := "ws" + strings.TrimPrefix(baseURL, "http") + "/ws/table/" + tableID
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	return conn
}

func readMessageType(t *testing.T, conn *websocket.Conn, expected string) testWSMessage {
	t.Helper()

	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		message := readMessage(t, conn)
		if message.Type == expected {
			return message
		}
	}

	t.Fatalf("timed out waiting for websocket message type %q", expected)
	return testWSMessage{}
}

func readMessage(t *testing.T, conn *websocket.Conn) testWSMessage {
	t.Helper()

	if err := conn.SetReadDeadline(time.Now().Add(4 * time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}

	var message testWSMessage
	if err := conn.ReadJSON(&message); err != nil {
		t.Fatalf("read websocket message: %v", err)
	}
	return message
}
