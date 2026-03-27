---
# hearts-zsjb
title: Remove events panel from table UI; log to console in dev mode
status: todo
type: task
priority: normal
tags:
    - frontend
created_at: 2026-03-27T12:12:00Z
updated_at: 2026-03-27T12:12:26Z
parent: hearts-dfll
---

## Context

The table UI has a collapsible `<details>` section (`#eventsSection` / `#logs`) that displays a timestamped feed of game events. It's gated behind the `?events=true` query param. This is a developer tool that doesn't belong in the DOM — `console.log` is the right place for it.

## Higher Goal

Simplify the table UI and reduce shipped markup. Dev-only tooling should use the established `-dev` flag pattern (like `debugBot()`), not a query param.

## Acceptance Criteria

- [ ] `#eventsSection` / `#logs` HTML removed from `table.html`
- [ ] DOM references (`eventsSectionEl`, `logsEl`) removed from `dom.js`
- [ ] `?events=true` query param mechanism removed
- [ ] In dev mode (`-dev` flag), event messages are logged to `console.log` — following the same injection pattern as `routes_dev.go` uses for `debugBot()`
- [ ] No event logging in production mode

## Out of Scope

- Changing the `debugBot()` mechanism or other dev tooling
- Adding any new dev-tools UI
