package webui

import (
	"net/http"
	"testing"
)

func TestCheckOriginFunc(t *testing.T) {
	tests := []struct {
		name    string
		devMode bool
		origin  string
		host    string
		want    bool
	}{
		{"dev mode allows any origin", true, "http://evil.com", "localhost:8080", true},
		{"prod allows matching origin", false, "http://localhost:8080", "localhost:8080", true},
		{"prod allows https matching origin", false, "https://example.com", "example.com", true},
		{"prod rejects mismatched origin", false, "http://evil.com", "localhost:8080", false},
		{"prod allows missing origin", false, "", "localhost:8080", true},
		{"prod allows origin with path (host still matches)", false, "https://example.com/foo", "example.com", true},
		{"prod handles uppercase scheme", false, "HTTPS://example.com", "example.com", true},
		{"prod matches origin with explicit port", false, "https://example.com:8080", "example.com:8080", true},
		{"prod rejects port mismatch", false, "https://example.com:9090", "example.com:8080", false},
		{"prod rejects bare string (no scheme)", false, "not-a-url", "example.com", false},
		{"prod rejects null origin (sandboxed iframe)", false, "null", "example.com", false},
		{"prod rejects userinfo in origin", false, "http://evil.com@example.com", "example.com", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := checkOriginFunc(tt.devMode)
			r := &http.Request{
				Host:   tt.host,
				Header: http.Header{},
			}
			if tt.origin != "" {
				r.Header.Set("Origin", tt.origin)
			}
			if got := check(r); got != tt.want {
				t.Errorf("checkOriginFunc(%v)(origin=%q, host=%q) = %v, want %v",
					tt.devMode, tt.origin, tt.host, got, tt.want)
			}
		})
	}
}
