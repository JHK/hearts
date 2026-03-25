---
# hearts-jani
title: Show game state in lobby
status: completed
type: feature
priority: normal
tags:
    - frontend
created_at: 2026-03-25T09:07:07Z
updated_at: 2026-03-25T09:52:29Z
parent: hearts-g7wu
---

Display table status (waiting/in progress) and player count in the lobby so players can find games that need them

## Context
The lobby lists tables but gives no indication whether a table is waiting for players or has a game in progress. Players can't tell where they'd be useful.

## Higher Goal
Help players find games that need them, making the lobby an effective matchmaking surface.

## Acceptance Criteria
- [x] Each table in the lobby shows its current state (e.g. waiting, in progress, finished)
- [x] The player count per table is visible (e.g. "2/4 players")
- [x] State updates in real-time without requiring a page refresh

## Out of Scope
- Filtering or sorting tables by state
- Showing detailed game progress (score, round number)

## Summary of Changes

Improved lobby table list to clearly display game state and player count:
- Each table now shows its name on one line with a colored status badge ("Waiting" in green, "In progress" in amber) and player count (e.g. "2/4 players") below
- Added CSS styles for `.table-info`, `.table-name`, `.table-meta`, `.badge`, `.badge-waiting`, and `.badge-active`
- Real-time updates were already in place via 1.5s polling
