---
# hearts-jani
title: Show game state in lobby
status: todo
type: feature
priority: normal
tags:
    - frontend
created_at: 2026-03-25T09:07:07Z
updated_at: 2026-03-25T09:07:33Z
parent: hearts-g7wu
---

Display table status (waiting/in progress) and player count in the lobby so players can find games that need them

## Context
The lobby lists tables but gives no indication whether a table is waiting for players or has a game in progress. Players can't tell where they'd be useful.

## Higher Goal
Help players find games that need them, making the lobby an effective matchmaking surface.

## Acceptance Criteria
- [ ] Each table in the lobby shows its current state (e.g. waiting, in progress, finished)
- [ ] The player count per table is visible (e.g. "2/4 players")
- [ ] State updates in real-time without requiring a page refresh

## Out of Scope
- Filtering or sorting tables by state
- Showing detailed game progress (score, round number)
