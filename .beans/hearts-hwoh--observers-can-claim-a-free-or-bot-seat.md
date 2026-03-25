---
# hearts-hwoh
title: Observers can claim a bot seat
status: todo
type: feature
priority: normal
created_at: 2026-03-18T10:48:36Z
updated_at: 2026-03-25T12:36:25Z
parent: hearts-g7wu
blocked_by:
    - hearts-sn8x
---

Allow observers to take a bot-controlled seat and join the game, in any phase

## Context

Observer mode is live — observers can watch but have no way to become a player. Seats may be bot-controlled either because the table was filled with bots at start, or because a human disconnected and was converted to a bot. Letting an observer claim any bot seat is the natural next step.

## Higher Goal

Allows observers to transition into active players without leaving and rejoining, lowering the barrier to participation.

## Acceptance Criteria

- [ ] An observer sees a "Take seat" affordance on any bot-controlled seat
- [ ] Claiming a seat replaces the bot immediately in any game phase (pre-round, passing, playing); the observer inherits the bot's state as-is, including any actions the bot already performed
- [ ] If multiple observers try to claim the same seat simultaneously, exactly one succeeds; the rest remain observers (guaranteed by the Table actor's serialized command channel)
- [ ] The newly seated player receives the full current game state (hand, trick, scores) upon claiming
- [ ] A returning human (matching token) still reclaims their original seat automatically, taking priority over observer claims

## Out of Scope

- Claiming a seat held by a connected human player
- Observers being prompted or queued for seats automatically
- Choosing which seat to observe from before claiming
- Any distinction between "original bot" and "human-who-left-and-became-bot" seats — all bot seats are claimable
