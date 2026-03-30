---
# hearts-8j8z
title: Smarter bot play
status: todo
type: epic
priority: normal
created_at: 2026-03-25T17:20:09Z
updated_at: 2026-03-30T09:36:50Z
---

Improve smart bot decision-making: moonshot pursuit, defensive play, pass strategy

## Vision
The smart bot should play Hearts at a level that feels competitive and realistic to human players. This means improving decision-making in areas like moonshot pursuit, defensive play, and pass strategy — informed by observed weaknesses during real games.

## Context
The current smart bot (`internal/game/bot/smart.go`) plays reasonably well but has known blind spots. For example, the moonshot abort logic is too aggressive — it kills a viable moonshot attempt simply because another player led a trick, without considering whether the bot still holds all penalty points. As the game matures, these rough edges become more visible.

## Validation

Follow this sequential, iterative approach:

1. **Implement one change at a time**: Make a single improvement, then measure it with a 50k simulation run.
2. **Sequential only**: Do NOT run multiple simulations in parallel. Wait for each run to complete before starting the next.
3. **Threshold**: An improvement must be at least **1pp** (percentage point) on a 50k run to be considered significant. 50k runs are sufficient for detecting improvements of this size.
4. **Iterate**: If the change improves win rates by ≥1pp, keep it and move to the next improvement. If not, revert and try a different approach.
5. **Update baseline**: When a 50k run shows a significant improvement (≥1pp), re-run the 250k baseline and update the table in the **Baseline** section below.

**Existing code/tests may be freely rewritten or removed** as long as a sim confirms a significant improvement (≥1pp on 50k) over the previous baseline. The goal is better play, not preserving existing heuristics.

## Subtask guidance
All subtasks under this epic must follow the validation approach above. The 250k baseline is already recorded below — subtasks compare their 50k runs against it. When a 50k run shows a significant improvement (≥1pp), re-run the 250k baseline and update the table.

## Baseline (250k games, 2026-03-30)

| Strategy | Wins | Win Rate | Moon Shots |
|----------|------|----------|------------|
| Hard     | 97,767 | 39.1% | 8,845 |
| Medium   | 87,369 | 34.9% | 3,298 |
| Easy     | 68,931 | 27.6% | 3,029 |
| Random   | 3,102 | 1.2% | 63,500 |

Sim config: `cmd/sim -n 250000`, MCSamples=3, strategies=[hard, medium, easy, random], seat permutation per game.

## Out of Scope
- ~~Adding new bot difficulty tiers~~ → moved to hearts-gavv
- AI/ML-based bot strategies
- Multiplayer matchmaking or ELO
