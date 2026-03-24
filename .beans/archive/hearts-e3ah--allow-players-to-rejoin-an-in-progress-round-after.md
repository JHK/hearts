---
# hearts-e3ah
title: Allow players to rejoin an in-progress round after disconnecting
status: completed
type: feature
priority: normal
created_at: 2026-03-17T08:11:19Z
updated_at: 2026-03-17T18:47:42Z
---

Preserve bot-replaced seats and add a grace period so disconnected players can reload and continue their game

## Context
When a player disconnects mid-round, the table runtime immediately converts their seat to a bot. If the player reconnects (same browser token), the join is rejected with "round already in progress" — so they're locked out of their own seat until the round ends. If they were the last human, the table is also destroyed immediately, making rejoining impossible entirely.

Reproducing the exact issue: player disconnects → converted to bot → player reloads browser → join rejected → they can no longer play in or even observe their own game.

## Current Behavior
- Mid-round disconnect: seat converted to bot; token cleared from `playersByToken`; rejoins rejected (`"round already in progress"`, runtime.go ~line 461)
- Last-human disconnect: table closed immediately (`manager.CloseTable`, server.go ~line 381)

## Desired Behavior
- A recently-disconnected player can rejoin an active round and reclaim their seat from the bot
- The table should not be destroyed immediately when the last human disconnects — a grace period should keep it alive for a short time so the player can reload and return
- Rejoining mid-round restores the player's hand and current game state

## Acceptance Criteria
- [ ] A disconnected player's token is preserved in `playersByToken` when they are converted to a bot mid-round
- [ ] A `join` with a matching token mid-round reclaims the seat: bot is removed, player resumes
- [ ] Table is not closed immediately when the last human leaves; a configurable grace period (e.g. 60 s) keeps the runtime alive
- [ ] After the grace period expires with no human rejoining, the table is closed normally
- [ ] Player reloads the browser, sees the current game state, and can continue playing

## Out of Scope
- Persistent game state across server restarts
- Spectator mode / observing a game you never joined
- Kicking bots from a game you were never part of

## Summary of Changes

Preserved player tokens in `playersByToken` on mid-round disconnect (instead of deleting them). Added seat-reclaim logic in `handleJoin`: when a token matches a bot-converted player, the bot is removed and the player resumes. Replaced immediate `CloseTable` on last-human-disconnect with a 60s grace period goroutine that only closes the table if no human has rejoined.
