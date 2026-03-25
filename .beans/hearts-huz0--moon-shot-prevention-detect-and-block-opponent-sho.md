---
# hearts-huz0
title: 'Moon-shot prevention: detect and block opponent shooting'
status: todo
type: task
priority: critical
tags:
    - backend
created_at: 2026-03-25T19:02:05Z
updated_at: 2026-03-25T19:02:39Z
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
- [ ] Hard bot tracks penalty-point distribution across players during a round
- [ ] When one opponent has taken all penalty points so far (after trick 3+), bot enters "block moon" mode
- [ ] In block-moon mode, bot saves highest heart to win a heart trick and break the shoot
- [ ] In block-moon mode, bot adjusts following/discarding to avoid feeding the shooter
- [ ] Benchmark: 50k+ sim iterations before/after; win-rate must not decrease
- [ ] Moon-shot success rate of opponents should decrease measurably

## Out of Scope
- Cooperative multi-bot moon blocking (each bot decides independently)
- Adjusting pass strategy to prevent moon-shots pre-emptively

## References
- [Wikibooks Hearts Strategy](https://en.wikibooks.org/wiki/Card_Games/Hearts/Strategy): hold A♥ to block, save 2-3 strong cards for final tricks
- [Mark's Advanced Hearts](https://mark.random-article.com/hearts/advanced.html): count hearts remaining to detect moon attempts
- [MobilityWare Hearts](https://mobilityware.helpshift.com/hc/en/42-hearts-card-game/faq/2628-what-s-a-good-shoot-the-moon-strategy/): detection signals — opponent passes low cards, leads high repeatedly
