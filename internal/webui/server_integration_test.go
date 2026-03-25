package webui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JHK/hearts/internal/protocol"
	"github.com/JHK/hearts/internal/session"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

type testWSMessage struct {
	Type  string          `json:"type"`
	Data  json.RawMessage `json:"data,omitempty"`
	Error string          `json:"error,omitempty"`
}

func TestImmutableCacheHeadersOnStaticAssets(t *testing.T) {
	manager := session.NewManager()
	defer manager.Close()

	handler, err := NewHandler(Config{}, manager)
	require.NoError(t, err, "new handler")

	srv := httptest.NewServer(handler)
	defer srv.Close()

	const wantCC = "public, max-age=31536000, immutable"

	for _, path := range []string{
		"/assets/cards/2_of_clubs.svg",
		"/favicon.ico",
		"/icon.svg",
		"/apple-touch-icon.png",
	} {
		resp, err := srv.Client().Get(srv.URL + path)
		require.NoError(t, err, "get %s", path)
		require.Equal(t, http.StatusOK, resp.StatusCode, "status for %s", path)
		require.Equal(t, wantCC, resp.Header.Get("Cache-Control"), "Cache-Control for %s", path)
		_ = resp.Body.Close()
	}
}

func TestServesExtractedScripts(t *testing.T) {
	manager := session.NewManager()
	defer manager.Close()

	handler, err := NewHandler(Config{}, manager)
	require.NoError(t, err, "new handler")

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
		require.NoError(t, err, "get %s", path)
		require.Equal(t, http.StatusOK, resp.StatusCode, "status for %s", path)
		require.True(t, strings.HasPrefix(resp.Header.Get("Content-Type"), "text/javascript"),
			"expected JavaScript content type for %s, got %q", path, resp.Header.Get("Content-Type"))
		_ = resp.Body.Close()
	}

	missing, err := srv.Client().Get(srv.URL + "/assets/js/unknown/main.js")
	require.NoError(t, err, "get missing script")
	defer missing.Body.Close()
	require.Equal(t, http.StatusNotFound, missing.StatusCode)
}

func TestWebSocketJoinAndStateFlow(t *testing.T) {
	manager := session.NewManager()
	defer manager.Close()

	handler, err := NewHandler(Config{}, manager)
	require.NoError(t, err, "new handler")

	srv := httptest.NewServer(handler)
	defer srv.Close()

	ws := mustDialTableSocket(t, srv.URL, "demo")
	defer ws.Close()

	_ = readMessage(t, ws)

	require.NoError(t, ws.WriteJSON(wsCommand{Type: "join", Name: "Alice", Token: "token-a"}), "join write")

	joinMsg := readMessageType(t, ws, "join_result")
	var joinResp protocol.JoinResponse
	require.NoError(t, json.Unmarshal(joinMsg.Data, &joinResp), "decode join result")
	require.True(t, joinResp.Accepted, "expected join accepted, got rejected: %s", joinResp.Reason)
	require.NotEmpty(t, joinResp.PlayerID)

	require.NoError(t, ws.WriteJSON(wsCommand{Type: "state"}), "state write")

	stateMsg := readMessageType(t, ws, "table_state")
	var snapshot session.Snapshot
	require.NoError(t, json.Unmarshal(stateMsg.Data, &snapshot), "decode table state")
	require.Equal(t, "demo", snapshot.TableID)
	require.Len(t, snapshot.Players, 1)

	require.NoError(t, ws.WriteJSON(wsCommand{Type: "start"}), "start write")

	startMsg := readMessageType(t, ws, "start_result")
	var startResp protocol.CommandResponse
	require.NoError(t, json.Unmarshal(startMsg.Data, &startResp), "decode start response")
	require.False(t, startResp.Accepted, "expected start rejected with insufficient players")

	require.NoError(t, ws.WriteJSON(wsCommand{Type: "unknown"}), "unknown write")

	errMsg := readMessageType(t, ws, "error")
	require.NotEmpty(t, strings.TrimSpace(errMsg.Error), "expected error message for unknown command")
}

func TestWebSocketJoinReusesPlayerByToken(t *testing.T) {
	manager := session.NewManager()
	defer manager.Close()

	handler, err := NewHandler(Config{}, manager)
	require.NoError(t, err, "new handler")

	srv := httptest.NewServer(handler)
	defer srv.Close()

	first := mustDialTableSocket(t, srv.URL, "rejoin")
	defer first.Close()
	_ = readMessage(t, first)
	require.NoError(t, first.WriteJSON(wsCommand{Type: "join", Name: "Alice", Token: "stable-token"}), "first join write")
	firstJoin := readMessageType(t, first, "join_result")
	var firstResp protocol.JoinResponse
	require.NoError(t, json.Unmarshal(firstJoin.Data, &firstResp), "decode first join")
	require.True(t, firstResp.Accepted, "first join rejected: %s", firstResp.Reason)
	require.NoError(t, first.Close(), "close first websocket")

	second := mustDialTableSocket(t, srv.URL, "rejoin")
	defer second.Close()
	_ = readMessage(t, second)
	require.NoError(t, second.WriteJSON(wsCommand{Type: "join", Name: "Renamed", Token: "stable-token"}), "second join write")
	secondJoin := readMessageType(t, second, "join_result")
	var secondResp protocol.JoinResponse
	require.NoError(t, json.Unmarshal(secondJoin.Data, &secondResp), "decode second join")

	require.True(t, secondResp.Accepted, "second join rejected: %s", secondResp.Reason)
	require.Equal(t, firstResp.PlayerID, secondResp.PlayerID, "expected same player id on reconnect")
	require.Equal(t, firstResp.Seat, secondResp.Seat, "expected same seat on reconnect")
}

func TestDisconnectLeavesTableBeforeRoundStart(t *testing.T) {
	manager := session.NewManager()
	defer manager.Close()

	handler, err := NewHandler(Config{}, manager)
	require.NoError(t, err, "new handler")

	srv := httptest.NewServer(handler)
	defer srv.Close()

	alice := mustDialTableSocket(t, srv.URL, "leave-before-start")
	defer alice.Close()
	_ = readMessage(t, alice)
	require.NoError(t, alice.WriteJSON(wsCommand{Type: "join", Name: "Alice", Token: "alice-token"}), "alice join write")
	_ = readMessageType(t, alice, "join_result")

	bob := mustDialTableSocket(t, srv.URL, "leave-before-start")
	defer bob.Close()
	_ = readMessage(t, bob)
	require.NoError(t, bob.WriteJSON(wsCommand{Type: "join", Name: "Bob", Token: "bob-token"}), "bob join write")
	_ = readMessageType(t, bob, "join_result")

	require.NoError(t, alice.Close(), "close alice websocket")

	assertEventually(t, 2*time.Second, func() bool {
		snapshot := requestStateSnapshot(t, bob)
		if len(snapshot.Players) != 1 {
			return false
		}
		return snapshot.Players[0].Name == "Bob"
	})
}

func TestWebSocketAutoCreatesTable(t *testing.T) {
	manager := session.NewManager()
	defer manager.Close()

	handler, err := NewHandler(Config{}, manager)
	require.NoError(t, err, "new handler")

	srv := httptest.NewServer(handler)
	defer srv.Close()

	ws := mustDialTableSocket(t, srv.URL, "auto-create")
	defer ws.Close()

	_ = readMessage(t, ws)

	_, ok := manager.Get("auto-create")
	require.True(t, ok, "expected table to be created when websocket connected")
}

func TestTableClosesWhenLastHumanLeaves(t *testing.T) {
	manager := session.NewManager()
	defer manager.Close()

	handler, err := NewHandler(Config{}, manager)
	require.NoError(t, err, "new handler")

	srv := httptest.NewServer(handler)
	defer srv.Close()

	ws := mustDialTableSocket(t, srv.URL, "close-on-empty")
	_ = readMessage(t, ws)
	require.NoError(t, ws.WriteJSON(wsCommand{Type: "join", Name: "Alice", Token: "alice-token"}), "join write")
	_ = readMessageType(t, ws, "join_result")
	require.NoError(t, ws.Close(), "close websocket")

	assertEventually(t, 2*time.Second, func() bool {
		_, ok := manager.Get("close-on-empty")
		return !ok
	})
}

func TestPassPhaseAndReviewFlowOverWebSocket(t *testing.T) {
	manager := session.NewManager()
	defer manager.Close()

	handler, err := NewHandler(Config{}, manager)
	require.NoError(t, err, "new handler")

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
		require.NoError(t, players[i].conn.WriteJSON(wsCommand{Type: "join", Name: players[i].name, Token: players[i].token}),
			"join write for %s", players[i].name)
		join := readMessageType(t, players[i].conn, "join_result")
		var joinResp protocol.JoinResponse
		require.NoError(t, json.Unmarshal(join.Data, &joinResp), "decode join response for %s", players[i].name)
		require.True(t, joinResp.Accepted, "join rejected for %s: %s", players[i].name, joinResp.Reason)
	}

	require.NoError(t, players[0].conn.WriteJSON(wsCommand{Type: "start"}), "start write")
	startMsg := readMessageType(t, players[0].conn, "start_result")
	var startResp protocol.CommandResponse
	require.NoError(t, json.Unmarshal(startMsg.Data, &startResp), "decode start response")
	require.True(t, startResp.Accepted, "expected start accepted, got %s", startResp.Reason)

	passCardsByPlayer := make([][]string, len(players))
	for i := range players {
		var snapshot session.Snapshot
		assertEventually(t, 2*time.Second, func() bool {
			snapshot = requestStateSnapshot(t, players[i].conn)
			return snapshot.Phase == "passing" && len(snapshot.Hand) == 13
		})
		passCardsByPlayer[i] = append([]string(nil), snapshot.Hand[:3]...)
	}

	for i := range players {
		require.NoError(t, players[i].conn.WriteJSON(wsCommand{Type: "pass", Cards: passCardsByPlayer[i]}),
			"pass write for %s", players[i].name)
		passMsg := readMessageType(t, players[i].conn, "pass_result")
		var passResp protocol.CommandResponse
		require.NoError(t, json.Unmarshal(passMsg.Data, &passResp), "decode pass response for %s", players[i].name)
		require.True(t, passResp.Accepted, "pass rejected for %s: %s", players[i].name, passResp.Reason)
	}

	assertEventually(t, 2*time.Second, func() bool {
		snapshot := requestStateSnapshot(t, players[0].conn)
		return snapshot.Phase == "pass_review" && len(snapshot.PassReceived) == 3
	})

	for i := 0; i < len(players)-1; i++ {
		require.NoError(t, players[i].conn.WriteJSON(wsCommand{Type: "ready_after_pass"}),
			"ready write for %s", players[i].name)
		readyMsg := readMessageType(t, players[i].conn, "ready_after_pass_result")
		var readyResp protocol.CommandResponse
		require.NoError(t, json.Unmarshal(readyMsg.Data, &readyResp), "decode ready response for %s", players[i].name)
		require.True(t, readyResp.Accepted, "ready rejected for %s: %s", players[i].name, readyResp.Reason)
	}

	require.Equal(t, "pass_review", requestStateSnapshot(t, players[0].conn).Phase,
		"expected still pass_review before last ready")

	require.NoError(t, players[len(players)-1].conn.WriteJSON(wsCommand{Type: "ready_after_pass"}), "last ready write")
	lastReady := readMessageType(t, players[len(players)-1].conn, "ready_after_pass_result")
	var lastReadyResp protocol.CommandResponse
	require.NoError(t, json.Unmarshal(lastReady.Data, &lastReadyResp), "decode last ready response")
	require.True(t, lastReadyResp.Accepted, "last ready rejected: %s", lastReadyResp.Reason)

	assertEventually(t, 2*time.Second, func() bool {
		snapshot := requestStateSnapshot(t, players[0].conn)
		return snapshot.Phase == "playing" && snapshot.TurnPlayerID != ""
	})
}

func mustDialTableSocket(t *testing.T, baseURL, tableID string) *websocket.Conn {
	t.Helper()

	wsURL := "ws" + strings.TrimPrefix(baseURL, "http") + "/ws/table/" + tableID
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err, "dial websocket")
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

	require.NoError(t, conn.SetReadDeadline(time.Now().Add(4*time.Second)), "set read deadline")

	var message testWSMessage
	require.NoError(t, conn.ReadJSON(&message), "read websocket message")
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

func requestStateSnapshot(t *testing.T, conn *websocket.Conn) session.Snapshot {
	t.Helper()

	require.NoError(t, conn.WriteJSON(wsCommand{Type: "state"}), "state write")

	stateMsg := readMessageType(t, conn, "table_state")
	var snapshot session.Snapshot
	require.NoError(t, json.Unmarshal(stateMsg.Data, &snapshot), "decode table state")

	return snapshot
}
