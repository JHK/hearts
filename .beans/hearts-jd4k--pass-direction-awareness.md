---
# hearts-jd4k
title: Pass direction awareness
status: todo
type: task
priority: normal
tags:
    - backend
created_at: 2026-03-25T19:02:07Z
updated_at: 2026-03-25T19:03:03Z
parent: hearts-8j8z
---

Adjust pass risk scoring based on whether passing left (dangerous), right (safer), or across


## Context
The hard bot's `passRisk()` scoring ignores pass direction entirely. In Hearts, pass direction significantly affects optimal strategy:
- **Left:** You pass to the player who plays *after* you. They can use your passed cards against you immediately.
- **Right:** You pass to the player who plays *before* you. Less dangerous — they act before seeing your play.
- **Across:** Neutral — the recipient is two seats away.
- **Hold (no pass):** No passing occurs.

## Higher Goal
Improve pass quality by accounting for the positional relationship with the recipient.

## Implementation Notes
Key adjustments by direction:

**Passing left (most dangerous):**
- Higher priority to pass dangerous cards (Q♠, high hearts) — recipient plays right after you
- Avoid passing cards that create a void the recipient can exploit against you
- Be cautious about passing high spades — they may lead them back at you

**Passing right (safer):**
- Passing high cards is less risky — recipient acts before you, so you can react
- Can afford to keep slightly more dangerous hands

**Passing across:**
- Moderate risk; recipient doesn't directly follow you in play order
- Standard risk scoring is roughly appropriate

The `ChoosePass` method receives pass direction. Adjust `passRisk()` weights or apply a direction-based multiplier.

## Acceptance Criteria
- [ ] `passRisk()` or pass selection incorporates pass direction as a factor
- [ ] Passing left increases the weight toward shedding dangerous cards
- [ ] Passing right allows slightly more risk retention
- [ ] Benchmark: 50k+ sim iterations before/after; win-rate must not decrease

## Out of Scope
- Modeling what the recipient will pass to *their* neighbor
- Remembering what was passed in previous rounds

## References
- [Wikibooks Hearts Strategy](https://en.wikibooks.org/wiki/Card_Games/Hearts/Strategy): pass direction affects card danger
- [Mark's Advanced Hearts](https://mark.random-article.com/hearts/advanced.html): pair high spades with clubs when passing to disrupt opponent voids
