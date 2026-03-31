package webui

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDetectLocale(t *testing.T) {
	supported := newLocaleSet([]string{"de", "en"})

	tests := []struct {
		name   string
		accept string
		want   string
	}{
		{"empty header", "", "en"},
		{"exact en", "en", "en"},
		{"exact de", "de", "de"},
		{"en-US maps to en", "en-US", "en"},
		{"de-DE maps to de", "de-DE,en;q=0.9", "de"},
		{"quality ordering", "en;q=0.8,de;q=0.9", "de"},
		{"unsupported falls back to en", "fr,ja", "en"},
		{"mixed with unsupported", "fr,de;q=0.5", "de"},
		{"wildcard ignored", "*", "en"},
		{"case insensitive", "DE-de", "de"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := http.NewRequest("GET", "/", nil)
			if tt.accept != "" {
				r.Header.Set("Accept-Language", tt.accept)
			}
			got := detectLocale(r, supported)
			if got != tt.want {
				t.Errorf("detectLocale(%q) = %q, want %q", tt.accept, got, tt.want)
			}
		})
	}
}

func TestLoadI18n(t *testing.T) {
	i18n, err := loadI18n()
	if err != nil {
		t.Fatalf("loadI18n() error: %v", err)
	}
	if len(i18n.locales) == 0 {
		t.Fatal("expected at least one locale")
	}
	if i18n.allJSON == "" {
		t.Fatal("expected non-empty allJSON")
	}
	// en must be present
	found := false
	for _, l := range i18n.locales {
		if l == "en" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'en' in locales")
	}
}

func TestLocalizedPageServe(t *testing.T) {
	tmpl := template.Must(template.New("test").Parse(`locale={{.ServerLocale}} i18n={{.I18nJSON}}`))
	data := templateData{
		I18nJSON: `{"en":{"hello":"Hello"},"de":{"hello":"Hallo"}}`,
	}
	page := buildLocalizedPage(tmpl, data, []string{"en", "de"})

	tests := []struct {
		name       string
		accept     string
		wantLocale string
	}{
		{"default to en", "", "en"},
		{"german speaker", "de-DE,de;q=0.9,en;q=0.8", "de"},
		{"english speaker", "en-US,en;q=0.9", "en"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			if tt.accept != "" {
				r.Header.Set("Accept-Language", tt.accept)
			}
			w := httptest.NewRecorder()
			page.serve(w, r)

			resp := w.Result()

			// Verify Vary header
			if got := resp.Header.Get("Vary"); got != "Accept-Language" {
				t.Errorf("Vary = %q, want %q", got, "Accept-Language")
			}

			// Verify ETag is present
			if resp.Header.Get("ETag") == "" {
				t.Error("expected ETag header")
			}

			// Verify correct locale variant was served
			body := w.Body.String()
			want := "locale=" + tt.wantLocale
			if !strings.Contains(body, want) {
				t.Errorf("body does not contain %q:\n%s", want, body)
			}
		})
	}

	// Verify different locales produce different ETags
	enReq := httptest.NewRequest("GET", "/", nil)
	enReq.Header.Set("Accept-Language", "en")
	enW := httptest.NewRecorder()
	page.serve(enW, enReq)

	deReq := httptest.NewRequest("GET", "/", nil)
	deReq.Header.Set("Accept-Language", "de")
	deW := httptest.NewRecorder()
	page.serve(deW, deReq)

	enETag := enW.Result().Header.Get("ETag")
	deETag := deW.Result().Header.Get("ETag")
	if enETag == deETag {
		t.Errorf("en and de ETags should differ, both are %q", enETag)
	}

	// Verify 304 on matching ETag
	r304 := httptest.NewRequest("GET", "/", nil)
	r304.Header.Set("Accept-Language", "en")
	r304.Header.Set("If-None-Match", enETag)
	w304 := httptest.NewRecorder()
	page.serve(w304, r304)
	if w304.Code != http.StatusNotModified {
		t.Errorf("expected 304, got %d", w304.Code)
	}
}
