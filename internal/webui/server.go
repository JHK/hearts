package webui

import (
	"context"
	"embed"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/JHK/hearts/internal/session"
	"github.com/JHK/hearts/internal/webui/tracker"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//go:embed assets/index.html assets/table.html assets/styles.css assets/cards/*.svg assets/js assets/icon.svg assets/favicon.ico assets/apple-touch-icon.png
var assetsFS embed.FS

type Config struct {
	Addr string
	Dev  bool
}

func Run(cfg Config) error {
	if strings.TrimSpace(cfg.Addr) == "" {
		cfg.Addr = ":8080"
	}

	manager := session.NewManager()
	connTracker := tracker.NewConnTracker()

	handler, err := NewHandler(cfg, manager, connTracker)
	if err != nil {
		return err
	}

	srv := &http.Server{
		Addr:    cfg.Addr,
		Handler: handler,
	}

	// Start server in a goroutine so we can block on signal handling.
	errCh := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		slog.Info("shutdown signal received", "signal", sig)
	case err := <-errCh:
		return err
	}
	signal.Stop(quit)

	const shutdownTimeout = 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Stop accepting new connections.
	slog.Info("shutting down HTTP server", "timeout", shutdownTimeout)
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("HTTP shutdown error", "err", err)
	}

	// Interrupt active WebSocket read loops and wait for handlers to exit.
	slog.Info("draining WebSocket connections")
	connTracker.Shutdown()
	connTracker.Wait(ctx)

	slog.Info("closing lobby hub")
	handler.Close()

	slog.Info("closing all tables")
	manager.Close()

	slog.Info("shutdown complete")
	return nil
}

// Handler wraps the HTTP handler and owns background actors that need
// graceful shutdown.
type Handler struct {
	http.Handler
	lobby          *lobbyHub
	presence       *tracker.HumanPresence
	playerPresence *tracker.PlayerPresence
}

// Close shuts down background actors owned by the handler.
func (h *Handler) Close() {
	h.lobby.Shutdown()
	h.presence.Shutdown()
	h.playerPresence.Shutdown()
}

func NewHandler(cfg Config, manager *session.Manager, ct *tracker.ConnTracker) (*Handler, error) {
	if manager == nil {
		manager = session.NewManager()
	}
	if ct == nil {
		ct = tracker.NewConnTracker()
	}

	r := chi.NewRouter()
	r.Use(middleware.Recoverer)

	presence := tracker.NewHumanPresence()
	playerPresence := tracker.NewPlayerPresence()
	lobby := newLobbyHub()

	// Page handlers, static assets, API — with gzip compression
	compressed := r.With(middleware.Compress(5))
	if err := registerPageRoutes(compressed, cfg, manager); err != nil {
		return nil, err
	}
	registerAPIRoutes(compressed, cfg, manager)

	// WebSocket endpoints — no compression (upgrade must pass through unmodified)
	registerWSRoutes(r, manager, lobby, presence, playerPresence, ct)

	return &Handler{
		Handler:        r,
		lobby:          lobby,
		presence:       presence,
		playerPresence: playerPresence,
	}, nil
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
