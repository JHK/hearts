---
# hearts-lpy9
title: i18n infrastructure & locale detection
status: completed
type: task
priority: normal
created_at: 2026-03-29T12:19:53Z
updated_at: 2026-03-29T12:42:33Z
parent: hearts-xf2i
---

Set up translation pipeline: locale JSON files, server-side inlining, Accept-Language detection, LocalStorage override, t() function

**Note:** Read the parent epic (`hearts-xf2i`) for the full approach and architectural decisions before starting.

## Context
There is no i18n infrastructure. This ticket lays the groundwork that all subsequent i18n tickets build on.

## Higher Goal
Enable client-side translation with zero flash-of-untranslated-content, as defined in the i18n epic.

## Acceptance Criteria
- [x] `locales/en.json` exists with a small set of sample keys (enough to prove the pipeline works, e.g. 5–10 strings from one page)
- [x] Server reads locale JSON files at startup and makes them available to templates
- [x] Locale detection: server reads `Accept-Language` header, extracts best match (`en` or `de`), passes it into the template context
- [x] A `<script>` block in `<head>` (before body renders): reads LocalStorage for an explicit override, falls back to the server-provided Accept-Language value, falls back to `en`, then sets `window.__i18n` to the inlined strings for that locale
- [x] A `t(key)` JS function is available globally, returns the translated string or the key itself as fallback
- [x] At least one page has a few strings replaced with `t()` calls as proof the pipeline works end-to-end
- [x] Inlined translation object is fingerprint-safe (no cache staleness issues)
- [x] Documentation updated (CLAUDE.md, architecture.md as needed)

## Out of Scope
- Full string extraction (that's the next ticket)
- German translations
- Language switcher UI
- Pluralization / interpolation beyond simple string lookup


## Summary of Changes

Added i18n infrastructure:
- `internal/webui/locales/en.json` — 10 sample translation keys from the lobby page
- `internal/webui/i18n.go` — loads embedded locale JSON files at startup, parses Accept-Language headers with quality-value ordering
- `internal/webui/i18n_test.go` — tests for locale detection and loading
- HTML templates get an inline `<script>` in `<head>` that inlines all locale data, detects locale (LocalStorage > Accept-Language > en), and exposes a global `t(key)` function
- Pages are pre-rendered per locale with `Vary: Accept-Language` for correct caching; no per-request template rendering needed
- Lobby JS uses `t()` for badge text, join button, and create-table aria label as proof of end-to-end pipeline
- Updated CLAUDE.md and architecture.md
