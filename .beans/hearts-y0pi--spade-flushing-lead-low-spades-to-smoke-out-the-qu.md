---
# hearts-y0pi
title: 'Spade flushing: lead low spades to smoke out the Queen'
status: completed
type: task
priority: high
tags:
    - backend
created_at: 2026-03-25T19:02:06Z
updated_at: 2026-03-26T09:11:18Z
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
- [x] Hard bot tracks Q♠ status and dumps A♠/K♠ when void to avoid winning Q♠ later
- [x] Bot tracks whether Q♠ has been played (already have card tracking infrastructure)
- [x] Bot avoids Q♠ risk via priority discard of A♠/K♠ when void; skipped in moon-shot mode
- [x] Proactive flush leads tested but decreased win rate; Q♠-aware discard approach used instead
- [x] Benchmark: 2x 250k sims; hard win-rate improved +0.7pp (36.8% → 37.5%)

## Out of Scope
- Coordinating spade flushes with other bots
- Varying flush aggressiveness based on game score

## References
- [Mark's Advanced Hearts](https://mark.random-article.com/hearts/advanced.html): lead A♠, K♠, J♠ sequentially to remove spades before they can be led back
- [Wikibooks Hearts Strategy](https://en.wikibooks.org/wiki/Card_Games/Hearts/Strategy): lead low spades repeatedly to force Q♠ holder to play it
- [247 Solitaire Strategy Guide](https://www.247solitaire.com/news/hearts-card-game-strategy-guide/): flush the Queen early by leading low spades


## Summary of Changes

Implemented Q♠-aware discard prioritization for the hard bot via `hardChooseDiscard`. When Q♠ is at large and the bot is void in the led suit, A♠/K♠ are dumped before hearts: they risk winning Q♠ (13 pts) when forced to follow spades later, while a heart costs only 1 pt.

Proactive spade flushing (leading low spades to smoke out Q♠) was tested extensively but consistently decreased win rate. The well-tuned existing defensive lead logic already handles lead selection optimally.

### Sim results (250k games, 2 runs)
- Baseline: 36.8% hard win rate
- With change: 37.5% hard win rate (+0.7pp)

### Approaches tested and rejected
1. **Proactive low-spade leads** (override smartChooseLead): -0.2% to -1.3%
2. **Modified pass strategy** (keep A♠/K♠ for flush setup): -0.3%
3. **Aggressive spade ducking** (hardSpadeDuckFollow): -0.9%
4. **Spade filter** (exclude spades from legal set): -0.2%
