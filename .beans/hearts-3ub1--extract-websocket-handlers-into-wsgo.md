---
# hearts-3ub1
title: Extract WebSocket handlers into ws.go
status: completed
type: task
priority: low
created_at: 2026-03-26T09:55:30Z
updated_at: 2026-03-26T11:06:36Z
parent: hearts-aazx
---

## Context

WebSocket handling is the single biggest concern in server.go (~250 lines): two handlers, message/command types, goroutine-based reader/writer, JSON protocol, and command dispatch. This is the extraction with the highest impact on readability.

## Higher Goal

Part of the server.go decomposition (hearts-aazx). Reduce server.go to pure wiring by extracting each concern into its own file.

## Acceptance Criteria

- [x] WebSocket handlers, types, and dispatch live in `ws.go`
- [x] `server.go` calls registration functions from the new file(s)
- [x] All existing tests pass without modification
- [x] Re-evaluated current state of server.go before extracting

## Guidance

- The WS handlers depend on session.Manager, presence trackers, and lobbyHub. Assess whether these dependencies make the interface wide or narrow.
- If the WebSocket layer ends up with a clean, narrow interface (e.g. just needs a router group + a few dependencies injected), it's a strong candidate for a sub-package (`internal/webui/ws`). Evaluate at implementation time.

## Out of Scope

- Changing the WebSocket protocol or message format
- Refactoring command dispatch logic


> **Update (hearts-5fnr):** Presence trackers and ConnTracker now live in `internal/webui/tracker/`. The WS handlers already import this package, so extracting to `ws.go` (same package) means no new imports. If extracting to a sub-package (`internal/webui/ws`), both `tracker` and `session` would be imported — still a narrow interface.

## Summary of Changes

Extracted WebSocket handlers (`handleLobbyWebSocket`, `handleTableWebSocket`), message types (`wsMessage`, `wsCommand`), constants, and `truncateUTF8` from `server.go` into `ws.go`. server.go reduced from 723 to 393 lines. Route registration and upgrader creation remain in `server.go`. Kept as a single file since lobby and table handlers share types.
