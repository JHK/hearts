---
# hearts-lpy9
title: i18n infrastructure & locale detection
status: todo
type: task
priority: normal
created_at: 2026-03-29T12:19:53Z
updated_at: 2026-03-29T12:20:14Z
parent: hearts-xf2i
---

Set up translation pipeline: locale JSON files, server-side inlining, Accept-Language detection, LocalStorage override, t() function

**Note:** Read the parent epic (`hearts-xf2i`) for the full approach and architectural decisions before starting.

## Context
There is no i18n infrastructure. This ticket lays the groundwork that all subsequent i18n tickets build on.

## Higher Goal
Enable client-side translation with zero flash-of-untranslated-content, as defined in the i18n epic.

## Acceptance Criteria
- [ ] `locales/en.json` exists with a small set of sample keys (enough to prove the pipeline works, e.g. 5–10 strings from one page)
- [ ] Server reads locale JSON files at startup and makes them available to templates
- [ ] Locale detection: server reads `Accept-Language` header, extracts best match (`en` or `de`), passes it into the template context
- [ ] A `<script>` block in `<head>` (before body renders): reads LocalStorage for an explicit override, falls back to the server-provided Accept-Language value, falls back to `en`, then sets `window.__i18n` to the inlined strings for that locale
- [ ] A `t(key)` JS function is available globally, returns the translated string or the key itself as fallback
- [ ] At least one page has a few strings replaced with `t()` calls as proof the pipeline works end-to-end
- [ ] Inlined translation object is fingerprint-safe (no cache staleness issues)
- [ ] Documentation updated (CLAUDE.md, architecture.md as needed)

## Out of Scope
- Full string extraction (that's the next ticket)
- German translations
- Language switcher UI
- Pluralization / interpolation beyond simple string lookup
