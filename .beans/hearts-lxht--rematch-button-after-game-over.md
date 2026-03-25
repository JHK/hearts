---
# hearts-lxht
title: Rematch button after game over
status: todo
type: feature
priority: normal
tags:
    - frontend
created_at: 2026-03-25T09:05:56Z
updated_at: 2026-03-25T09:06:00Z
parent: hearts-g7wu
---

Add a 'Play again' button on the game-over screen so players can quickly start another game with the same group

## Context
After a game ends, players who want to play again must manually navigate back to the lobby and rejoin. There's no way to quickly start another game with the same group.

## Higher Goal
Reduce friction between games to keep a group of players together.

## Acceptance Criteria
- [ ] A "Play again" button appears on the game-over screen
- [ ] When all remaining human players confirm, a new game starts at the same table
- [ ] Bot seats are automatically filled again
- [ ] Players who leave instead of rematching free their seat

## Out of Scope
- Changing table settings between games
- Matchmaking or inviting new players for the rematch
- Preserving cumulative stats across rematches
