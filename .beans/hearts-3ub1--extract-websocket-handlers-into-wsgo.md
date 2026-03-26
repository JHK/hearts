---
# hearts-3ub1
title: Extract WebSocket handlers into ws.go
status: todo
type: task
priority: low
created_at: 2026-03-26T09:55:30Z
updated_at: 2026-03-26T09:56:03Z
parent: hearts-aazx
---

## Context

WebSocket handling is the single biggest concern in server.go (~250 lines): two handlers, message/command types, goroutine-based reader/writer, JSON protocol, and command dispatch. This is the extraction with the highest impact on readability.

## Higher Goal

Part of the server.go decomposition (hearts-aazx). Reduce server.go to pure wiring by extracting each concern into its own file.

## Acceptance Criteria

- [ ] WebSocket handlers, types, and dispatch live in `ws.go` (or split into `ws_lobby.go` + `ws_table.go` if that reads better after re-evaluation)
- [ ] `server.go` calls registration functions from the new file(s)
- [ ] All existing tests pass without modification
- [ ] Re-evaluated current state of server.go before extracting

## Guidance

- The WS handlers depend on session.Manager, presence trackers, and lobbyHub. Assess whether these dependencies make the interface wide or narrow.
- If the WebSocket layer ends up with a clean, narrow interface (e.g. just needs a router group + a few dependencies injected), it's a strong candidate for a sub-package (`internal/webui/ws`). Evaluate at implementation time.

## Out of Scope

- Changing the WebSocket protocol or message format
- Refactoring command dispatch logic
