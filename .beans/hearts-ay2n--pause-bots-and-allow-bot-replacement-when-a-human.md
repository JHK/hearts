---
# hearts-ay2n
title: Pause bots and allow bot replacement when a human disconnects
status: completed
type: feature
priority: normal
tags:
    - backend
    - frontend
created_at: 2026-03-25T09:04:51Z
updated_at: 2026-03-25T12:24:41Z
parent: hearts-g7wu
---

When a human disconnects, pause all bots and let remaining players choose to replace the disconnected player with a bot

## Context
When a human player disconnects mid-game, bots currently keep playing as if nothing happened. This creates a poor experience for the remaining humans who are waiting for the disconnected player.

## Higher Goal
Make multiplayer games resilient to transient disconnections without forcing the game to continue on autopilot.

## Acceptance Criteria
- [x] When a human disconnects, all bots at the table immediately stop playing
- [x] Remaining human players see that the game is paused and who disconnected
- [x] Any remaining human player can choose to replace the disconnected player with a bot
- [x] Once replaced (or reconnected), bots resume normal play
- [x] If the disconnected player reconnects before being replaced, they reclaim their seat and play resumes

## Out of Scope
- Automatic bot replacement (must be a manual choice)
- Timeout-based auto-replacement
- Replacing a human who is still connected


## Summary of Changes

- Added `game_paused` and `game_resumed` protocol events
- Modified `handleLeave` to pause the game (mark player disconnected) instead of silently converting to a bot
- Added `ReplaceWithBot` command allowing any seated human to replace a disconnected player with a smart bot
- On reconnection, disconnected players reclaim their seat and the game resumes
- All game commands (play, pass, ready) and bot actions are blocked while paused
- Frontend shows a pause overlay with who disconnected and a "Replace with Bot" button
- Added tests for pause, replace, and reconnect flows
