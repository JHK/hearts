---
# hearts-yipb
title: Show player presence in lobby
status: completed
type: feature
priority: normal
tags:
    - frontend
created_at: 2026-03-25T09:13:06Z
updated_at: 2026-03-25T15:51:23Z
parent: hearts-g7wu
---

Display names of other players currently in the lobby to create a sense of liveness

## Context
The lobby currently gives no indication of whether other humans are online and looking for a game. A player arriving at an empty-looking lobby doesn't know if anyone else is around.

## Higher Goal
Create a sense of liveness so players know it's worth waiting or creating a table.

## Acceptance Criteria
- [x] The lobby shows the names of other players currently browsing it
- [x] When there are too many players to display, overflow is summarized (e.g. "Alice, Bob, Carol and 5 others are waiting")
- [x] The list updates in real-time as players arrive and leave
- [x] Players joining a table are removed from the lobby presence list

## Out of Scope
- Chat or direct interaction between lobby players
- "Looking for game" status or matchmaking queue

## Summary of Changes

Added real-time lobby presence showing which players are currently browsing the lobby:

- **Backend**: New `lobbyHub` (`lobby_hub.go`) tracks connected lobby browsers by token→name with pub/sub broadcasting. Each player gets a unique lobby ID so clients can filter themselves out.
- **WebSocket endpoint**: `/ws/lobby` — clients send `announce` with name+token, server broadcasts presence snapshots to all subscribers. On disconnect, player is automatically removed.
- **Frontend**: Lobby JS connects via WebSocket, displays "Alice, Bob are also in the lobby" with overflow summary for >8 players. Name changes are reflected in real-time.
- **Tests**: Integration tests for join/leave and name update flows.
