---
# hearts-2lkv
title: Handle stale or invalid table URLs gracefully
status: completed
type: feature
priority: normal
tags:
    - frontend
created_at: 2026-03-25T09:08:40Z
updated_at: 2026-03-25T10:25:42Z
parent: hearts-g7wu
---

Show an info message and auto-redirect to lobby when visiting a table URL that no longer exists

## Context
Visiting a non-existent table URL (e.g. from a stale browser tab, a bookmark to a finished game, or after a server restart) currently auto-creates a new table. The player expects to rejoin a game that no longer exists.

## Higher Goal
Give players clear feedback when a game is gone and guide them back to the lobby instead of silently creating an empty table.

## Acceptance Criteria
- [x] Visiting a stale or invalid table URL redirects to the lobby
- [x] No new table is created as a side effect of visiting an invalid URL

## Out of Scope
- Remembering which table the player was in and suggesting alternatives
- Deep-link resurrection (recreating the exact game state)

## Summary of Changes

- WebSocket handler (`server.go`) now uses `manager.Get()` instead of `manager.Create()` — sends a `table_not_found` message and closes the connection if the table doesn't exist
- Client JS (`table/main.js`) handles `table_not_found` by showing an info message and auto-redirecting to the lobby after 3 seconds, with no reconnect attempts
- Updated integration tests to pre-create tables and added `TestWebSocketRejectsNonExistentTable`
- Initial WebSocket connect retries up to 2 times before redirecting, to tolerate transient network failures
