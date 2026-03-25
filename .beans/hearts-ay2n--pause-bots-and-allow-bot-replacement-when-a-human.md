---
# hearts-ay2n
title: Pause bots and allow bot replacement when a human disconnects
status: todo
type: feature
priority: normal
tags:
    - backend
    - frontend
created_at: 2026-03-25T09:04:51Z
updated_at: 2026-03-25T09:04:56Z
parent: hearts-g7wu
---

When a human disconnects, pause all bots and let remaining players choose to replace the disconnected player with a bot

## Context
When a human player disconnects mid-game, bots currently keep playing as if nothing happened. This creates a poor experience for the remaining humans who are waiting for the disconnected player.

## Higher Goal
Make multiplayer games resilient to transient disconnections without forcing the game to continue on autopilot.

## Acceptance Criteria
- [ ] When a human disconnects, all bots at the table immediately stop playing
- [ ] Remaining human players see that the game is paused and who disconnected
- [ ] Any remaining human player can choose to replace the disconnected player with a bot
- [ ] Once replaced (or reconnected), bots resume normal play
- [ ] If the disconnected player reconnects before being replaced, they reclaim their seat and play resumes

## Out of Scope
- Automatic bot replacement (must be a manual choice)
- Timeout-based auto-replacement
- Replacing a human who is still connected
