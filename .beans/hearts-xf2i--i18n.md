---
# hearts-xf2i
title: i18n
status: completed
type: epic
priority: normal
created_at: 2026-03-15T16:38:51Z
updated_at: 2026-03-29T14:46:10Z
---

## Vision
All user-facing text in the Hearts UI is translatable. Users can switch between English and German, with the choice persisted in LocalStorage (and later in a user account). The experience is flash-free — the correct language is always known before first paint.

## Context
The UI is English-only. The expected user base skews German-speaking, so German support is a priority. There is no i18n infrastructure yet.

## Approach
**Client-side translation with server-inlined strings.** The server reads locale JSON files (`locales/en.json`, `locales/de.json`) at startup. On each page request, it picks the active locale (LocalStorage cookie > `Accept-Language` header > `en` fallback) and inlines the matching translation object as a `<script>` in `<head>` before any rendering. A small `t(key)` JS function performs lookups against this global object. Templates keep English as the literal baseline; `t()` calls replace them at runtime for non-English locales. Language switching writes to LocalStorage and reloads.

**Detection priority:** explicit user setting (LocalStorage) > `Accept-Language` header > `en`.

**Not pursued:** IP geolocation (unreliable, Accept-Language is a stronger signal), loading translations as separate async files (unnecessary complexity for two small locales), server-side template translation (conflicts with WebSocket multiplayer where players may have different locales).

## Out of Scope
- Database-backed user accounts / persisting language preference server-side
- IP geolocation
- Locales beyond en/de
- Date/number formatting differences
- Translating log output or server-side error messages
