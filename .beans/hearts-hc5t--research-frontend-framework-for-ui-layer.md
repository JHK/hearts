---
# hearts-hc5t
title: Research frontend framework for UI layer
status: todo
type: task
priority: normal
tags:
    - frontend
created_at: 2026-03-25T08:52:53Z
updated_at: 2026-03-25T08:53:53Z
blocked_by:
    - hearts-71qw
---

Evaluate Svelte, Preact, Lit vs keeping vanilla ES6 — decide whether a compiled framework is worth the build complexity

## Context

The game UI is ~1000 lines of vanilla ES6 across two pages (lobby, game table), using manual DOM
manipulation driven by WebSocket state updates. There is no JS build step — scripts are embedded
directly via `go:embed`.

This is simple and works, but may not scale well if the UI grows in complexity. A compiled framework
like Svelte could provide reactive state → DOM binding, component structure, and scoped CSS without
adding runtime overhead. However, it introduces a build step that must integrate with the Go embed
pipeline.

This research should happen after the backend framework decision (`hearts-71qw`), since that choice
affects asset serving, fingerprinting, and templating — all of which constrain how a frontend build
would integrate.

## Options to Evaluate

1. **Keep vanilla ES6** — status quo, no build step, assess what pain points actually exist
2. **Svelte** — compiles to vanilla JS, no runtime, component model, scoped CSS
3. **Preact** — 3KB runtime, JSX, React-like API, lightweight alternative
4. **Lit** — web components standard, small runtime, no build step required (optional)

For each option, evaluate:
- **Build integration**: how it fits into the `go:embed` pipeline, impact on `mise run` commands
- **Developer experience**: HMR / dev server, how it interacts with the Go dev server
- **Bundle size**: runtime overhead vs current zero-dependency JS
- **WebSocket integration**: how naturally it handles push-driven state updates
- **Migration path**: can it be adopted incrementally (one page at a time) or is it all-or-nothing

## Acceptance Criteria

- [ ] Current JS pain points (if any) are documented — enumerate concrete problems, not hypotheticals
- [ ] Each option has a small prototype (e.g. the lobby page) showing build integration with `go:embed`
- [ ] Options are compared with trade-offs documented
- [ ] A follow-up ticket is created, or a decision to stay with vanilla JS is recorded

## Out of Scope

- Full migration of the game table UI — this ticket produces a decision only
- Backend framework choice (covered by `hearts-71qw`)
- CSS framework evaluation (Tailwind stays)
