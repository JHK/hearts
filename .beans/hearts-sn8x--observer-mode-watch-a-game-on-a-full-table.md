---
# hearts-sn8x
title: 'Observer mode: watch a game on a full table'
status: todo
type: feature
priority: normal
created_at: 2026-03-18T10:47:25Z
updated_at: 2026-03-18T10:47:32Z
---

Allow users to join a full table as a silent spectator with real-time game updates

## Context

A table is currently closed to anyone once all 4 seats are taken. There's no way to watch a game in progress. This blocks casual spectating — e.g. waiting for a seat to open, or watching a bot game play out.

## Higher Goal

Makes the game social beyond its 4 active players and enables passive engagement with ongoing games.

## Acceptance Criteria

- [ ] A user can navigate to a full table and join as an observer without taking a seat
- [ ] Observers receive real-time game updates over WebSocket (tricks played, scores, passing phase, etc.)
- [ ] Observers see only what a real spectator would: the current trick, played cards, scores, and which player's turn it is — not any player's hand
- [ ] Observers are not prompted for any game action (pass, play card)
- [ ] The number of observers is not capped

## Out of Scope

- Observer chat or reactions
- Observers seeing any player's held cards
- Replay or history — observers only see live state from the moment they join
- Differentiated observer UI (e.g. "choose a player's POV") — single neutral view only
