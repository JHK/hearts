---
# hearts-g7wu
title: Multiplayer
status: todo
type: epic
priority: normal
created_at: 2026-03-24T11:04:40Z
updated_at: 2026-03-25T09:17:50Z
---

## Vision
Make the game a proper multiplayer experience where humans can find each other, play together, and handle the messy realities of real connections — players dropping, games ending, stale links.

## Context
The game currently works well for a single human playing with bots. The multiplayer infrastructure (WebSocket, observer mode, reconnection) exists, but the UX doesn't guide players toward each other or handle disruptions gracefully.

Note: "Persist game state across restarts" was split into its own epic (hearts-oeb4).
