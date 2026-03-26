package webui

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JHK/hearts/internal/session"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

func dialLobbyWS(t *testing.T, srvURL string) *websocket.Conn {
	t.Helper()
	wsURL := "ws" + srvURL[len("http"):] + "/ws/lobby"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	t.Cleanup(func() { conn.Close() })
	return conn
}

func readLobbyMsg(t *testing.T, conn *websocket.Conn) testWSMessage {
	t.Helper()
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var msg testWSMessage
	require.NoError(t, conn.ReadJSON(&msg))
	return msg
}

func lobbyAnnounce(t *testing.T, conn *websocket.Conn, name, token string) {
	t.Helper()
	require.NoError(t, conn.WriteJSON(map[string]string{"type": "announce", "name": name, "token": token}))
	// Read messages until we get lobby_self (may receive lobby_tables first).
	for {
		msg := readLobbyMsg(t, conn)
		if msg.Type == "lobby_self" {
			return
		}
	}
}

func readLobbySnap(t *testing.T, conn *websocket.Conn) lobbySnapshot {
	t.Helper()
	// Skip non-presence messages (e.g. lobby_tables).
	for {
		msg := readLobbyMsg(t, conn)
		if msg.Type == "lobby_presence" {
			var snap lobbySnapshot
			require.NoError(t, json.Unmarshal(msg.Data, &snap))
			return snap
		}
	}
}

func newLobbyTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	manager := session.NewManager()
	t.Cleanup(func() { manager.Close() })
	handler, err := NewHandler(Config{}, manager, nil)
	require.NoError(t, err)
	t.Cleanup(handler.Close)
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func TestLobbyPresence_JoinAndLeave(t *testing.T) {
	srv := newLobbyTestServer(t)

	// Connect first client.
	c1 := dialLobbyWS(t, srv.URL)

	// Should receive initial empty presence.
	snap := readLobbySnap(t, c1)
	require.Empty(t, snap.Players)

	// Announce as Alice.
	lobbyAnnounce(t, c1, "Alice", "tok-1")

	// Should get updated presence with Alice.
	snap = readLobbySnap(t, c1)
	require.Len(t, snap.Players, 1)
	require.Equal(t, "Alice", snap.Players[0].Name)

	// Connect second client.
	c2 := dialLobbyWS(t, srv.URL)

	// c2 gets initial snapshot with Alice.
	snap = readLobbySnap(t, c2)
	require.Len(t, snap.Players, 1)

	// Announce as Bob.
	lobbyAnnounce(t, c2, "Bob", "tok-2")

	// Both clients should now see 2 players.
	snap = readLobbySnap(t, c2)
	require.Len(t, snap.Players, 2)

	snap = readLobbySnap(t, c1)
	require.Len(t, snap.Players, 2)

	// Disconnect c2 (Bob leaves).
	c2.Close()

	// c1 should see only Alice again.
	snap = readLobbySnap(t, c1)
	require.Len(t, snap.Players, 1)
	require.Equal(t, "Alice", snap.Players[0].Name)
}

func TestLobbyPresence_NameUpdate(t *testing.T) {
	srv := newLobbyTestServer(t)

	c1 := dialLobbyWS(t, srv.URL)
	_ = readLobbySnap(t, c1) // initial empty

	lobbyAnnounce(t, c1, "Alice", "tok-1")
	_ = readLobbySnap(t, c1) // presence with Alice

	// Update name via re-announce.
	require.NoError(t, c1.WriteJSON(map[string]string{"type": "announce", "name": "Alicia", "token": "tok-1"}))

	snap := readLobbySnap(t, c1)
	require.Len(t, snap.Players, 1)
	require.Equal(t, "Alicia", snap.Players[0].Name)
}

func TestLobbyPresence_MultiTab(t *testing.T) {
	srv := newLobbyTestServer(t)

	// Two tabs with the same token.
	tab1 := dialLobbyWS(t, srv.URL)
	_ = readLobbySnap(t, tab1) // initial empty
	lobbyAnnounce(t, tab1, "Alice", "tok-1")
	_ = readLobbySnap(t, tab1) // Alice present

	tab2 := dialLobbyWS(t, srv.URL)
	snap := readLobbySnap(t, tab2) // initial with Alice
	require.Len(t, snap.Players, 1)
	lobbyAnnounce(t, tab2, "Alice", "tok-1")

	// Both tabs registered — still just one player.
	snap = readLobbySnap(t, tab2)
	require.Len(t, snap.Players, 1)

	// Also arrives on tab1.
	snap = readLobbySnap(t, tab1)
	require.Len(t, snap.Players, 1)

	// Close first tab — player should remain (second tab still open).
	tab1.Close()

	// Add a second user to trigger a broadcast so we can observe the state.
	observer := dialLobbyWS(t, srv.URL)
	snap = readLobbySnap(t, observer) // initial snapshot
	require.Len(t, snap.Players, 1, "Alice should still be present after closing one tab")
	require.Equal(t, "Alice", snap.Players[0].Name)

	// Now close second tab — player should be removed.
	tab2.Close()

	snap = readLobbySnap(t, observer)
	require.Empty(t, snap.Players, "Alice should be gone after both tabs closed")
}

func TestLobbyPresence_NameTruncated(t *testing.T) {
	srv := newLobbyTestServer(t)

	c1 := dialLobbyWS(t, srv.URL)
	_ = readLobbySnap(t, c1) // initial empty

	longName := strings.Repeat("A", 100)
	lobbyAnnounce(t, c1, longName, "tok-1")

	snap := readLobbySnap(t, c1)
	require.Len(t, snap.Players, 1)
	require.Len(t, snap.Players[0].Name, maxLobbyNameLen)
}
