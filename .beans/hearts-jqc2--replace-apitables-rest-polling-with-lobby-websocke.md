---
# hearts-jqc2
title: Replace /api/tables REST polling with lobby WebSocket commands
status: todo
type: feature
priority: normal
tags:
    - backend
    - frontend
created_at: 2026-03-26T13:46:09Z
updated_at: 2026-03-26T13:46:23Z
---

Move table listing and creation from polled REST endpoints to push-based WebSocket messages on /ws/lobby, then remove the /api/tables endpoint entirely


## Context

The lobby page polls `GET /api/tables` every 1.5 seconds (`setInterval(fetchTables, 1500)` in `lobby/main.js:211`) and uses `POST /api/tables` to create tables. Meanwhile, a lobby WebSocket (`/ws/lobby`) is already connected but only carries player presence data (`lobby_presence`, `lobby_self`). Both REST consumers are internal to this codebase — no external callers exist.

## Higher Goal

Consistent push-based architecture. The table game is fully WebSocket-driven, and the lobby should follow the same pattern. Eliminates unnecessary HTTP polling traffic and gives users instant feedback when tables are created or their state changes.

## Acceptance Criteria

- [ ] Server pushes a table list event (e.g. `lobby_tables`) through `/ws/lobby` whenever a table is created, a player joins/leaves, a game starts, or a game ends
- [ ] Initial table list is sent to the client on WebSocket connect (no separate fetch needed on page load)
- [ ] Table creation uses a WebSocket command (e.g. `create_table`) instead of `POST /api/tables`
- [ ] Lobby JS renders table list and creates tables entirely via WebSocket — no REST calls remain
- [ ] `GET /api/tables` and `POST /api/tables` endpoints are removed
- [ ] `routes_api.go` is removed or reduced to only dev/debug routes
- [ ] No visible regression in lobby UX — table list updates at least as fast as before

## Out of Scope

- Changes to the `/ws/table/{id}` game WebSocket protocol
- Adding new table metadata beyond what `TableInfo` already provides
- Lobby presence changes (already WebSocket-based)
