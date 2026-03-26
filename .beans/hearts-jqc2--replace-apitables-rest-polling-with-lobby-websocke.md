---
# hearts-jqc2
title: Replace /api/tables REST polling with lobby WebSocket commands
status: completed
type: feature
priority: normal
tags:
    - backend
    - frontend
created_at: 2026-03-26T13:46:09Z
updated_at: 2026-03-26T14:02:49Z
---

Move table listing and creation from polled REST endpoints to push-based WebSocket messages on /ws/lobby, then remove the /api/tables endpoint entirely


## Context

The lobby page polls `GET /api/tables` every 1.5 seconds (`setInterval(fetchTables, 1500)` in `lobby/main.js:211`) and uses `POST /api/tables` to create tables. Meanwhile, a lobby WebSocket (`/ws/lobby`) is already connected but only carries player presence data (`lobby_presence`, `lobby_self`). Both REST consumers are internal to this codebase — no external callers exist.

## Higher Goal

Consistent push-based architecture. The table game is fully WebSocket-driven, and the lobby should follow the same pattern. Eliminates unnecessary HTTP polling traffic and gives users instant feedback when tables are created or their state changes.

## Acceptance Criteria

- [x] Server pushes a table list event (e.g. `lobby_tables`) through `/ws/lobby` whenever a table is created, a player joins/leaves, a game starts, or a game ends
- [x] Initial table list is sent to the client on WebSocket connect (no separate fetch needed on page load)
- [x] Table creation uses a WebSocket command (e.g. `create_table`) instead of `POST /api/tables`
- [x] Lobby JS renders table list and creates tables entirely via WebSocket — no REST calls remain
- [x] `GET /api/tables` and `POST /api/tables` endpoints are removed
- [x] `routes_api.go` is removed or reduced to only dev/debug routes
- [x] No visible regression in lobby UX — table list updates at least as fast as before

## Out of Scope

- Changes to the `/ws/table/{id}` game WebSocket protocol
- Adding new table metadata beyond what `TableInfo` already provides
- Lobby presence changes (already WebSocket-based)

## Summary of Changes

Replaced REST polling with push-based WebSocket messages on `/ws/lobby`:

- **Manager subscription**: Added `Subscribe()/notifyChange()` to `session.Manager` so lobby clients get notified when tables are created, closed, or change state.
- **Table onChange callback**: `Table` now accepts an `onChange` callback fired on lobby-relevant events (player join/leave, game start/end, pause/resume).
- **Lobby WebSocket extended**: `handleLobbyWebSocket` now subscribes to both presence and table changes. Sends `lobby_tables` on connect and on every change. Handles `create_table` command.
- **Frontend rewritten**: Lobby JS receives table list via WebSocket (`lobby_tables` messages) and creates tables via `create_table` command. No REST calls remain.
- **REST endpoints removed**: `GET /api/tables` and `POST /api/tables` removed. `routes_api.go` reduced to dev/debug routes only.
- **Tests updated**: Lobby hub tests updated to handle new `lobby_tables` messages. All tests pass.
