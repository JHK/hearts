package webui

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/JHK/hearts/internal/session"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// registerAPIRoutes mounts all REST API handlers under /api.
func registerAPIRoutes(r chi.Router, cfg Config, manager *session.Manager) {
	r.Route("/api", func(api chi.Router) {
		api.Use(requestLoggingMiddleware)
		api.Use(middleware.Timeout(10 * time.Second))

		api.Get("/tables", func(w http.ResponseWriter, r *http.Request) {
			handleTablesAPI(manager, w, r)
		})
		api.Post("/tables", func(w http.ResponseWriter, r *http.Request) {
			handleTablesAPI(manager, w, r)
		})

		if cfg.Dev {
			registerDevAPIRoutes(api, manager)
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
			writeJSON(w, map[string]any{"type": "error", "error": "invalid JSON"})
			return
		}

		runtime, created, err := manager.Create(req.TableID)
		if err != nil {
			writeJSON(w, map[string]any{"type": "error", "error": err.Error()})
			return
		}

		writeJSON(w, map[string]any{"table_id": runtime.ID(), "created": created})
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
