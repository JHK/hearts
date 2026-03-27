---
# hearts-zsjb
title: Remove events panel from table UI; log to console in dev mode
status: completed
type: task
priority: normal
tags:
    - frontend
created_at: 2026-03-27T12:12:00Z
updated_at: 2026-03-27T12:16:39Z
parent: hearts-dfll
---

## Context

The table UI has a collapsible `<details>` section (`#eventsSection` / `#logs`) that displays a timestamped feed of game events. It's gated behind the `?events=true` query param. This is a developer tool that doesn't belong in the DOM — `console.log` is the right place for it.

## Higher Goal

Simplify the table UI and reduce shipped markup. Dev-only tooling should use the established `-dev` flag pattern (like `debugBot()`), not a query param.

## Acceptance Criteria

- [x] `#eventsSection` / `#logs` HTML removed from `table.html`
- [x] DOM references (`eventsSectionEl`, `logsEl`) removed from `dom.js`
- [x] `?events=true` query param mechanism removed
- [x] In dev mode (`-dev` flag), event messages are logged to `console.log` — following the same injection pattern as `routes_dev.go` uses for `debugBot()`
- [x] No event logging in production mode

## Out of Scope

- Changing the `debugBot()` mechanism or other dev tooling
- Adding any new dev-tools UI

## Summary of Changes

- Removed `#eventsSection` / `#logs` HTML from `table.html`
- Removed `eventsSectionEl`, `logsEl` DOM references and `eventsEnabled` param from `dom.js`
- Removed `?events=true` query param parsing from `main.js`
- Changed `log()` function to use `console.log` gated behind `window.__HEARTS_DEV__`
- Set `window.__HEARTS_DEV__ = true` in `dev.js` (injected by `routes_dev.go`), following the existing `debugBot()` pattern
