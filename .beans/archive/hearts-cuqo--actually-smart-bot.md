---
# hearts-cuqo
title: Actually smart bot
status: completed
type: feature
priority: normal
created_at: 2026-03-15T16:38:51Z
updated_at: 2026-03-20T10:42:58Z
---

## Context

The current bot (`dumb`) plays with a naive strategy that frequently backfires in normal (non-moon-shot) play:

- It leads and plays high cards (A, K, Q) even when not attempting to shoot the moon, causing it to win penalty tricks.
- When it can't follow suit, it discards low-value cards instead of shedding its most dangerous cards (e.g. high hearts, Q♠).
- It passes cards without strategic intent during the trade phase.

This makes bots easy to exploit: human players can dump unwanted cards onto bot-won tricks with low risk.

## Higher Goal

Provide a credible computer opponent that plays with a coherent per-round strategy — chosen at the trade phase, maintained through play, and adapted when circumstances change — so games against bots feel competitive.

## Acceptance Criteria

- [ ] The existing bot is renamed to `dumb` (struct, file, registration) with no behavior changes
- [ ] A new `smart` bot is implemented and selectable at the table
- [ ] `smart` tracks which cards have been played across the round
- [ ] **Trade phase:** `smart` selects a strategy at the start of each round before passing, choosing from:
  - *Moon shot:* if holding a long dominant suit (e.g. 5+ high cards in one suit), pass cards that don't support the attempt
  - *Void creation:* if nearly void in a short suit, pass the remaining low cards to go fully void — enabling future dangerous-card dumps
  - *Defensive:* otherwise pass the most dangerous cards (high hearts, Q♠, high cards in short suits)
- [ ] `smart` does not lead high cards (A, K, Q) unless they are safe (dangerous cards in that suit already played), it is pursuing a moon shot, or blocking one
- [ ] `smart` plays high cards aggressively when safe (e.g. A after K and Q are gone) — to take a harmless trick or block a potential moon shot
- [ ] `smart` prioritizes shedding dangerous cards (high hearts, Q♠) when unable to follow suit, or exploits voids to dump them
- [ ] `smart` evaluates moon-shot viability at round start and after each trick: triggers when it holds a long suit with high cards giving strong control
- [ ] `smart` pursues the moon shot aggressively once triggered: leads high, tries to win every trick
- [ ] `smart` aborts the moon-shot attempt immediately if any other player takes a point card — falling back to defensive play
- [ ] `smart` passes tests verifying: trade strategy selection, card-tracking, safe-high-card play, void exploitation, moon-shot triggering, moon-shot pursuit, and abort-on-leak behaviors in isolation
- [ ] Both `dumb` and `smart` remain available as bot options (no removal of `dumb`)

## Out of Scope

- UI for selecting bot difficulty at the table
- Tracking opponents' likely voids or hand composition
