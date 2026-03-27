---
# hearts-8j8z
title: Smarter bot play
status: todo
type: epic
priority: normal
created_at: 2026-03-25T17:20:09Z
updated_at: 2026-03-27T09:47:45Z
---

Improve smart bot decision-making: moonshot pursuit, defensive play, pass strategy

## Vision
The smart bot should play Hearts at a level that feels competitive and realistic to human players. This means improving decision-making in areas like moonshot pursuit, defensive play, and pass strategy — informed by observed weaknesses during real games.

## Context
The current smart bot (`internal/game/bot/smart.go`) plays reasonably well but has known blind spots. For example, the moonshot abort logic is too aggressive — it kills a viable moonshot attempt simply because another player led a trick, without considering whether the bot still holds all penalty points. As the game matures, these rough edges become more visible.

## Validation

Follow this sequential, iterative approach:

1. **Baseline first**: Run a single 250k simulation to establish the current baseline win rates. Record these numbers.
2. **Implement one change at a time**: Make a single improvement, then measure it with a 50k simulation run.
3. **Sequential only**: Do NOT run multiple simulations in parallel. Wait for each run to complete before starting the next.
4. **Threshold**: An improvement must be at least **1pp** (percentage point) on a 50k run to be considered significant. 50k runs are sufficient for detecting improvements of this size.
5. **Iterate**: If the change improves win rates by ≥1pp, keep it and move to the next improvement. If not, revert and try a different approach.

**Existing code/tests may be freely rewritten or removed** as long as a sim confirms a significant improvement (≥1pp on 50k) over the previous baseline. The goal is better play, not preserving existing heuristics.

## Out of Scope
- ~~Adding new bot difficulty tiers~~ → moved to hearts-gavv
- AI/ML-based bot strategies
- Multiplayer matchmaking or ELO
