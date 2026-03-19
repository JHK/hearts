---
# hearts-m7mt
title: End game at 100 points with winner screen
status: todo
type: feature
priority: normal
created_at: 2026-03-19T07:24:21Z
updated_at: 2026-03-19T07:25:20Z
---

Detect when any player hits 100+ cumulative points, emit game-over event, show ending screen with winner(s), and close the table.

## Context
Hearts games currently run indefinitely — there is no end condition. Standard Hearts ends when any player's cumulative score reaches or exceeds 100 points at the end of a round, and the player with the *lowest* score at that point wins.

## Higher Goal
Make a complete, playable game of Hearts rather than an endless loop of rounds.

## Acceptance Criteria
- [ ] After each round, the table checks whether any player's cumulative score >= 100
- [ ] If the threshold is met, the table emits a game-over event (not a new round)
- [ ] The game-over event includes final cumulative scores and the winner(s) (lowest score)
- [ ] The frontend shows an ending screen with the final scores and who won
- [ ] The table is closed/shut down after the game ends (no further commands accepted)
- [ ] Ties for lowest score are handled — multiple co-winners are shown

## Out of Scope
- Persistent leaderboards or game history across server restarts
- A rematch / start-new-game flow from the ending screen (potential follow-up ticket)
- Any changes to how a reload on a closed table is handled
