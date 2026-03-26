package webui

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/JHK/hearts/internal/session"
	"github.com/go-chi/chi/v5"
)

// templateData holds the values injected into HTML templates.
type templateData struct {
	StylesURL    string
	ScriptURL    string
	ChartJSURL   string
	ExtraScripts template.HTML
}

// registerPageRoutes mounts HTML page handlers, static card/icon assets, and
// dev-only routes onto the router.
func registerPageRoutes(r chi.Router, cfg Config, manager *session.Manager) error {
	indexTmpl, err := template.New("index").Parse(string(mustReadAsset("assets/index.html")))
	if err != nil {
		return fmt.Errorf("parse index template: %w", err)
	}
	tableTmpl, err := template.New("table").Parse(string(mustReadAsset("assets/table.html")))
	if err != nil {
		return fmt.Errorf("parse table template: %w", err)
	}

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
			return fmt.Errorf("build fingerprinted assets: %w", fpErr)
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

	// HTML pages
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

	// Static card assets — immutable cache
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
			return fmt.Errorf("read embedded %s: %w", f.asset, err)
		}
		contentType := f.contentType
		content := data
		r.With(immutableCacheMiddleware).Get(f.path, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", contentType)
			_, _ = w.Write(content)
		})
	}

	// Dev-only routes
	if cfg.Dev {
		registerDevRoutes(r)
	}

	return nil
}

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

// immutableCacheMiddleware sets Cache-Control headers for immutable static assets.
func immutableCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
}
