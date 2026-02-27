package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/JHK/hearts/internal/game"
	"github.com/JHK/hearts/internal/protocol"
	natswire "github.com/JHK/hearts/internal/transport/nats"
)

type uiState struct {
	mu     sync.Mutex
	status string
	hand   []string
	logs   []string
}

func (s *uiState) setStatus(status string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = status
}

func (s *uiState) setHand(cards []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hand = append(s.hand[:0], cards...)
}

func (s *uiState) appendLog(format string, args ...any) {
	line := fmt.Sprintf(format, args...)
	prefix := time.Now().Format("15:04:05")

	s.mu.Lock()
	defer s.mu.Unlock()

	s.logs = append(s.logs, fmt.Sprintf("[%s] %s", prefix, line))
	if len(s.logs) > 300 {
		s.logs = s.logs[len(s.logs)-300:]
	}
}

func (s *uiState) snapshot() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()

	hand := "(empty)"
	if len(s.hand) > 0 {
		hand = strings.Join(s.hand, " ")
	}

	logs := make([]string, len(s.logs))
	copy(logs, s.logs)

	return map[string]any{
		"status": s.status,
		"hand":   hand,
		"logs":   logs,
	}
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func runWebMode(cfg webCommand) {
	if strings.TrimSpace(cfg.Name) == "" {
		cfg.Name = "Player"
	}
	if strings.TrimSpace(cfg.TableID) == "" {
		cfg.TableID = "default"
	}
	if strings.TrimSpace(cfg.Server) == "" {
		cfg.Server = "nats://127.0.0.1:4222"
	}

	state := &uiState{}
	session := NewSession(cfg.Name, cfg.Server)
	defer session.Shutdown()
	var actionMu sync.Mutex

	refreshStatus := func() {
		status := session.Status()
		connected := "no"
		if status.Connected {
			connected = "yes"
		}

		table := status.TableID
		if table == "" {
			table = "none"
		}

		state.setStatus(fmt.Sprintf("player=%s connected=%s server=%s table=%s", status.PlayerName, connected, status.Server, table))
		state.setHand(session.HandSnapshot())
	}

	handlers := natswire.ParticipantEventHandlers{
		OnPlayerJoined: func(data protocol.PlayerJoinedData) {
			state.appendLog("%s joined (seat %d)", data.Player.Name, data.Player.Seat)
		},
		OnGameStarted: func() {
			state.appendLog("round started")
		},
		OnTurnChanged: func(data protocol.TurnChangedData) {
			state.appendLog("turn: %s (trick %d)", data.PlayerID, data.TrickNumber+1)
		},
		OnCardPlayed: func(data protocol.CardPlayedData) {
			state.appendLog("%s played %s", data.PlayerID, data.Card)
		},
		OnTrickCompleted: func(data protocol.TrickCompletedData) {
			state.appendLog("trick %d won by %s (+%d)", data.TrickNumber+1, data.WinnerPlayerID, data.Points)
		},
		OnRoundCompleted: func(data protocol.RoundCompletedData) {
			state.appendLog("round over points=%v totals=%v", data.RoundPoints, data.TotalPoints)
		},
		OnHandUpdated: func(_ game.PlayerID, data protocol.HandUpdatedData) {
			state.setHand(data.Cards)
		},
		OnYourTurn: func(_ game.PlayerID, data protocol.YourTurnData) {
			state.appendLog("your turn (trick %d)", data.TrickNumber+1)
		},
		OnDecodeError: func(err error) {
			state.appendLog("event decode error: %v", err)
		},
	}

	reconnectAndJoin := func() error {
		actionMu.Lock()
		defer actionMu.Unlock()

		if err := session.Connect(cfg.Server); err != nil {
			refreshStatus()
			return err
		}

		joinResp, err := session.JoinTable(cfg.TableID, handlers)
		if err != nil {
			refreshStatus()
			return err
		}

		state.appendLog("joined table %s as %s (seat %d)", joinResp.TableID, cfg.Name, joinResp.Seat)
		refreshStatus()
		return nil
	}

	refreshStatus()
	state.appendLog("startup: connecting to %s and joining table %s", cfg.Server, cfg.TableID)
	go func() {
		if err := reconnectAndJoin(); err != nil {
			state.appendLog("startup join failed: %v", err)
		}
	}()

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(pageHTML))
	})

	mux.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, state.snapshot())
	})

	mux.HandleFunc("/api/reconnect", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := reconnectAndJoin(); err != nil {
			state.appendLog("reconnect/join failed: %v", err)
			writeJSON(w, map[string]any{"ok": false, "error": err.Error()})
			return
		}

		writeJSON(w, map[string]any{"ok": true})
	})

	mux.HandleFunc("/api/start", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		actionMu.Lock()
		defer actionMu.Unlock()

		if err := session.StartRound(); err != nil {
			state.appendLog("start rejected: %v", err)
			refreshStatus()
			writeJSON(w, map[string]any{"ok": false, "error": err.Error()})
			return
		}

		state.appendLog("start accepted")
		refreshStatus()
		writeJSON(w, map[string]any{"ok": true})
	})

	mux.HandleFunc("/api/addbot", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		actionMu.Lock()
		defer actionMu.Unlock()

		added, err := session.AddBot("")
		if err != nil {
			state.appendLog("addbot failed: %v", err)
			refreshStatus()
			writeJSON(w, map[string]any{"ok": false, "error": err.Error()})
			return
		}

		state.appendLog("added bot %s (%s strategy=%s) seat %d", added.JoinResponse.PlayerID, added.Name, added.Strategy, added.JoinResponse.Seat)
		refreshStatus()
		writeJSON(w, map[string]any{"ok": true})
	})

	type playRequest struct {
		Card string `json:"card"`
	}

	mux.HandleFunc("/api/play", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		actionMu.Lock()
		defer actionMu.Unlock()

		var req playRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, map[string]any{"ok": false, "error": "invalid JSON"})
			return
		}

		card := strings.ToUpper(strings.TrimSpace(req.Card))
		if card == "" {
			writeJSON(w, map[string]any{"ok": false, "error": "card is required"})
			return
		}

		if err := session.PlayCard(card); err != nil {
			state.appendLog("play rejected: %v", err)
			refreshStatus()
			writeJSON(w, map[string]any{"ok": false, "error": err.Error()})
			return
		}

		state.appendLog("played %s", card)
		refreshStatus()
		writeJSON(w, map[string]any{"ok": true})
	})

	serverURL := fmt.Sprintf("http://%s", cfg.Addr)
	state.appendLog("web UI listening at %s", serverURL)

	if cfg.OpenBrowser {
		go openBrowser(serverURL)
	}

	log.Printf("Hearts web UI: %s", serverURL)
	if err := http.ListenAndServe(cfg.Addr, mux); err != nil {
		log.Printf("web server failed: %v", err)
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		return
	}
	_ = cmd.Start()
}

const pageHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>Hearts Web</title>
  <style>
    :root {
      --bg1: #f4f7fb;
      --bg2: #dfe9f3;
      --card: #ffffffee;
      --ink: #132237;
      --muted: #4f6580;
      --accent: #b4412f;
      --accent-2: #1f6f5f;
      --line: #c3d2e3;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: "Segoe UI", "Noto Sans", "Liberation Sans", sans-serif;
      color: var(--ink);
      background: radial-gradient(circle at 20% 0%, #ffffff 0, var(--bg1) 35%, var(--bg2) 100%);
      min-height: 100vh;
    }
    .wrap {
      max-width: 980px;
      margin: 0 auto;
      padding: 20px;
      display: grid;
      gap: 14px;
    }
    .panel {
      background: var(--card);
      border: 1px solid var(--line);
      border-radius: 14px;
      padding: 14px;
      box-shadow: 0 8px 24px rgba(20, 40, 60, 0.09);
    }
    h1 { margin: 0 0 10px 0; font-size: 1.35rem; }
    .muted { color: var(--muted); }
    .row { display: flex; gap: 10px; flex-wrap: wrap; align-items: center; }
    button {
      border: none;
      border-radius: 10px;
      padding: 10px 14px;
      color: white;
      background: linear-gradient(135deg, var(--accent), #da6a51);
      cursor: pointer;
      font-weight: 600;
    }
    button.secondary { background: linear-gradient(135deg, var(--accent-2), #2c9883); }
    input {
      border: 1px solid var(--line);
      border-radius: 10px;
      padding: 10px 12px;
      min-width: 180px;
      font-size: 1rem;
    }
    pre {
      margin: 0;
      white-space: pre-wrap;
      line-height: 1.35;
      max-height: 56vh;
      overflow: auto;
      font-family: "Cascadia Mono", "DejaVu Sans Mono", monospace;
      color: #173452;
    }
    @media (max-width: 640px) {
      .wrap { padding: 10px; }
      button, input { width: 100%; }
      .row { align-items: stretch; }
    }
  </style>
</head>
<body>
  <main class="wrap">
    <section class="panel">
      <h1>Hearts Web</h1>
      <div id="status" class="muted">loading...</div>
      <div id="hand" style="margin-top:8px;"></div>
    </section>

    <section class="panel">
      <div class="row">
        <button id="reconnect">Reconnect + Join Table</button>
        <button id="addbot" class="secondary">Add Bot</button>
        <button id="start" class="secondary">Start</button>
      </div>
      <div class="row" style="margin-top: 10px;">
        <input id="card" placeholder="Card, e.g. QS" />
        <button id="play">Play</button>
      </div>
      <div id="result" class="muted" style="margin-top:8px;"></div>
    </section>

    <section class="panel">
      <h1>Event Log</h1>
      <pre id="logs"></pre>
    </section>
  </main>

  <script>
    const statusEl = document.getElementById('status');
    const handEl = document.getElementById('hand');
    const logsEl = document.getElementById('logs');
    const resultEl = document.getElementById('result');
    const cardEl = document.getElementById('card');

    async function refresh() {
      const res = await fetch('/api/state');
      const data = await res.json();
      statusEl.textContent = data.status || '';
      handEl.textContent = 'hand: ' + (data.hand || '(empty)');
      logsEl.textContent = (data.logs || []).join('\n');
      logsEl.scrollTop = logsEl.scrollHeight;
    }

    async function postJSON(path, payload) {
      const res = await fetch(path, {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify(payload || {})
      });
      return res.json();
    }

    async function doAction(path, payload) {
      const data = await postJSON(path, payload);
      resultEl.textContent = data.ok ? 'ok' : ('error: ' + (data.error || 'unknown'));
      await refresh();
    }

    document.getElementById('reconnect').onclick = () => doAction('/api/reconnect');
    document.getElementById('addbot').onclick = () => doAction('/api/addbot');
    document.getElementById('start').onclick = () => doAction('/api/start');
    document.getElementById('play').onclick = () => {
      const card = cardEl.value.trim().toUpperCase();
      doAction('/api/play', {card});
      cardEl.value = '';
    };
    cardEl.addEventListener('keydown', (e) => {
      if (e.key === 'Enter') {
        document.getElementById('play').click();
      }
    });

    refresh();
    setInterval(refresh, 1000);
  </script>
</body>
</html>
`
