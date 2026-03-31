package webui

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/JHK/hearts/internal/protocol"
	"github.com/JHK/hearts/internal/session"
	"github.com/JHK/hearts/internal/webui/tracker"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

// wsMessage is the envelope for server→client WebSocket messages.
type wsMessage struct {
	Type  string `json:"type"`
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

// wsCommand is the envelope for client→server WebSocket commands.
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

const maxLobbyNameLen = 32
const maxLobbyTokenLen = 128

// registerWSRoutes mounts WebSocket upgrade endpoints.
func registerWSRoutes(r chi.Router, manager *session.Manager, lobby *lobbyHub, presence *tracker.HumanPresence, playerPresence *tracker.PlayerPresence, ct *tracker.ConnTracker, devMode bool) {
	upgrader := websocket.Upgrader{CheckOrigin: checkOriginFunc(devMode)}
	r.Get("/ws/lobby", func(w http.ResponseWriter, r *http.Request) {
		handleLobbyWebSocket(manager, lobby, ct, upgrader, w, r)
	})
	r.Get("/ws/table/{tableID}", func(w http.ResponseWriter, r *http.Request) {
		handleTableWebSocket(manager, presence, playerPresence, ct, upgrader, w, r)
	})
}

// checkOriginFunc returns a CheckOrigin function for the WebSocket upgrader.
// In dev mode all origins are accepted. In production the Origin header must
// match the request Host.
func checkOriginFunc(devMode bool) func(*http.Request) bool {
	if devMode {
		return func(r *http.Request) bool { return true }
	}
	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true // non-browser clients (e.g. curl) don't send Origin
		}
		u, err := url.Parse(origin)
		if err != nil || u.Host == "" || u.User != nil {
			slog.Warn("websocket bad origin", "origin", origin, "remote", r.RemoteAddr)
			return false
		}
		if u.Host != r.Host {
			slog.Warn("websocket origin mismatch", "origin", origin, "host", r.Host, "remote", r.RemoteAddr)
			return false
		}
		return true
	}
}

func handleLobbyWebSocket(manager *session.Manager, lobby *lobbyHub, ct *tracker.ConnTracker, upgrader websocket.Upgrader, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	ct.Track(conn)
	defer ct.Untrack(conn)

	slog.Info("lobby client connected", "event", "lobby_connected", "addr", r.RemoteAddr)

	presenceEvents, unsubPresence := lobby.Subscribe()
	tableEvents, unsubTables := manager.Subscribe()

	out := make(chan wsMessage, 16)
	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		for msg := range out {
			_ = conn.SetWriteDeadline(time.Now().Add(15 * time.Second))
			if err := conn.WriteJSON(msg); err != nil {
				return
			}
		}
	}()

	send := func(msg wsMessage) {
		select {
		case out <- msg:
		default:
		}
	}

	// Send initial snapshots before starting the events goroutines so that
	// broadcasts triggered between Subscribe and here can't overtake them.
	send(wsMessage{Type: "lobby_presence", Data: lobby.Snapshot()})
	send(wsMessage{Type: "lobby_tables", Data: map[string]any{"tables": manager.List()}})

	// Forward presence updates to the client.
	presenceDone := make(chan struct{})
	go func() {
		defer close(presenceDone)
		for snap := range presenceEvents {
			send(wsMessage{Type: "lobby_presence", Data: snap})
		}
	}()

	// Forward table list changes to the client.
	tablesDone := make(chan struct{})
	go func() {
		defer close(tablesDone)
		for range tableEvents {
			send(wsMessage{Type: "lobby_tables", Data: map[string]any{"tables": manager.List()}})
		}
	}()

	var token string
	var myID int
	for {
		var cmd wsCommand
		if err := conn.ReadJSON(&cmd); err != nil {
			break
		}
		switch cmd.Type {
		case "announce":
			if cmd.Token == "" {
				send(wsMessage{Type: "error", Error: "token required"})
				continue
			}
			if len(cmd.Token) > maxLobbyTokenLen {
				send(wsMessage{Type: "error", Error: "token too long"})
				continue
			}
			name := truncateUTF8(cmd.Name, maxLobbyNameLen)
			if name == "" {
				name = "Player"
			}
			if token == "" {
				token = cmd.Token
				myID = lobby.Join(token, name)
				send(wsMessage{Type: "lobby_self", Data: map[string]any{"id": myID}})
				lobby.Broadcast()
			} else if cmd.Token == token {
				lobby.UpdateName(token, name)
			}
		case "create_table":
			runtime, created, err := manager.Create(cmd.TableID)
			if err != nil {
				send(wsMessage{Type: "error", Error: err.Error()})
				continue
			}
			send(wsMessage{Type: "create_table_result", Data: map[string]any{
				"table_id": runtime.ID(),
				"created":  created,
			}})
		}
	}

	if token != "" {
		lobby.Leave(token)
	}

	slog.Info("lobby client disconnected", "event", "lobby_disconnected", "addr", r.RemoteAddr)

	unsubPresence()
	unsubTables()
	<-presenceDone
	<-tablesDone
	close(out)
	<-writerDone
}

func handleTableWebSocket(manager *session.Manager, presence *tracker.HumanPresence, playerPresence *tracker.PlayerPresence, ct *tracker.ConnTracker, upgrader websocket.Upgrader, w http.ResponseWriter, r *http.Request) {
	tableID := chi.URLParam(r, "tableID")
	if tableID == "" {
		http.NotFound(w, r)
		return
	}

	runtime, ok := manager.Get(tableID)
	if !ok {
		http.NotFound(w, r)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	ct.Track(conn)
	defer ct.Untrack(conn)

	slog.Info("player connected", "event", "player_connected", "table_id", tableID, "addr", r.RemoteAddr)

	events, unsubscribe := runtime.Subscribe()

	out := make(chan wsMessage, 256)

	var playerMu sync.RWMutex
	var playerID protocol.PlayerID
	humanJoined := false

	setPlayerID := func(id protocol.PlayerID) {
		playerMu.Lock()
		playerID = id
		playerMu.Unlock()
	}

	getPlayerID := func() protocol.PlayerID {
		playerMu.RLock()
		defer playerMu.RUnlock()
		return playerID
	}

	send := func(message wsMessage) {
		select {
		case out <- message:
		default:
		}
	}

	writerDone := make(chan struct{})
	go func() {
		defer close(writerDone)
		for message := range out {
			_ = conn.SetWriteDeadline(time.Now().Add(15 * time.Second))
			if err := conn.WriteJSON(message); err != nil {
				return
			}
		}
	}()

	eventsDone := make(chan struct{})
	go func() {
		defer close(eventsDone)
		for event := range events {
			current := getPlayerID()
			if event.PrivateTo != "" && current != event.PrivateTo {
				continue
			}
			send(wsMessage{Type: string(event.Type), Data: event.Data})
		}
	}()

	send(wsMessage{Type: "table_state", Data: runtime.Snapshot("")})

	for {
		var cmd wsCommand
		if err := conn.ReadJSON(&cmd); err != nil {
			break
		}

		switch cmd.Type {
		case "join":
			joinResp, err := runtime.Join(cmd.Name, cmd.Token)
			if err != nil {
				send(wsMessage{Type: "error", Error: err.Error()})
				continue
			}
			send(wsMessage{Type: "join_result", Data: joinResp})
			if joinResp.Accepted {
				setPlayerID(joinResp.PlayerID)
				if !humanJoined {
					humanJoined = true
					presence.Join(runtime.ID())
					playerPresence.Join(runtime.ID(), string(joinResp.PlayerID))
				}
				send(wsMessage{Type: "table_state", Data: runtime.Snapshot(joinResp.PlayerID)})
			}
		case "state":
			send(wsMessage{Type: "table_state", Data: runtime.Snapshot(getPlayerID())})
		case "start":
			current := getPlayerID()
			if current == "" {
				send(wsMessage{Type: "error", Error: "join first"})
				continue
			}
			send(wsMessage{Type: "start_result", Data: runtime.Start(current)})
		case "play":
			current := getPlayerID()
			if current == "" {
				send(wsMessage{Type: "error", Error: "join first"})
				continue
			}
			send(wsMessage{Type: "play_result", Data: runtime.Play(current, cmd.Card)})
		case "pass":
			current := getPlayerID()
			if current == "" {
				send(wsMessage{Type: "error", Error: "join first"})
				continue
			}
			send(wsMessage{Type: "pass_result", Data: runtime.Pass(current, cmd.Cards)})
		case "ready_after_pass":
			current := getPlayerID()
			if current == "" {
				send(wsMessage{Type: "error", Error: "join first"})
				continue
			}
			send(wsMessage{Type: "ready_after_pass_result", Data: runtime.ReadyAfterPass(current)})
		case "add_bot":
			added, err := runtime.AddBot(cmd.Strategy)
			if err != nil {
				send(wsMessage{Type: "error", Error: err.Error()})
				continue
			}
			send(wsMessage{Type: "add_bot_result", Data: added})
		case "resume_game":
			current := getPlayerID()
			if current == "" {
				send(wsMessage{Type: "error", Error: "join first"})
				continue
			}
			send(wsMessage{Type: "resume_game_result", Data: runtime.ResumeGame(current)})
		case "rematch":
			current := getPlayerID()
			if current == "" {
				send(wsMessage{Type: "error", Error: "join first"})
				continue
			}
			send(wsMessage{Type: "rematch_result", Data: runtime.Rematch(current)})
		case "rename":
			current := getPlayerID()
			if current == "" {
				send(wsMessage{Type: "error", Error: "join first"})
				continue
			}
			name := truncateUTF8(cmd.Name, maxLobbyNameLen)
			send(wsMessage{Type: "rename_result", Data: runtime.Rename(current, name)})
		case "claim_seat":
			if getPlayerID() != "" {
				send(wsMessage{Type: "error", Error: "already seated"})
				continue
			}
			if cmd.Seat == nil {
				send(wsMessage{Type: "error", Error: "seat is required"})
				continue
			}
			claimResp, err := runtime.ClaimSeat(*cmd.Seat, cmd.Name, cmd.Token)
			if err != nil {
				send(wsMessage{Type: "error", Error: err.Error()})
				continue
			}
			send(wsMessage{Type: "claim_seat_result", Data: claimResp})
			if claimResp.Accepted {
				setPlayerID(claimResp.PlayerID)
				if !humanJoined {
					humanJoined = true
					presence.Join(runtime.ID())
					playerPresence.Join(runtime.ID(), string(claimResp.PlayerID))
				}
				send(wsMessage{Type: "table_state", Data: runtime.Snapshot(claimResp.PlayerID)})
			}
		default:
			send(wsMessage{Type: "error", Error: fmt.Sprintf("unknown command %q", cmd.Type)})
		}
	}

	_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"), time.Now().Add(2*time.Second))
	unsubscribe()
	<-eventsDone
	close(out)
	<-writerDone

	slog.Info("player disconnected", "event", "player_disconnected", "table_id", tableID, "player_id", string(getPlayerID()), "addr", r.RemoteAddr)

	if humanJoined {
		pid := getPlayerID()
		if playerPresence.Leave(runtime.ID(), string(pid)) == 0 {
			runtime.Leave(pid)
		}
	}

	if humanJoined && presence.Leave(runtime.ID()) == 0 {
		slog.Warn("table orphaned", "event", "table_orphaned", "table_id", runtime.ID())
		gracePeriod := 60 * time.Second
		if info := runtime.Info(); !info.Started {
			gracePeriod = 500 * time.Millisecond
		}
		scheduleOrphanCleanup(runtime.ID(), gracePeriod, presence, manager.CloseTable)
	}
}

// scheduleOrphanCleanup starts a background timer that closes the table
// if no humans reconnect within the grace period.
func scheduleOrphanCleanup(tableID string, gracePeriod time.Duration, presence *tracker.HumanPresence, closeTable func(string) bool) {
	go func() {
		timer := time.NewTimer(gracePeriod)
		defer timer.Stop()
		<-timer.C
		if presence.Count(tableID) == 0 {
			slog.Warn("closing orphaned table after grace period", "event", "table_closed_orphaned", "table_id", tableID)
			closeTable(tableID)
		}
	}()
}

// truncateUTF8 truncates s to at most maxRunes runes without splitting
// multi-byte characters.
func truncateUTF8(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes])
}
