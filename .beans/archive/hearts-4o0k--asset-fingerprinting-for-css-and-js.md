---
# hearts-4o0k
title: Asset fingerprinting for CSS and JS
status: completed
type: feature
priority: normal
created_at: 2026-03-25T06:55:06Z
updated_at: 2026-03-25T08:09:47Z
parent: hearts-e5b4
---

Content-hash CSS/JS URLs for aggressive caching; bypass in dev mode


## Context

CSS and JS files change with each build but are currently served at fixed paths (`/assets/styles.css`, `/assets/js/...`) with no caching headers. Browsers must re-download them every time. Unlike card SVGs, these files can't simply get long-lived cache headers — stale CSS/JS after a redeploy would break the UI.

## Higher Goal

Enable aggressive browser caching for CSS and JS while guaranteeing clients always get the latest version after a deploy. This is the core of the web caching epic (hearts-e5b4) and delivers the biggest bandwidth savings for frequently-visited pages.

## Acceptance Criteria

- [x] CSS and JS are served at content-hashed URLs (e.g. `/assets/styles.a1b2c3.css`)
- [x] Fingerprinted asset responses include `Cache-Control: public, max-age=31536000, immutable`
- [x] HTML templates reference fingerprinted URLs (not hardcoded paths)
- [x] JS module imports between files use fingerprinted paths or are otherwise cache-safe
- [x] Old (non-fingerprinted) asset paths return 404 or redirect
- [x] In dev mode (`mise dev`), assets are served at plain paths without fingerprinting or cache headers, so hot-reload works without stale-cache issues
- [x] Test coverage for fingerprinted URL generation and serving
- [x] Documentation updated if architecture or asset pipeline changes

## Out of Scope

- Card SVG fingerprinting (covered by immutable caching in hearts-tnu6)
- HTML page caching (separate ticket)
- Build-time fingerprinting or external tooling — hashes should be computed at startup from `embed.FS`

## Summary of Changes

- Added `internal/webui/fingerprint.go`: computes SHA256 content hashes for CSS/JS at startup, builds URL mapping, rewrites JS import paths to use fingerprinted filenames
- Modified `internal/webui/server.go`: in production mode, HTML templates are rewritten with fingerprinted asset URLs; fingerprinted assets served with immutable cache headers; plain CSS/JS paths 404. In dev mode, assets served at plain paths with no cache headers
- Updated tests: `TestDevModeServesPlainAssetPaths`, `TestFingerprintedAssetURLsAndCaching`, `TestFingerprintedJSImportsRewritten`
- Updated `CLAUDE.md` and `architecture.md` with asset fingerprinting documentation
