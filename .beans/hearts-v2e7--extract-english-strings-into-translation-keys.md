---
# hearts-v2e7
title: Extract English strings into translation keys
status: todo
type: task
priority: normal
created_at: 2026-03-29T12:25:15Z
updated_at: 2026-03-29T12:25:35Z
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
- [ ] `t()` supports variable interpolation (e.g. `t("game.trick_won_by", {name: "Alice"})` → `"Alice won the trick"`)
- [ ] `t()` supports basic pluralization via count (e.g. `t("game.points", {count: 3})` → `"3 points"`)
- [ ] All user-facing strings in HTML templates are replaced with `t()` calls
- [ ] All user-facing strings in JS (toasts, dynamic messages, WebSocket event rendering) are replaced with `t()` calls
- [ ] `locales/en.json` is complete — covers every key used across the app
- [ ] Keys follow a consistent naming convention (e.g. `page.section.element`)
- [ ] No visible regressions — app looks and behaves identically with `en` locale
- [ ] Documentation updated if the key naming convention or `t()` API changed

## Design Decisions
- **No ICU MessageFormat.** With only two locales and no complex grammatical cases, a simple interpolation (`{{name}}`) and count-based pluralization (`one`/`other`) is sufficient. Revisit if more locales are added.

## Out of Scope
- German translations (next ticket)
- Language switcher UI
- Changing any visual design or layout
