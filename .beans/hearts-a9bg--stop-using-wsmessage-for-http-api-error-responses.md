---
# hearts-a9bg
title: Stop using wsMessage for HTTP API error responses
status: todo
type: task
priority: low
created_at: 2026-03-26T11:16:52Z
updated_at: 2026-03-26T11:16:57Z
parent: hearts-aazx
---

handleTablesAPI uses wsMessage for HTTP errors — replace with plain map/struct to remove cross-concern dependency on ws.go

## Context

`handleTablesAPI` in `server.go` uses `wsMessage{Type: "error", Error: "..."}` to format HTTP JSON error responses (lines 304, 310). This worked fine when `wsMessage` lived in the same file, but after extracting WebSocket types to `ws.go` (hearts-3ub1), it's a cross-concern dependency: an HTTP handler importing a WebSocket message type as its error envelope.

## Higher Goal

Keep the server.go decomposition (hearts-aazx) clean by ensuring each concern uses its own types rather than leaking across boundaries.

## Acceptance Criteria

- [ ] `handleTablesAPI` uses a plain `map[string]any` or a local struct for error responses instead of `wsMessage`
- [ ] HTTP error response shape (`{"type":"error","error":"..."}`) is preserved for client compatibility
- [ ] All existing tests pass without modification

## Out of Scope

- Changing the WebSocket message types or protocol
- Introducing a shared error type across HTTP and WS — the duplication is fine for two call sites
