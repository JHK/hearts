---
# hearts-ebsx
title: Extract page handlers into routes_pages.go
status: todo
type: task
priority: low
created_at: 2026-03-26T09:55:28Z
updated_at: 2026-03-26T09:55:51Z
parent: hearts-aazx
---

## Context

server.go mixes page-serving handlers (lobby index, table page, favicon/icons) with unrelated concerns. These handlers share a cohesive theme: rendering HTML pages and serving static icons with ETag support.

## Higher Goal

Part of the server.go decomposition (hearts-aazx). Reduce server.go to pure wiring by extracting each concern into its own file.

## Acceptance Criteria

- [ ] Page handlers and ETag helpers live in `routes_pages.go` (or a better name if warranted after re-evaluation)
- [ ] `server.go` calls a registration function from the new file
- [ ] All existing tests pass without modification
- [ ] Re-evaluated current state of server.go before extracting (code will have changed if earlier siblings landed)

## Guidance

- If the interface between page handlers and the rest of the package is narrow (e.g. only needs the router and template data), consider whether a sub-package would reduce the package's surface area. Likely not worth it here, but assess.

## Out of Scope

- Changing handler behavior or response format
- Refactoring template rendering logic
