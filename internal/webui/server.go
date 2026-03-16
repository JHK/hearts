package webui

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/JHK/hearts/internal/game"
	"github.com/JHK/hearts/internal/table"
	"github.com/gorilla/websocket"
)

//go:embed assets/index.html assets/table.html assets/styles.css assets/cards/*.svg assets/js assets/icon.svg assets/favicon.ico assets/apple-touch-icon.png
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
	Type     string   `json:"type"`
	Name     string   `json:"name,omitempty"`
	Token    string   `json:"token,omitempty"`
	Card     string   `json:"card,omitempty"`
	Cards    []string `json:"cards,omitempty"`
	Strategy string   `json:"strategy,omitempty"`
}

type humanPresenceTracker struct {
	mu     sync.Mutex
	counts map[string]int
}

func newHumanPresenceTracker() *humanPresenceTracker {
	return &humanPresenceTracker{counts: make(map[string]int)}
}

func (t *humanPresenceTracker) Join(tableID string) {
	t.mu.Lock()
	t.counts[tableID]++
	t.mu.Unlock()
}

func (t *humanPresenceTracker) Leave(tableID string) int {
	t.mu.Lock()
	defer t.mu.Unlock()

	remaining := t.counts[tableID] - 1
	if remaining <= 0 {
		delete(t.counts, tableID)
		return 0
	}

	t.counts[tableID] = remaining
	return remaining
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

	embeddedStyles, err := assetsFS.ReadFile("assets/styles.css")
	if err != nil {
		return nil, fmt.Errorf("read embedded styles css: %w", err)
	}

	mux := http.NewServeMux()
	presence := newHumanPresenceTracker()

	stylesPath := filepath.Join("internal", "webui", "assets", "styles.css")

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

	mux.HandleFunc("/assets/styles.css", func(w http.ResponseWriter, r *http.Request) {
		styles, err := os.ReadFile(stylesPath)
		if err != nil {
			styles = embeddedStyles
		}

		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		_, _ = w.Write(styles)
	})

	mux.HandleFunc("/assets/js/", func(w http.ResponseWriter, r *http.Request) {
		scriptFile := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/assets/js/"))
		if scriptFile == "" || strings.HasPrefix(scriptFile, "/") || strings.Contains(scriptFile, "..") || !strings.HasSuffix(scriptFile, ".js") {
			http.NotFound(w, r)
			return
		}

		scriptPath := "assets/js/" + scriptFile

		script, err := assetsFS.ReadFile(scriptPath)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		_, _ = w.Write(script)
	})

	mux.HandleFunc("/assets/cards/", func(w http.ResponseWriter, r *http.Request) {
		cardFile := strings.TrimPrefix(r.URL.Path, "/assets/cards/")
		cardFile = strings.TrimSpace(cardFile)
		if cardFile == "" || strings.Contains(cardFile, "/") || !strings.HasSuffix(cardFile, ".svg") {
			http.NotFound(w, r)
			return
		}

		content, err := assetsFS.ReadFile("assets/cards/" + cardFile)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "image/svg+xml")
		_, _ = w.Write(content)
	})

	for _, f := range []struct {
		path        string
		asset       string
		contentType string
	}{
		{"/favicon.ico", "assets/favicon.ico", "image/x-icon"},
		{"/icon.svg", "assets/icon.svg", "image/svg+xml"},
		{"/apple-touch-icon.png", "assets/apple-touch-icon.png", "image/png"},
	} {
		f := f
		data, err := assetsFS.ReadFile(f.asset)
		if err != nil {
			return nil, fmt.Errorf("read embedded %s: %w", f.asset, err)
		}
		mux.HandleFunc(f.path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", f.contentType)
			_, _ = w.Write(data)
		})
	}

	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	mux.HandleFunc("/ws/table/", func(w http.ResponseWriter, r *http.Request) {
		handleTableWebSocket(manager, presence, upgrader, w, r)
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

func handleTableWebSocket(manager *table.Manager, presence *humanPresenceTracker, upgrader websocket.Upgrader, w http.ResponseWriter, r *http.Request) {
	tableID := strings.TrimPrefix(r.URL.Path, "/ws/table/")
	tableID = strings.TrimSpace(tableID)
	if tableID == "" || strings.Contains(tableID, "/") {
		http.NotFound(w, r)
		return
	}

	runtime, _, err := manager.Create(tableID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	events, unsubscribe := runtime.Subscribe()

	out := make(chan wsMessage, 256)

	var playerMu sync.RWMutex
	var playerID game.PlayerID
	humanJoined := false

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
				if !humanJoined {
					humanJoined = true
					presence.Join(runtime.ID())
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
		default:
			send(wsMessage{Type: "error", Error: fmt.Sprintf("unknown command %q", cmd.Type)})
		}
	}

	_ = conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"), time.Now().Add(2*time.Second))
	unsubscribe()
	<-eventsDone
	close(out)
	<-writerDone

	if humanJoined {
		runtime.Leave(getPlayerID())
	}

	if humanJoined && presence.Leave(runtime.ID()) == 0 {
		manager.CloseTable(runtime.ID())
	}
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
