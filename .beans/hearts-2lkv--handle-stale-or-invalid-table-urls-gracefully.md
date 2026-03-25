---
# hearts-2lkv
title: Handle stale or invalid table URLs gracefully
status: todo
type: feature
priority: normal
tags:
    - frontend
created_at: 2026-03-25T09:08:40Z
updated_at: 2026-03-25T09:08:48Z
parent: hearts-g7wu
---

Show an info message and auto-redirect to lobby when visiting a table URL that no longer exists

## Context
Visiting a non-existent table URL (e.g. from a stale browser tab, a bookmark to a finished game, or after a server restart) currently auto-creates a new table. The player expects to rejoin a game that no longer exists.

## Higher Goal
Give players clear feedback when a game is gone and guide them back to the lobby instead of silently creating an empty table.

## Acceptance Criteria
- [ ] Visiting a table URL that doesn't exist shows an informational message ("This game no longer exists" or similar)
- [ ] The player is automatically redirected to the lobby after a brief delay
- [ ] No new table is created as a side effect of visiting an invalid URL

## Out of Scope
- Remembering which table the player was in and suggesting alternatives
- Deep-link resurrection (recreating the exact game state)
