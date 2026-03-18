---
# hearts-jjfg
title: Auto-reload dev server on changes to Go, CSS, JS, and HTML files
status: todo
type: task
priority: normal
created_at: 2026-03-18T09:17:52Z
updated_at: 2026-03-18T09:17:59Z
---

Rebuild and restart the dev server automatically on file changes, following the beans project pattern

## Context
`mise run dev` starts the server once and requires a manual restart for every
change to Go source, embedded HTML/JS, or Tailwind CSS files. This adds
unnecessary friction during development.

## Higher Goal
Tighter feedback loop during development — the server reflects changes
automatically, similar to how the beans project handles it.

## Approach
Mirror the beans pattern from `hmans/beans/mise.toml`:

- Annotate a `serve` task with `sources` covering all watched paths
- Replace the `dev` run command with `mise watch serve --restart`

Since HTML, JS, and CSS are embedded via `//go:embed`, any change to them
requires a Go rebuild anyway. CSS additionally requires an npm build step
first, which can be modeled as a `depends` on the serve task.

## Acceptance Criteria
- [ ] `mise run dev` rebuilds and restarts the server automatically when
  any `.go` file changes
- [ ] CSS is rebuilt and the server is restarted when Tailwind source files
  change
- [ ] Embedded JS and HTML template changes also trigger a restart
- [ ] No manual server restart is needed during normal development

## Out of Scope
- Browser-side live reload / hot module replacement (no page auto-refresh)
- Watching files outside the project directory
