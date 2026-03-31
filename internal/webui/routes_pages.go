package webui

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"html/template"
	"net/http"
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
	I18nJSON     template.JS
	ServerLocale string
}

// localizedPage holds pre-rendered HTML variants, one per supported locale.
type localizedPage struct {
	variants  map[string]pageVariant
	localeSet localeSet
}

type pageVariant struct {
	content []byte
	etag    string
}

func buildLocalizedPage(tmpl *template.Template, data templateData, locales []string) *localizedPage {
	lp := &localizedPage{
		variants:  make(map[string]pageVariant, len(locales)),
		localeSet: newLocaleSet(locales),
	}
	for _, locale := range locales {
		data.ServerLocale = locale
		content := mustRenderTemplate(tmpl, data)
		lp.variants[locale] = pageVariant{
			content: content,
			etag:    contentETag(content),
		}
	}
	return lp
}

func (lp *localizedPage) serve(w http.ResponseWriter, r *http.Request) {
	locale := detectLocale(r, lp.localeSet)
	v, ok := lp.variants[locale]
	if !ok {
		v = lp.variants["en"]
	}
	w.Header().Set("Vary", "Accept-Language")
	serveHTMLWithETag(w, r, v.content, v.etag)
}

// registerPageRoutes mounts HTML page handlers, static card/icon assets, and
// dev-only routes onto the router.
func registerPageRoutes(r chi.Router, cfg Config, manager *session.Manager) error {
	i18n, err := loadI18n()
	if err != nil {
		return fmt.Errorf("load i18n: %w", err)
	}

	partials := string(mustReadAsset("assets/_i18n.html")) + string(mustReadAsset("assets/_settings_panel.html")) + string(mustReadAsset("assets/_page_header.html"))
	indexTmpl, err := template.New("index").Parse(partials + string(mustReadAsset("assets/index.html")))
	if err != nil {
		return fmt.Errorf("parse index template: %w", err)
	}
	tableTmpl, err := template.New("table").Parse(partials + string(mustReadAsset("assets/table.html")))
	if err != nil {
		return fmt.Errorf("parse table template: %w", err)
	}

	indexData := templateData{
		StylesURL: "/assets/styles.css",
		ScriptURL: "/assets/js/lobby/main.js",
		I18nJSON:  template.JS(i18n.allJSON),
	}
	tableData := templateData{
		StylesURL:  "/assets/styles.css",
		ScriptURL:  "/assets/js/table/main.js",
		ChartJSURL: "/assets/js/vendor/chart.umd.js",
		I18nJSON:   template.JS(i18n.allJSON),
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

	indexPage := buildLocalizedPage(indexTmpl, indexData, i18n.locales)
	tablePage := buildLocalizedPage(tableTmpl, tableData, i18n.locales)

	// HTML pages
	r.Group(func(pages chi.Router) {
		pages.Use(requestLoggingMiddleware)

		pages.Get("/", func(w http.ResponseWriter, r *http.Request) {
			indexPage.serve(w, r)
		})

		pages.Get("/table/{tableID}", func(w http.ResponseWriter, r *http.Request) {
			tableID := chi.URLParam(r, "tableID")
			if tableID == "" {
				http.NotFound(w, r)
				return
			}

			if _, ok := manager.Get(tableID); !ok {
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}

			tablePage.serve(w, r)
		})
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
