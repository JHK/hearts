package webui

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/JHK/hearts/internal/protocol"
	"github.com/JHK/hearts/internal/session"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
)

// templateData holds the values injected into HTML templates.
type templateData struct {
	StylesURL    string
	ScriptURL    string
	ChartJSURL   string
	ExtraScripts template.HTML
}

//go:embed assets/index.html assets/table.html assets/styles.css assets/cards/*.svg assets/js assets/icon.svg assets/favicon.ico assets/apple-touch-icon.png
var assetsFS embed.FS

type Config struct {
	Addr string
	Dev  bool
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
	Seat     *int     `json:"seat,omitempty"`
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

func (t *humanPresenceTracker) Count(tableID string) int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.counts[tableID]
}

// playerPresenceTracker counts active WebSocket connections per player per table.
// This prevents spurious Leave calls when a player has multiple tabs open.
type playerPresenceTracker struct {
	mu     sync.Mutex
	counts map[string]int // key: "tableID\x00playerID"
}

func newPlayerPresenceTracker() *playerPresenceTracker {
	return &playerPresenceTracker{counts: make(map[string]int)}
}

func (t *playerPresenceTracker) Join(tableID string, playerID string) {
	t.mu.Lock()
	t.counts[tableID+"\x00"+playerID]++
	t.mu.Unlock()
}

// Leave decrements the count and returns the remaining connections.
func (t *playerPresenceTracker) Leave(tableID string, playerID string) int {
	t.mu.Lock()
	defer t.mu.Unlock()

	key := tableID + "\x00" + playerID
	remaining := t.counts[key] - 1
	if remaining <= 0 {
		delete(t.counts, key)
		return 0
	}
	t.counts[key] = remaining
	return remaining
}

func Run(cfg Config) error {
	if strings.TrimSpace(cfg.Addr) == "" {
		cfg.Addr = ":8080"
	}

	handler, err := NewHandler(cfg, nil)
	if err != nil {
		return err
	}

	return http.ListenAndServe(cfg.Addr, handler)
}

func NewHandler(cfg Config, manager *session.Manager) (http.Handler, error) {
	if manager == nil {
		manager = session.NewManager()
	}

	indexTmpl, err := template.New("index").Parse(string(mustReadAsset("assets/index.html")))
	if err != nil {
		return nil, fmt.Errorf("parse index template: %w", err)
	}
	tableTmpl, err := template.New("table").Parse(string(mustReadAsset("assets/table.html")))
	if err != nil {
		return nil, fmt.Errorf("parse table template: %w", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	presence := newHumanPresenceTracker()
	playerPresence := newPlayerPresenceTracker()
	lobby := newLobbyHub()

	indexData := templateData{
		StylesURL: "/assets/styles.css",
		ScriptURL: "/assets/js/lobby/main.js",
	}
	tableData := templateData{
		StylesURL:  "/assets/styles.css",
		ScriptURL:  "/assets/js/table/main.js",
		ChartJSURL: "/assets/js/vendor/chart.umd.js",
	}

	if cfg.Dev {
		registerDevAssetHandlers(r)
		tableData.ExtraScripts = `<script src="/dev.js"></script>`
	} else {
		fp, fpErr := buildFingerprintedAssets(assetsFS)
		if fpErr != nil {
			return nil, fmt.Errorf("build fingerprinted assets: %w", fpErr)
		}
		indexData.StylesURL = fp.urlMapping["/assets/styles.css"]
		indexData.ScriptURL = fp.urlMapping["/assets/js/lobby/main.js"]
		tableData.StylesURL = fp.urlMapping["/assets/styles.css"]
		tableData.ScriptURL = fp.urlMapping["/assets/js/table/main.js"]
		tableData.ChartJSURL = fp.urlMapping["/assets/js/vendor/chart.umd.js"]
		registerFingerprintedAssetHandlers(r, fp)
	}

	indexHTML := mustRenderTemplate(indexTmpl, indexData)
	tableHTML := mustRenderTemplate(tableTmpl, tableData)

	indexETag := contentETag(indexHTML)
	tableETag := contentETag(tableHTML)

	// HTML pages group
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		serveHTMLWithETag(w, r, indexHTML, indexETag)
	})

	r.Get("/table/{tableID}", func(w http.ResponseWriter, r *http.Request) {
		tableID := chi.URLParam(r, "tableID")
		if tableID == "" {
			http.NotFound(w, r)
			return
		}

		if _, ok := manager.Get(tableID); !ok {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		serveHTMLWithETag(w, r, tableHTML, tableETag)
	})

	// Static assets group — immutable cache middleware
	r.Route("/assets", func(assets chi.Router) {
		assets.Use(immutableCacheMiddleware)

		assets.Get("/cards/{cardFile}", func(w http.ResponseWriter, r *http.Request) {
			cardFile := chi.URLParam(r, "cardFile")
			if cardFile == "" || !strings.HasSuffix(cardFile, ".svg") {
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
	})

	// Favicon/icon assets — immutable cache
	for _, f := range []struct {
		path        string
		asset       string
		contentType string
	}{
		{"/favicon.ico", "assets/favicon.ico", "image/x-icon"},
		{"/icon.svg", "assets/icon.svg", "image/svg+xml"},
		{"/apple-touch-icon.png", "assets/apple-touch-icon.png", "image/png"},
	} {
		data, err := assetsFS.ReadFile(f.asset)
		if err != nil {
			return nil, fmt.Errorf("read embedded %s: %w", f.asset, err)
		}
		contentType := f.contentType
		content := data
		r.With(immutableCacheMiddleware).Get(f.path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", contentType)
			_, _ = w.Write(content)
		})
	}

	// API endpoints group
	r.Route("/api", func(api chi.Router) {
		api.Get("/tables", func(w http.ResponseWriter, r *http.Request) {
			handleTablesAPI(manager, w, r)
		})
		api.Post("/tables", func(w http.ResponseWriter, r *http.Request) {
			handleTablesAPI(manager, w, r)
		})
	})

	// Dev-only routes group
	if cfg.Dev {
		devJS := []byte(`window.debugBot = async function(opts) {
  opts = opts || {};
  const tableID = opts.tableID || window.location.pathname.replace('/table/', '');
  const fmt = opts.json ? 'json' : 'markdown';
  const r = await fetch('/api/debug/bots?table_id=' + encodeURIComponent(tableID) + '&format=' + fmt);
  if (!r.ok) { console.error('debugBot:', r.status, await r.text()); return; }
  if (fmt === 'json') { const data = await r.json(); console.log(data); return data; }
  const text = await r.text();
  console.log(text);
  return text;
};
console.log('[dev] debugBot() — full bot decision context (markdown). debugBot({json:true}) for JSON.');
`)
		r.Get("/dev.js", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
			_, _ = w.Write(devJS)
		})

		r.Get("/api/debug/bots", func(w http.ResponseWriter, r *http.Request) {
			tableID := r.URL.Query().Get("table_id")
			rt, ok := manager.Get(tableID)
			if !ok {
				http.Error(w, "table not found", http.StatusNotFound)
				return
			}
			snap := rt.DebugBotContext()
			if snap == nil {
				http.Error(w, "table stopped", http.StatusGone)
				return
			}
			if r.URL.Query().Get("format") == "json" {
				writeJSON(w, snap)
				return
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, _ = w.Write([]byte(snap.FormatMarkdown()))
		})
	}

	// WebSocket endpoints group
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	r.Get("/ws/lobby", func(w http.ResponseWriter, r *http.Request) {
		handleLobbyWebSocket(lobby, upgrader, w, r)
	})
	r.Get("/ws/table/{tableID}", func(w http.ResponseWriter, r *http.Request) {
		handleTableWebSocket(manager, presence, playerPresence, upgrader, w, r)
	})

	return r, nil
}

// immutableCacheMiddleware sets Cache-Control headers for immutable static assets.
func immutableCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
}

func handleTablesAPI(manager *session.Manager, w http.ResponseWriter, r *http.Request) {
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

const maxLobbyNameLen = 32
const maxLobbyTokenLen = 128

func handleLobbyWebSocket(lobby *lobbyHub, upgrader websocket.Upgrader, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	slog.Info("lobby client connected", "event", "lobby_connected", "addr", r.RemoteAddr)

	events, unsubscribe := lobby.Subscribe()

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

	// Send initial snapshot before starting the events goroutine so that
	// broadcasts triggered between Subscribe and here can't overtake it.
	send(wsMessage{Type: "lobby_presence", Data: lobby.Snapshot()})

	// Forward presence updates to the client.
	eventsDone := make(chan struct{})
	go func() {
		defer close(eventsDone)
		for snap := range events {
			send(wsMessage{Type: "lobby_presence", Data: snap})
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
			} else if cmd.Token == token {
				lobby.UpdateName(token, name)
			}
		}
	}

	if token != "" {
		lobby.Leave(token)
	}

	slog.Info("lobby client disconnected", "event", "lobby_disconnected", "addr", r.RemoteAddr)

	unsubscribe()
	<-eventsDone
	close(out)
	<-writerDone
}

func handleTableWebSocket(manager *session.Manager, presence *humanPresenceTracker, playerPresence *playerPresenceTracker, upgrader websocket.Upgrader, w http.ResponseWriter, r *http.Request) {
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
		tableID := runtime.ID()
		gracePeriod := 60 * time.Second
		if info := runtime.Info(); !info.Started {
			gracePeriod = 500 * time.Millisecond
		}
		go func() {
			timer := time.NewTimer(gracePeriod)
			defer timer.Stop()
			<-timer.C
			if presence.Count(tableID) == 0 {
				slog.Warn("closing orphaned table after grace period", "event", "table_closed_orphaned", "table_id", tableID)
				manager.CloseTable(tableID)
			}
		}()
	}
}

// registerDevAssetHandlers serves CSS and JS at their plain paths without
// fingerprinting or cache headers, so hot-reload works without stale-cache issues.
func registerDevAssetHandlers(router chi.Router) {
	stylesPath := filepath.Join("internal", "webui", "assets", "styles.css")
	embeddedStyles, _ := assetsFS.ReadFile("assets/styles.css")

	router.Get("/assets/styles.css", func(w http.ResponseWriter, r *http.Request) {
		styles, err := os.ReadFile(stylesPath)
		if err != nil {
			styles = embeddedStyles
		}
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		_, _ = w.Write(styles)
	})

	router.Get("/assets/js/*", func(w http.ResponseWriter, r *http.Request) {
		scriptFile := strings.TrimSpace(strings.TrimPrefix(r.URL.Path, "/assets/js/"))
		if scriptFile == "" || strings.HasPrefix(scriptFile, "/") || strings.Contains(scriptFile, "..") || !strings.HasSuffix(scriptFile, ".js") {
			http.NotFound(w, r)
			return
		}

		script, err := assetsFS.ReadFile("assets/js/" + scriptFile)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		_, _ = w.Write(script)
	})
}

func mustReadAsset(name string) []byte {
	data, err := assetsFS.ReadFile(name)
	if err != nil {
		panic(fmt.Sprintf("read embedded %s: %v", name, err))
	}
	return data
}

func mustRenderTemplate(tmpl *template.Template, data templateData) []byte {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		panic(fmt.Sprintf("render template %s: %v", tmpl.Name(), err))
	}
	return buf.Bytes()
}

func contentETag(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf(`"%x"`, h[:16])
}

func serveHTMLWithETag(w http.ResponseWriter, r *http.Request, content []byte, etag string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("ETag", etag)

	if match := r.Header.Get("If-None-Match"); match == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	_, _ = w.Write(content)
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
