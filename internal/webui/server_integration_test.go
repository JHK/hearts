package webui

import (
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

func TestServesExtractedScripts(t *testing.T) {
	manager := table.NewManager()
	defer manager.Close()

	handler, err := NewHandler(manager)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	srv := httptest.NewServer(handler)
	defer srv.Close()

	for _, path := range []string{
		"/assets/js/lobby/main.js",
		"/assets/js/table/main.js",
		"/assets/js/table/dom.js",
		"/assets/js/table/render.js",
		"/assets/js/table/cards.js",
	} {
		resp, err := srv.Client().Get(srv.URL + path)
		if err != nil {
			t.Fatalf("get %s: %v", path, err)
		}
		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			t.Fatalf("expected status 200 for %s, got %d", path, resp.StatusCode)
		}
		if contentType := resp.Header.Get("Content-Type"); !strings.HasPrefix(contentType, "text/javascript") {
			_ = resp.Body.Close()
			t.Fatalf("expected JavaScript content type for %s, got %q", path, contentType)
		}
		_ = resp.Body.Close()
	}

	missing, err := srv.Client().Get(srv.URL + "/assets/js/unknown/main.js")
	if err != nil {
		t.Fatalf("get missing script: %v", err)
	}
	defer missing.Body.Close()
	if missing.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown script, got %d", missing.StatusCode)
	}
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

	first := mustDialTableSocket(t, srv.URL, "rejoin")
	defer first.Close()
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
	if err := first.Close(); err != nil {
		t.Fatalf("close first websocket: %v", err)
	}

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

func TestDisconnectLeavesTableBeforeRoundStart(t *testing.T) {
	manager := table.NewManager()
	defer manager.Close()

	handler, err := NewHandler(manager)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	srv := httptest.NewServer(handler)
	defer srv.Close()

	alice := mustDialTableSocket(t, srv.URL, "leave-before-start")
	defer alice.Close()
	_ = readMessage(t, alice)
	if err := alice.WriteJSON(wsCommand{Type: "join", Name: "Alice", Token: "alice-token"}); err != nil {
		t.Fatalf("alice join write: %v", err)
	}
	_ = readMessageType(t, alice, "join_result")

	bob := mustDialTableSocket(t, srv.URL, "leave-before-start")
	defer bob.Close()
	_ = readMessage(t, bob)
	if err := bob.WriteJSON(wsCommand{Type: "join", Name: "Bob", Token: "bob-token"}); err != nil {
		t.Fatalf("bob join write: %v", err)
	}
	_ = readMessageType(t, bob, "join_result")

	if err := alice.Close(); err != nil {
		t.Fatalf("close alice websocket: %v", err)
	}

	assertEventually(t, 2*time.Second, func() bool {
		snapshot := requestStateSnapshot(t, bob)
		if len(snapshot.Players) != 1 {
			return false
		}
		return snapshot.Players[0].Name == "Bob"
	})
}

func TestWebSocketAutoCreatesTable(t *testing.T) {
	manager := table.NewManager()
	defer manager.Close()

	handler, err := NewHandler(manager)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	srv := httptest.NewServer(handler)
	defer srv.Close()

	ws := mustDialTableSocket(t, srv.URL, "auto-create")
	defer ws.Close()

	_ = readMessage(t, ws)

	if _, ok := manager.Get("auto-create"); !ok {
		t.Fatalf("expected table to be created when websocket connected")
	}
}

func TestTableClosesWhenLastHumanLeaves(t *testing.T) {
	manager := table.NewManager()
	defer manager.Close()

	handler, err := NewHandler(manager)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	srv := httptest.NewServer(handler)
	defer srv.Close()

	ws := mustDialTableSocket(t, srv.URL, "close-on-empty")
	_ = readMessage(t, ws)
	if err := ws.WriteJSON(wsCommand{Type: "join", Name: "Alice", Token: "alice-token"}); err != nil {
		t.Fatalf("join write: %v", err)
	}
	_ = readMessageType(t, ws, "join_result")
	if err := ws.Close(); err != nil {
		t.Fatalf("close websocket: %v", err)
	}

	assertEventually(t, 2*time.Second, func() bool {
		_, ok := manager.Get("close-on-empty")
		return !ok
	})
}

func TestPassPhaseAndReviewFlowOverWebSocket(t *testing.T) {
	manager := table.NewManager()
	defer manager.Close()

	handler, err := NewHandler(manager)
	if err != nil {
		t.Fatalf("new handler: %v", err)
	}

	srv := httptest.NewServer(handler)
	defer srv.Close()

	players := []struct {
		name  string
		token string
		conn  *websocket.Conn
	}{
		{name: "Alice", token: "token-a"},
		{name: "Bob", token: "token-b"},
		{name: "Carol", token: "token-c"},
		{name: "Dave", token: "token-d"},
	}

	for i := range players {
		players[i].conn = mustDialTableSocket(t, srv.URL, "pass-flow")
		defer players[i].conn.Close()
		_ = readMessage(t, players[i].conn)
		if err := players[i].conn.WriteJSON(wsCommand{Type: "join", Name: players[i].name, Token: players[i].token}); err != nil {
			t.Fatalf("join write for %s: %v", players[i].name, err)
		}
		join := readMessageType(t, players[i].conn, "join_result")
		var joinResp protocol.JoinResponse
		if err := json.Unmarshal(join.Data, &joinResp); err != nil {
			t.Fatalf("decode join response for %s: %v", players[i].name, err)
		}
		if !joinResp.Accepted {
			t.Fatalf("join rejected for %s: %s", players[i].name, joinResp.Reason)
		}
	}

	if err := players[0].conn.WriteJSON(wsCommand{Type: "start"}); err != nil {
		t.Fatalf("start write: %v", err)
	}
	startMsg := readMessageType(t, players[0].conn, "start_result")
	var startResp protocol.CommandResponse
	if err := json.Unmarshal(startMsg.Data, &startResp); err != nil {
		t.Fatalf("decode start response: %v", err)
	}
	if !startResp.Accepted {
		t.Fatalf("expected start accepted, got %s", startResp.Reason)
	}

	passCardsByPlayer := make([][]string, len(players))
	for i := range players {
		var snapshot table.Snapshot
		assertEventually(t, 2*time.Second, func() bool {
			snapshot = requestStateSnapshot(t, players[i].conn)
			return snapshot.Phase == "passing" && len(snapshot.Hand) == 13
		})
		passCardsByPlayer[i] = append([]string(nil), snapshot.Hand[:3]...)
	}

	for i := range players {
		if err := players[i].conn.WriteJSON(wsCommand{Type: "pass", Cards: passCardsByPlayer[i]}); err != nil {
			t.Fatalf("pass write for %s: %v", players[i].name, err)
		}
		passMsg := readMessageType(t, players[i].conn, "pass_result")
		var passResp protocol.CommandResponse
		if err := json.Unmarshal(passMsg.Data, &passResp); err != nil {
			t.Fatalf("decode pass response for %s: %v", players[i].name, err)
		}
		if !passResp.Accepted {
			t.Fatalf("pass rejected for %s: %s", players[i].name, passResp.Reason)
		}
	}

	assertEventually(t, 2*time.Second, func() bool {
		snapshot := requestStateSnapshot(t, players[0].conn)
		return snapshot.Phase == "pass_review" && len(snapshot.PassReceived) == 3
	})

	for i := 0; i < len(players)-1; i++ {
		if err := players[i].conn.WriteJSON(wsCommand{Type: "ready_after_pass"}); err != nil {
			t.Fatalf("ready write for %s: %v", players[i].name, err)
		}
		readyMsg := readMessageType(t, players[i].conn, "ready_after_pass_result")
		var readyResp protocol.CommandResponse
		if err := json.Unmarshal(readyMsg.Data, &readyResp); err != nil {
			t.Fatalf("decode ready response for %s: %v", players[i].name, err)
		}
		if !readyResp.Accepted {
			t.Fatalf("ready rejected for %s: %s", players[i].name, readyResp.Reason)
		}
	}

	if phase := requestStateSnapshot(t, players[0].conn).Phase; phase != "pass_review" {
		t.Fatalf("expected still pass_review before last ready, got %q", phase)
	}

	if err := players[len(players)-1].conn.WriteJSON(wsCommand{Type: "ready_after_pass"}); err != nil {
		t.Fatalf("last ready write: %v", err)
	}
	lastReady := readMessageType(t, players[len(players)-1].conn, "ready_after_pass_result")
	var lastReadyResp protocol.CommandResponse
	if err := json.Unmarshal(lastReady.Data, &lastReadyResp); err != nil {
		t.Fatalf("decode last ready response: %v", err)
	}
	if !lastReadyResp.Accepted {
		t.Fatalf("last ready rejected: %s", lastReadyResp.Reason)
	}

	assertEventually(t, 2*time.Second, func() bool {
		snapshot := requestStateSnapshot(t, players[0].conn)
		return snapshot.Phase == "playing" && snapshot.TurnPlayerID != ""
	})
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

func assertEventually(t *testing.T, timeout time.Duration, predicate func() bool) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if predicate() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}

	t.Fatalf("condition was not met within %s", timeout)
}

func requestStateSnapshot(t *testing.T, conn *websocket.Conn) table.Snapshot {
	t.Helper()

	if err := conn.WriteJSON(wsCommand{Type: "state"}); err != nil {
		t.Fatalf("state write: %v", err)
	}

	stateMsg := readMessageType(t, conn, "table_state")
	var snapshot table.Snapshot
	if err := json.Unmarshal(stateMsg.Data, &snapshot); err != nil {
		t.Fatalf("decode table state: %v", err)
	}

	return snapshot
}
