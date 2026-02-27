package webui

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/JHK/hearts/internal/game"
	"github.com/JHK/hearts/internal/table"
	"github.com/gorilla/websocket"
)

//go:embed assets/index.html assets/table.html
var assetsFS embed.FS

type Config struct {
	Addr string
}

type wsMessage struct {
	Type  string `json:"type"`
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

type wsCommand struct {
	Type     string `json:"type"`
	Name     string `json:"name,omitempty"`
	Token    string `json:"token,omitempty"`
	Card     string `json:"card,omitempty"`
	Strategy string `json:"strategy,omitempty"`
}

func Run(cfg Config) error {
	if strings.TrimSpace(cfg.Addr) == "" {
		cfg.Addr = "127.0.0.1:8080"
	}

	handler, err := NewHandler(nil)
	if err != nil {
		return err
	}

	return http.ListenAndServe(cfg.Addr, handler)
}

func NewHandler(manager *table.Manager) (http.Handler, error) {
	if manager == nil {
		manager = table.NewManager()
	}

	indexHTML, err := assetsFS.ReadFile("assets/index.html")
	if err != nil {
		return nil, fmt.Errorf("read embedded index html: %w", err)
	}

	tableHTML, err := assetsFS.ReadFile("assets/table.html")
	if err != nil {
		return nil, fmt.Errorf("read embedded table html: %w", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(indexHTML)
	})

	mux.HandleFunc("/table/", func(w http.ResponseWriter, r *http.Request) {
		tableID := strings.TrimPrefix(r.URL.Path, "/table/")
		tableID = strings.TrimSpace(tableID)
		if tableID == "" || strings.Contains(tableID, "/") {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(tableHTML)
	})

	mux.HandleFunc("/api/tables", func(w http.ResponseWriter, r *http.Request) {
		handleTablesAPI(manager, w, r)
	})

	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	mux.HandleFunc("/ws/table/", func(w http.ResponseWriter, r *http.Request) {
		handleTableWebSocket(manager, upgrader, w, r)
	})

	return mux, nil
}

func handleTablesAPI(manager *table.Manager, w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, map[string]any{"tables": manager.List()})
	case http.MethodPost:
		var req struct {
			TableID string `json:"table_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, wsMessage{Type: "error", Error: "invalid JSON"})
			return
		}

		runtime, created, err := manager.Create(req.TableID)
		if err != nil {
			writeJSON(w, wsMessage{Type: "error", Error: err.Error()})
			return
		}

		writeJSON(w, map[string]any{"table_id": runtime.ID(), "created": created})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleTableWebSocket(manager *table.Manager, upgrader websocket.Upgrader, w http.ResponseWriter, r *http.Request) {
	tableID := strings.TrimPrefix(r.URL.Path, "/ws/table/")
	tableID = strings.TrimSpace(tableID)
	if tableID == "" || strings.Contains(tableID, "/") {
		http.NotFound(w, r)
		return
	}

	runtime, ok := manager.Get(tableID)
	if !ok {
		http.Error(w, "table not found", http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	events, unsubscribe := runtime.Subscribe()
	defer unsubscribe()

	out := make(chan wsMessage, 256)
	defer close(out)

	var playerMu sync.RWMutex
	var playerID game.PlayerID

	setPlayerID := func(id game.PlayerID) {
		playerMu.Lock()
		playerID = id
		playerMu.Unlock()
	}

	getPlayerID := func() game.PlayerID {
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
		case "add_bot":
			added, err := runtime.AddBot(cmd.Strategy)
			if err != nil {
				send(wsMessage{Type: "error", Error: err.Error()})
				continue
			}
			send(wsMessage{Type: "add_bot_result", Data: added})
		default:
			send(wsMessage{Type: "error", Error: fmt.Sprintf("unknown command %q", cmd.Type)})
		}
	}

	_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"), time.Now().Add(2*time.Second))
	<-eventsDone
	<-writerDone
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
