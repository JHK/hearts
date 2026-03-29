---
# hearts-v2e7
title: Extract English strings into translation keys
status: completed
type: task
priority: normal
created_at: 2026-03-29T12:25:15Z
updated_at: 2026-03-29T13:41:10Z
parent: hearts-xf2i
blocked_by:
    - hearts-lpy9
---

Replace all hardcoded English strings with t() calls, add interpolation and pluralization support

**Note:** Read the parent epic (`hearts-xf2i`) for the full approach and architectural decisions before starting.

## Context
After `hearts-lpy9`, the i18n pipeline works end-to-end for a handful of sample strings. This ticket completes the extraction — every user-facing string goes through `t()`.

## Higher Goal
Make the entire UI translatable by ensuring no English literals remain hardcoded in templates or JS.

## Acceptance Criteria
- [x] `t()` supports variable interpolation (e.g. `t("game.trick_won_by", {name: "Alice"})` → `"Alice won the trick"`)
- [x] `t()` supports basic pluralization via count (e.g. `t("game.points", {count: 3})` → `"3 points"`)
- [x] All user-facing strings in HTML templates are replaced with `t()` calls
- [x] All user-facing strings in JS (toasts, dynamic messages, WebSocket event rendering) are replaced with `t()` calls
- [x] `locales/en.json` is complete — covers every key used across the app
- [x] Keys follow a consistent naming convention (e.g. `page.section.element`)
- [x] No visible regressions — app looks and behaves identically with `en` locale
- [x] Documentation updated if the key naming convention or `t()` API changed

## Design Decisions
- **No ICU MessageFormat.** With only two locales and no complex grammatical cases, a simple interpolation (`{{name}}`) and count-based pluralization (`one`/`other`) is sufficient. Revisit if more locales are added.

## Out of Scope
- German translations (next ticket)
- Language switcher UI
- Changing any visual design or layout

## Summary of Changes

- Enhanced `t(key, params)` with `{{var}}` interpolation and `{one, other}` pluralization
- Added `data-i18n` / `data-i18n-*` attribute system for static HTML translation on DOMContentLoaded
- Extracted all hardcoded English strings from HTML templates (index.html, table.html, partials) and JS files (lobby/main.js, table/main.js, table/render.js)
- Built complete `en.json` with 40+ keys following `section.element` naming convention
- Updated CLAUDE.md with new `t()` API docs and `data-i18n` attribute conventions
- All tests pass
