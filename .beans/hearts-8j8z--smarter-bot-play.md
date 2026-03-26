---
# hearts-8j8z
title: Smarter bot play
status: todo
type: epic
priority: normal
created_at: 2026-03-25T17:20:09Z
updated_at: 2026-03-26T09:21:16Z
---

Improve smart bot decision-making: moonshot pursuit, defensive play, pass strategy

## Vision
The smart bot should play Hearts at a level that feels competitive and realistic to human players. This means improving decision-making in areas like moonshot pursuit, defensive play, and pass strategy — informed by observed weaknesses during real games.

## Context
The current smart bot (`internal/game/bot/smart.go`) plays reasonably well but has known blind spots. For example, the moonshot abort logic is too aggressive — it kills a viable moonshot attempt simply because another player led a trick, without considering whether the bot still holds all penalty points. As the game matures, these rough edges become more visible.

## Validation
Every change under this epic must be validated with a simulation run (`cmd/sim`) of at least 50k iterations. Win rates must not decrease compared to the baseline before the change.

**Existing code/tests may be freely rewritten or removed** as long as a sim provides a significant improvement (at least 0.3pp on a single 250k run) over the previous baseline. The goal is better play, not preserving existing heuristics.

## Out of Scope
- ~~Adding new bot difficulty tiers~~ → moved to hearts-gavv
- AI/ML-based bot strategies
- Multiplayer matchmaking or ELO
