---
# hearts-1sxq
title: Migrate HTTP layer to Chi router
status: todo
type: feature
created_at: 2026-03-26T08:34:38Z
updated_at: 2026-03-26T08:34:38Z
---

Replace hand-rolled http.ServeMux routing with Chi for route groups and scoped middleware.

## Context

Research in hearts-71qw evaluated net/http 1.22+, Chi, Echo, and Fiber. Chi was chosen for its route groups with scoped middleware, full net/http compatibility, and zero migration friction. See docs/web-framework-evaluation.md for the full comparison.

## Scope

- Add github.com/go-chi/chi/v5 dependency
- Refactor NewHandler() in internal/webui/server.go to use Chi router with route groups:
  - Static assets group with immutable cache middleware
  - HTML pages group with ETag/no-cache middleware
  - WebSocket endpoints group
  - API endpoints group
  - Dev-only routes group (conditional)
- Add middleware.Recoverer for panic recovery
- Existing handler logic, WebSocket code, fingerprinting, and tests should remain largely unchanged

## Acceptance Criteria

- [ ] Chi router replaces http.ServeMux in server.go
- [ ] Cache headers are set via scoped middleware, not per-handler
- [ ] Route groups organize routes by concern (assets, pages, ws, api, dev)
- [ ] All existing integration tests pass
- [ ] No behavioral changes visible to clients (same URLs, same headers, same WS protocol)

## Out of Scope

- Replacing gorilla/websocket
- Changing the fingerprinting pipeline
- Adding new middleware (CORS, compression, etc.) — separate tickets
