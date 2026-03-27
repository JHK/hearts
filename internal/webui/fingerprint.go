package webui

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
)

// fingerprintedAssets maps original asset paths to fingerprinted URLs and
// pre-processed content. Built once at startup from embed.FS.
type fingerprintedAssets struct {
	// byFingerprintedURL maps fingerprinted URL → asset content (with rewritten imports).
	byFingerprintedURL map[string][]byte

	// urlMapping maps original URL → fingerprinted URL, used to rewrite HTML.
	urlMapping map[string]string
}

// buildFingerprintedAssets walks the embedded filesystem for CSS and JS files,
// computes content-based hashes, rewrites JS import paths to use fingerprinted
// filenames, and returns the result. The caller uses urlMapping to rewrite HTML
// and byFingerprintedURL to serve assets at their hashed paths.
func buildFingerprintedAssets(assets fs.FS) (*fingerprintedAssets, error) {
	fa := &fingerprintedAssets{
		byFingerprintedURL: make(map[string][]byte),
		urlMapping:         make(map[string]string),
	}

	// Collect CSS and JS asset paths.
	var paths []string
	if err := fs.WalkDir(assets, "assets/js", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(p, ".js") {
			paths = append(paths, p)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("walk js assets: %w", err)
	}
	paths = append(paths, "assets/styles.css")

	// Read all files and compute hashes.
	contents := make(map[string][]byte, len(paths))
	for _, p := range paths {
		data, err := fs.ReadFile(assets, p)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", p, err)
		}
		contents[p] = data
	}

	// Build URL mapping: /assets/foo.js → /assets/foo.HASH.js
	for _, p := range paths {
		hash := shortHash(contents[p])
		origURL := "/" + p
		ext := path.Ext(p)
		base := strings.TrimSuffix(p, ext)
		fpURL := "/" + base + "." + hash + ext
		fa.urlMapping[origURL] = fpURL
	}

	// Rewrite JS import paths and store content keyed by fingerprinted URL.
	jsImportRe := regexp.MustCompile(`(from\s+['"])(\.\.?/[^'"]+\.js)(['"])`)

	for _, p := range paths {
		data := contents[p]
		dir := path.Dir(p)

		if strings.HasSuffix(p, ".js") {
			data = jsImportRe.ReplaceAllFunc(data, func(match []byte) []byte {
				parts := jsImportRe.FindSubmatch(match)
				relPath := string(parts[2])
				absURL := "/" + path.Join(dir, relPath)
				fpURL, ok := fa.urlMapping[absURL]
				if !ok {
					return match
				}
				fpDir := path.Dir(fa.urlMapping["/"+p])
				rel, err := relativeImportPath(fpDir, fpURL)
				if err != nil {
					return match
				}
				var result []byte
				result = append(result, parts[1]...)
				result = append(result, rel...)
				result = append(result, parts[3]...)
				return result
			})
		}

		fpURL := fa.urlMapping["/"+p]
		fa.byFingerprintedURL[fpURL] = data
	}

	return fa, nil
}

// registerFingerprintedAssetHandlers registers handlers that serve CSS and JS
// at their fingerprinted URLs with immutable cache headers. Plain (non-fingerprinted)
// CSS/JS requests are not registered and will 404.
func registerFingerprintedAssetHandlers(router chi.Router, fa *fingerprintedAssets) {
	router.Get("/assets/js/*", func(w http.ResponseWriter, r *http.Request) {
		content, ok := fa.byFingerprintedURL[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		_, _ = w.Write(content)
	})

	cssURL := fa.urlMapping["/assets/styles.css"]
	cssContent := fa.byFingerprintedURL[cssURL]

	router.Get(cssURL, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		_, _ = w.Write(cssContent)
	})
}

// relativeImportPath computes the relative path from fromDir to toPath,
// ensuring it starts with "./" or "../" for use in ES module imports.
func relativeImportPath(fromDir, toPath string) (string, error) {
	fromDir = path.Clean(fromDir)
	toPath = path.Clean(toPath)
	toDir := path.Dir(toPath)
	toFile := path.Base(toPath)

	// Build relative directory traversal.
	fromParts := strings.Split(strings.TrimPrefix(fromDir, "/"), "/")
	toParts := strings.Split(strings.TrimPrefix(toDir, "/"), "/")

	// Find common prefix length.
	common := 0
	for common < len(fromParts) && common < len(toParts) && fromParts[common] == toParts[common] {
		common++
	}

	ups := len(fromParts) - common
	var rel string
	if ups == 0 {
		rel = "./"
	} else {
		rel = strings.Repeat("../", ups)
	}
	if remaining := toParts[common:]; len(remaining) > 0 && remaining[0] != "" {
		rel += strings.Join(remaining, "/") + "/"
	}
	return rel + toFile, nil
}

func shortHash(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h[:4])
}
