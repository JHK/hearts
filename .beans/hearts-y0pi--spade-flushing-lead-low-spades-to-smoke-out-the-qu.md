---
# hearts-y0pi
title: 'Spade flushing: lead low spades to smoke out the Queen'
status: todo
type: task
priority: high
tags:
    - backend
created_at: 2026-03-25T19:02:06Z
updated_at: 2026-03-25T19:02:51Z
parent: hearts-8j8z
---

Lead low spades to force Q♠ out; lead A♠/K♠ after low spades have thinned holdings


## Context
The hard bot never deliberately leads spades to flush out the Q♠. It avoids leading suits where opponents are void, but it doesn't proactively hunt the Queen. Leading low spades is one of the most widely recommended Hearts strategies — it forces the Q♠ holder to follow suit or use it, and creates opportunities for A♠/K♠ holders to safely play their high spades.

## Higher Goal
Reduce the frequency of being stuck with Q♠ (13 points) and increase control over when the Queen emerges.

## Implementation Notes
From research, the spade flushing strategy has two phases:

**Phase 1 — Smoke out:** When you hold low spades (2-7) and the Q♠ hasn't appeared yet, lead low spades. This forces all players to follow and draws out the Queen or depletes opponents' spade holdings.

**Phase 2 — Safe high play:** When you hold A♠ or K♠ without Q♠, lead them *after* some low spades have been played. With fewer spades in opponents' hands, the chance of drawing the Queen onto your trick decreases.

**When NOT to flush:**
- When you hold Q♠ yourself (you want to dump it, not lead spades)
- When you're pursuing a moon-shot (you want the Queen)
- When spades are already well-thinned (diminishing returns)

## Acceptance Criteria
- [ ] Hard bot leads low spades when Q♠ is still out and bot doesn't hold it
- [ ] Bot tracks whether Q♠ has been played (already have card tracking infrastructure)
- [ ] Bot avoids flushing when holding Q♠ or in moon-shot mode
- [ ] After some low spades have been led, bot leads A♠/K♠ if held (safer after thinning)
- [ ] Benchmark: 50k+ sim iterations before/after; win-rate must not decrease

## Out of Scope
- Coordinating spade flushes with other bots
- Varying flush aggressiveness based on game score

## References
- [Mark's Advanced Hearts](https://mark.random-article.com/hearts/advanced.html): lead A♠, K♠, J♠ sequentially to remove spades before they can be led back
- [Wikibooks Hearts Strategy](https://en.wikibooks.org/wiki/Card_Games/Hearts/Strategy): lead low spades repeatedly to force Q♠ holder to play it
- [247 Solitaire Strategy Guide](https://www.247solitaire.com/news/hearts-card-game-strategy-guide/): flush the Queen early by leading low spades
