package webui

import (
	"time"

	"github.com/JHK/hearts/internal/session"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// registerAPIRoutes mounts REST API handlers under /api.
// Table listing and creation have moved to the lobby WebSocket;
// only dev/debug routes remain.
func registerAPIRoutes(r chi.Router, cfg Config, manager *session.Manager) {
	if !cfg.Dev {
		return
	}

	r.Route("/api", func(api chi.Router) {
		api.Use(requestLoggingMiddleware)
		api.Use(middleware.Timeout(10 * time.Second))

		registerDevAPIRoutes(api, manager)
	})
}
