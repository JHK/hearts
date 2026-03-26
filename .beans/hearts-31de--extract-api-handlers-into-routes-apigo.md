---
# hearts-31de
title: Extract API handlers into routes_api.go
status: todo
type: task
priority: low
created_at: 2026-03-26T09:55:29Z
updated_at: 2026-03-26T09:55:54Z
parent: hearts-aazx
---

## Context

The tables API handler (GET/POST /api/tables) sits in server.go alongside unrelated page and WebSocket handlers. It's a self-contained REST endpoint.

## Higher Goal

Part of the server.go decomposition (hearts-aazx). Reduce server.go to pure wiring by extracting each concern into its own file.

## Acceptance Criteria

- [ ] API handlers live in `routes_api.go` (or a better name if warranted after re-evaluation)
- [ ] `server.go` calls a registration function from the new file
- [ ] All existing tests pass without modification
- [ ] Re-evaluated current state of server.go before extracting

## Guidance

- Currently only ~60 lines. If other API endpoints have been added by the time this task starts, re-evaluate scope.
- If the API surface grows, a sub-package (`internal/webui/api`) could make sense — but only if the interface is genuinely narrow.

## Out of Scope

- Changing API behavior or adding new endpoints
