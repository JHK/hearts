---
# hearts-71qw
title: Research web framework for HTTP layer
status: todo
type: task
priority: normal
tags:
    - backend
    - frontend
created_at: 2026-03-25T08:48:02Z
updated_at: 2026-03-25T08:48:19Z
---

Evaluate net/http 1.22+, Chi, Echo, and Fiber to replace hand-rolled HTTP plumbing with less code and better abstractions

## Context

The web layer (`internal/webui/server.go`) is built on raw `http.ServeMux` with entirely hand-rolled
HTTP plumbing: per-handler cache headers, ETag computation, content-type setting, path traversal
checks, and static file serving. There's no server-side templating — HTML pages are fully static,
with asset URLs hardcoded rather than injected (e.g. fingerprinted paths).

This works but results in repetitive, low-level code that a framework or structured approach would
handle by convention. The goal is less code, better abstractions, and a more standardized approach —
not necessarily 100% behavioral equivalence (e.g. response codes may differ).

## Options to Evaluate

1. **Go 1.22+ `net/http`** — enhanced routing (`GET /path/{id}`), evaluate how much boilerplate
   it eliminates vs the current approach when combined with middleware patterns
2. **Chi** — lightweight router, composable middleware, stays close to `net/http` interfaces
3. **Echo** — more opinionated, built-in middleware (gzip, CORS, static file serving, templating)
4. **Fiber** — fasthttp-based, different API surface, strongest opinions

For each option, evaluate:
- **Routing & middleware**: how much per-handler boilerplate disappears (cache headers, content types,
  path validation)
- **Static asset serving**: built-in support for embedded FS, content hashing / fingerprinting,
  immutable cache headers
- **Templating**: injecting asset hashes into HTML, potential for server-rendered fragments later
- **WebSocket support**: native or via gorilla/websocket compatibility (nice-to-have, not required)
- **Compatibility with `embed.FS`**: all assets are embedded into the binary
- **Community health & maintenance**: activity, Go version support, API stability

## Acceptance Criteria

- [ ] Each option has a proof of concept covering: routing, static asset serving with cache headers,
      and a templated HTML page with an injected asset hash
- [ ] Options are compared in a summary (bean comment or doc) with trade-offs documented
- [ ] A follow-up feature ticket is created based on the chosen approach
- [ ] The "Web Caching" epic (`hearts-e5b4`) is updated to reflect the decision

## Out of Scope

- Actually migrating the codebase — this ticket produces a decision, not an implementation
- Replacing the WebSocket message protocol or game state flow
- Frontend framework evaluation (React, Vue, etc.) — client-side JS stays as-is
- Performance benchmarking under load (this is a LAN game, not a high-traffic service)
