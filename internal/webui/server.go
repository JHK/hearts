package webui

import (
	"bytes"
	"context"
	"crypto/sha256"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/JHK/hearts/internal/session"
	"github.com/JHK/hearts/internal/webui/tracker"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
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

	slog.Info("closing all tables")
	manager.Close()

	slog.Info("shutdown complete")
	return nil
}

func NewHandler(cfg Config, manager *session.Manager, ct *tracker.ConnTracker) (http.Handler, error) {
	if manager == nil {
		manager = session.NewManager()
	}
	if ct == nil {
		ct = tracker.NewConnTracker()
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

	presence := tracker.NewHumanPresence()
	playerPresence := tracker.NewPlayerPresence()
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
	registerAPIRoutes(r, manager)

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
	registerWSRoutes(r, manager, lobby, presence, playerPresence, ct)

	return r, nil
}

// immutableCacheMiddleware sets Cache-Control headers for immutable static assets.
func immutableCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
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

