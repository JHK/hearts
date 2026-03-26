---
# hearts-i07j
title: Monte Carlo hand sampling for play selection
status: todo
type: task
priority: high
tags:
    - backend
created_at: 2026-03-25T19:02:08Z
updated_at: 2026-03-26T09:21:20Z
parent: hearts-8j8z
---

Sample hypothetical opponent hands consistent with observed voids/played cards, simulate remaining tricks, pick move minimizing expected score


## Context
The hard bot currently uses rule-based heuristics for every decision. While these are effective for common situations, they can't evaluate edge cases or surprising card distributions. Monte Carlo (MC) hand sampling is the technique used by the strongest Hearts AIs — it samples possible opponent hands, simulates outcomes, and picks the move that minimizes expected score.

Research from the Moving AI Lab shows their MC-based Hearts bot averaged 55.8 points/game vs. 75.1 for Hearts Deluxe. However, the same research notes that rule-based "strong heuristic" agents can be competitive with MC approaches, so this should be benchmarked carefully.

## Higher Goal
Move from purely rule-based decisions to evidence-based decisions where the bot evaluates actual outcomes across many possible worlds.

## Implementation Notes
**Algorithm outline:**
1. For each legal play, generate N hypothetical opponent hands consistent with:
   - Cards already played (known)
   - Known voids (opponent failed to follow suit)
   - Cards passed (if applicable)
   - Cards remaining = 52 - played - own hand
2. For each sampled world, simulate the remaining tricks using a fast heuristic (e.g., easy bot logic)
3. Score each candidate move by average penalty points across samples
4. Pick the move with lowest expected score

**Key considerations:**
- N doesn't need to be large — 50-100 samples per move may suffice
- Simulation speed matters: use simple heuristics for simulated opponents, not full hard-bot logic
- Can be applied selectively: only use MC for ambiguous decisions where heuristics are uncertain
- UCT (Upper Confidence bounds applied to Trees) can improve sample efficiency
- Must not slow down real-time play noticeably (bot think time)

**Phased approach:**
1. First: implement hand sampling consistent with known constraints
2. Then: implement fast trick simulation
3. Then: integrate into hard bot as a tiebreaker or override for close decisions
4. Finally: tune N and evaluate full replacement of heuristics

## Acceptance Criteria
- [ ] Hand sampler generates valid opponent hands respecting played cards and known voids
- [ ] Fast trick simulator can play out remaining tricks using simple heuristics
- [ ] MC evaluation integrated into hard bot for at least follow/lead decisions
- [ ] Bot think time remains under 100ms per decision (or configurable)
- [ ] Benchmark: 50k+ sim iterations before/after; win-rate must improve or stay neutral
- Note: existing code/tests may be freely rewritten or removed if a 250k sim shows ≥0.3pp improvement over the previous baseline
- [ ] Simulation runtime does not regress significantly (MC adds overhead per bot decision)

## Out of Scope
- Opponent modeling (inferring strategy from play patterns)
- Machine learning / neural network approaches
- MCTS with deep tree search (keep it to single-move evaluation)

## References
- [Moving AI Lab Hearts](https://www.movingai.com/hearts.html): MC sampling with UCT; averaged 55.8 vs 75.1 points/game against Hearts Deluxe
- [HeartsAI Framework](https://github.com/Devking/HeartsAI): open-source Hearts AI framework for algorithm comparison
- [Sturtevant & White](https://webdocs.cs.ualberta.ca/~nathanst/papers/heartslearning.pdf): feature construction for RL in Hearts
- [Stanford CS230 Hearts RL](http://cs230.stanford.edu/projects_spring_2021/reports/9.pdf): reinforcement learning approaches to Hearts
