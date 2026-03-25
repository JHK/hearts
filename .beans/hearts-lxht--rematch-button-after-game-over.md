---
# hearts-lxht
title: Rematch button after game over
status: completed
type: feature
priority: normal
tags:
    - frontend
created_at: 2026-03-25T09:05:56Z
updated_at: 2026-03-25T13:35:12Z
parent: hearts-g7wu
---

Add a 'Play again' button on the game-over screen so players can quickly start another game with the same group

## Context
After a game ends, players who want to play again must manually navigate back to the lobby and rejoin. There's no way to quickly start another game with the same group.

## Higher Goal
Reduce friction between games to keep a group of players together.

## Acceptance Criteria
- [x] A "Play again" button appears on the game-over screen
- [x] When all remaining human players confirm, a new game starts at the same table
- [x] Bot seats are automatically filled again
- [x] Players who leave instead of rematching free their seat

## Out of Scope
- Changing table settings between games
- Matchmaking or inviting new players for the rematch
- Preserving cumulative stats across rematches

## Summary of Changes

### Backend
- Added `rematch` command and `handleRematch` handler in session layer
- Added `EventRematchVote` and `EventRematchStarting` protocol events
- Added `RematchVoteData` and `RematchStartingData` protocol types
- Snapshot includes `rematch_votes`, `rematch_total`, `rematch_voted` fields
- `startRematch()` removes bots, resets game/round/history state, clears votes
- Leave clears rematch vote for departing player

### Frontend
- Added "Play Again" button and vote status on game-over overlay
- Added `rematch` WS command and handlers for `rematch_vote`/`rematch_starting` events
- `rematch_starting` resets all client-side game state (chart, animations, etc.)

### Tests
- `TestRematchResetsGameAfterAllHumansVote` — full rematch flow
- `TestRematchRejectsBot` — bots cannot vote
- `TestRematchLeaveClearsVote` — leaving clears vote
