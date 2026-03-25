---
# hearts-oeb4
title: Persist game state across server restarts
status: draft
type: epic
priority: normal
tags:
    - backend
created_at: 2026-03-25T09:16:28Z
updated_at: 2026-03-25T09:16:35Z
---

Introduce a storage backend so games survive server restarts and deploys

## Vision
Games survive server restarts. Players can reconnect to an in-progress game after a deploy or crash. This also lays the groundwork for features like game history or player accounts.

## Context
All state (tables, players, rounds, scores) lives in Go structs with no serialization or storage backend. A restart loses every active game. This is fine for development but blocks reliability for any real deployment.

Refinement should include a child ticket for choosing the storage approach (e.g. SQLite, file-based, Redis).

## Out of Scope
- Player accounts or authentication
- Game replay/history browsing
