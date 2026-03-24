---
# hearts-hwoh
title: Observers can claim a free or bot seat
status: draft
type: feature
priority: normal
created_at: 2026-03-18T10:48:36Z
updated_at: 2026-03-24T11:07:25Z
parent: hearts-g7wu
blocked_by:
    - hearts-sn8x
---

Allow observers to take an unoccupied or bot-controlled seat and join the game

## Context

Once observer mode exists, observers are passive — they have no way to join the game itself. A seat may be free because a player disconnected/left, or because it's occupied by a bot. Letting an observer claim it is the natural next step toward full participation.

## Higher Goal

Allows observers to transition into active players without leaving and rejoining, lowering the barrier to participation.

## Acceptance Criteria

- [ ] An observer sees a "Take seat" affordance on any unoccupied seat or bot-controlled seat
- [ ] Claiming a seat removes the bot (if any) immediately, even mid-hand, and the observer takes over from that point
- [ ] If multiple observers try to claim the same seat simultaneously, exactly one succeeds; the rest remain observers
- [ ] The newly seated player receives the full current game state (hand, trick, scores) upon claiming

## Out of Scope

- Claiming a seat held by a human player (disconnected or otherwise)
- Choosing which seat to observe from before claiming
- Observers being prompted or queued for seats automatically
