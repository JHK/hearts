---
# hearts-a27e
title: Notify remaining players when someone leaves after game over
status: completed
type: bug
priority: normal
created_at: 2026-03-25T14:00:38Z
updated_at: 2026-03-25T14:02:47Z
---

When a player leaves a table after game over: (1) remaining players aren't notified, (2) the table stays hidden from the lobby because gameOver remains true. Fix: add EventPlayerLeft, publish it on leave during game-over, and reset gameOver so the table becomes visible again.

## Summary of Changes

- Added `EventPlayerLeft` event type and `PlayerLeftData` struct to the protocol layer
- Updated `handleLeave` to publish `player_left` event when a player is removed (idle or game-over state)
- On leave during game-over: reset `gameOver`, `rematchVotes`, and `roundHistory` so the table becomes visible in the lobby again
- Frontend: handle `player_left` event with log message and snapshot refresh
- Frontend: auto-reset game-over chart state when transitioning out of game-over
- Updated existing test to verify game-over reset on leave
