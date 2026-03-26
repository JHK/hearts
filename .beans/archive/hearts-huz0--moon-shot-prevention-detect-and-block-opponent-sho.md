---
# hearts-huz0
title: 'Moon-shot prevention: detect and block opponent shooting'
status: completed
type: task
priority: critical
tags:
    - backend
created_at: 2026-03-25T19:02:05Z
updated_at: 2026-03-25T20:04:27Z
parent: hearts-8j8z
---

Detect when an opponent is collecting all penalty cards and save high hearts to block — a blocked moon-shot swings 52 points


## Context
The hard bot has sophisticated moon-shot *pursuit* logic but zero moon-shot *prevention*. When an opponent shoots the moon successfully, the score swing is 52 points (26 saved by them + 26 added to everyone else). Even preventing 1 in 10 opponent moon-shots would meaningfully improve win rates.

## Higher Goal
Make the hard bot competitive by closing the single largest scoring gap in its defensive play.

## Implementation Notes
Detection signals (from research):
- One player winning every trick so far
- One player collecting hearts without trying to avoid them
- A player passed you low cards (they kept high ones — possible moon intent)
- Track penalty points per player per round: if one player has taken all penalties so far, flag them

Blocking tactics:
- Save A♥ or a high heart to guarantee taking one heart trick
- When a shooter is detected, prioritize winning a trick that contains a heart — even at the cost of taking other penalty points
- Consider holding back a high card in the shooter's strong suit to steal a late trick

## Acceptance Criteria
- [x] Hard bot tracks penalty-point distribution across players during a round
- [x] When one opponent has taken all penalty points so far (after trick 4+, with 14+ points), bot enters "block moon" mode
- [x] In block-moon mode, bot leads safe high heart (guaranteed winner) to capture a heart trick and break the shoot
- [x] In block-moon mode, bot adjusts following/discarding to avoid feeding the shooter — NOTE: follow/discard adjustments were tested but regressed win rate; only lead intervention retained
- [x] Benchmark: 500k sim iterations before/after; win-rate 37.1% (unchanged from baseline)
- [x] Moon-shot success rate of opponents decreased ~1-2% (small but consistent across all opponent types)

## Out of Scope
- Cooperative multi-bot moon blocking (each bot decides independently)
- Adjusting pass strategy to prevent moon-shots pre-emptively

## References
- [Wikibooks Hearts Strategy](https://en.wikibooks.org/wiki/Card_Games/Hearts/Strategy): hold A♥ to block, save 2-3 strong cards for final tricks
- [Mark's Advanced Hearts](https://mark.random-article.com/hearts/advanced.html): count hearts remaining to detect moon attempts
- [MobilityWare Hearts](https://mobilityware.helpshift.com/hc/en/42-hearts-card-game/faq/2628-what-s-a-good-shoot-the-moon-strategy/): detection signals — opponent passes low cards, leads high repeatedly


## Summary of Changes

Added moon-shot prevention logic to the hard bot:

1. **Detection** (`detectMoonShooter`): After 4+ completed tricks, if one opponent holds all 14+ penalty points (Q♠ must be taken), flag them as a shooter.
2. **Lead blocking** (`hardBlockMoonLead`): When leading and a shooter is detected, lead the highest safe heart (guaranteed trick winner) to capture a heart and break the shoot. Falls back to defensive lead when no safe heart is available.
3. **State tracking**: Added `blockMoonTarget` field to `Hard` struct, reset at round/pass boundaries.

**Tested but removed**: Follow-suit and discard interventions consistently regressed win rate (~0.5-2%) due to the cost of voluntarily taking penalty tricks outweighing the blocking benefit at realistic false-positive rates. The lead-only intervention is nearly zero-cost because safe hearts guarantee winning the trick, and the Q♠ threshold ensures no Q♠ dump risk.

**Benchmark (500k games)**:
- Win rate: 37.1% → 37.1% (no regression)
- Opponent moon shots: ~1-2% reduction (consistent across medium/easy/random)

The effect is small because the detection window is narrow (leading + safe heart + shooter detected). More impactful blocking would require additional game state (trick seat info) not currently available in TurnInput.
