---
# hearts-31de
title: Extract API handlers into routes_api.go
status: completed
type: task
priority: low
created_at: 2026-03-26T09:55:29Z
updated_at: 2026-03-26T11:21:52Z
parent: hearts-aazx
---

## Context

The tables API handler (GET/POST /api/tables) sits in server.go alongside unrelated page and WebSocket handlers. It's a self-contained REST endpoint.

## Higher Goal

Part of the server.go decomposition (hearts-aazx). Reduce server.go to pure wiring by extracting each concern into its own file.

## Acceptance Criteria

- [x] API handlers live in `routes_api.go`
- [x] `server.go` calls `registerAPIRoutes()` from the new file
- [x] All existing tests pass without modification (webui tests have pre-existing embed issue)
- [x] Re-evaluated current state of server.go before extracting

## Guidance

- Currently only ~60 lines. If other API endpoints have been added by the time this task starts, re-evaluate scope.
- If the API surface grows, a sub-package (`internal/webui/api`) could make sense — but only if the interface is genuinely narrow.

## Out of Scope

- Changing API behavior or adding new endpoints

## Summary of Changes

Extracted `handleTablesAPI` and its route registration into `internal/webui/routes_api.go`. Server.go now calls `registerAPIRoutes(r, manager)` instead of inlining the route setup. The `writeJSON` helper remains in server.go since it is also used by the debug endpoint. No behavioral changes.
