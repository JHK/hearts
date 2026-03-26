package webui

import (
	"encoding/json"
	"net/http"

	"github.com/JHK/hearts/internal/session"
	"github.com/go-chi/chi/v5"
)

// registerAPIRoutes mounts all REST API handlers under /api.
func registerAPIRoutes(r chi.Router, cfg Config, manager *session.Manager) {
	r.Route("/api", func(api chi.Router) {
		api.Get("/tables", func(w http.ResponseWriter, r *http.Request) {
			handleTablesAPI(manager, w, r)
		})
		api.Post("/tables", func(w http.ResponseWriter, r *http.Request) {
			handleTablesAPI(manager, w, r)
		})

		if cfg.Dev {
			api.Get("/debug/bots", func(w http.ResponseWriter, r *http.Request) {
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
