package webui

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/JHK/hearts/internal/session"
	"github.com/go-chi/chi/v5"
)

// registerDevRoutes mounts dev-only non-API routes (scripts, assets).
func registerDevRoutes(r chi.Router) {
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

// registerDevAPIRoutes mounts dev-only API routes (debug endpoints).
func registerDevAPIRoutes(api chi.Router, manager *session.Manager) {
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
