package webui

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

//go:embed locales
var localesFS embed.FS

// i18nData holds all loaded locale translations, built once at startup.
type i18nData struct {
	// allJSON is the JSON-encoded object of all locales: {"en":{...},"de":{...}}
	allJSON string
	// locales is the sorted list of supported locale codes.
	locales []string
}

// loadI18n reads all JSON files from the embedded locales directory
// and returns the combined translation data.
func loadI18n() (*i18nData, error) {
	entries, err := localesFS.ReadDir("locales")
	if err != nil {
		return nil, fmt.Errorf("read locales dir: %w", err)
	}

	all := make(map[string]json.RawMessage)
	var locales []string

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		locale := strings.TrimSuffix(entry.Name(), ".json")
		data, err := localesFS.ReadFile("locales/" + entry.Name())
		if err != nil {
			return nil, fmt.Errorf("read locale %s: %w", locale, err)
		}
		// Validate JSON.
		var check json.RawMessage
		if err := json.Unmarshal(data, &check); err != nil {
			return nil, fmt.Errorf("invalid JSON in %s: %w", entry.Name(), err)
		}
		all[locale] = data
		locales = append(locales, locale)
	}

	sort.Strings(locales)

	jsonBytes, err := json.Marshal(all)
	if err != nil {
		return nil, fmt.Errorf("marshal i18n data: %w", err)
	}

	if _, ok := all["en"]; !ok {
		return nil, fmt.Errorf("locales/en.json is required but missing")
	}

	return &i18nData{
		allJSON: string(jsonBytes),
		locales: locales,
	}, nil
}

// localeSet is a pre-built set of supported locale codes, constructed once
// at startup to avoid per-request allocation in detectLocale.
type localeSet map[string]bool

func newLocaleSet(locales []string) localeSet {
	s := make(localeSet, len(locales))
	for _, l := range locales {
		s[l] = true
	}
	return s
}

// detectLocale parses the Accept-Language header and returns the best matching
// supported locale, falling back to "en".
func detectLocale(r *http.Request, supported localeSet) string {
	accept := r.Header.Get("Accept-Language")
	if accept == "" {
		return "en"
	}

	// Parse Accept-Language entries sorted by quality (highest first).
	// Format: en-US,en;q=0.9,de;q=0.8
	type langQ struct {
		lang string
		q    float64
	}
	var entries []langQ

	for part := range strings.SplitSeq(accept, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		lang := part
		q := 1.0
		if before, after, ok := strings.Cut(part, ";"); ok {
			lang = before
			qPart := strings.TrimSpace(after)
			if strings.HasPrefix(qPart, "q=") {
				fmt.Sscanf(qPart[2:], "%f", &q)
			}
		}
		entries = append(entries, langQ{lang: strings.TrimSpace(lang), q: q})
	}

	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].q > entries[j].q
	})

	for _, e := range entries {
		code := strings.ToLower(e.lang)
		if supported[code] {
			return code
		}
		// Try base language: "en-US" → "en"
		if idx := strings.IndexByte(code, '-'); idx > 0 {
			if supported[code[:idx]] {
				return code[:idx]
			}
		}
	}

	return "en"
}
