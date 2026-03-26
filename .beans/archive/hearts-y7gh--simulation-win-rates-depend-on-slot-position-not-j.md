---
# hearts-y7gh
title: Simulation win rates depend on slot position, not just strategy strength
status: completed
type: bug
priority: normal
tags:
    - backend
created_at: 2026-03-25T17:47:26Z
updated_at: 2026-03-25T17:58:09Z
parent: hearts-8j8z
---

Slot index biases sim win rates — same strategy wins ~41% in slot 0 but ~34% in slot 3

## Context
The `cmd/sim` simulation runner is used to validate bot strategy changes under the smarter-bot-play epic (hearts-8j8z). The parallelization of iterations was introduced in hearts-87m7. Win rate results are unreliable because slot position dramatically affects outcomes.

## Current Behavior
When running 50k iterations, slot 0 consistently wins ~40% regardless of which strategy occupies it. Moving a strategy from slot 0 to slot 3 drops its win rate significantly (e.g. hard: 40.7% in slot 0 → 33.6% in slot 3). This makes the simulation useless for comparing strategy strength.

Example runs:
- hard in slot 0: 40.7% wins; hard in slot 3: 33.6% wins
- medium in slot 0: 40.6% wins; medium in slot 2: 37.2% wins
- random always ~1.7% regardless of slot (expected), but even easy jumps from 24.5% (slot 2) to 30.3% (slot 1)

## Desired Behavior
Win rates for a given strategy should be consistent regardless of which slot index it occupies. The simulation should produce strategy-dependent results, not position-dependent results.

## Acceptance Criteria
- [x] Root cause identified: fixed strategy-to-seat mapping caused neighbor-dependent bias (passing targets, trick order)
- [x] Fix applied: randomly permute strategy-to-seat assignment each game in sim.runGame
- [x] Validated: hard=36.4-37.0%, medium=36.5-37.2%, easy=27.3-28.0%, random=1.6-1.7% across 3 slot configurations at 50k iterations

## Out of Scope
- Adding a permutation mode to the sim CLI (separate feature)
- Changing bot strategies themselves

## Summary of Changes

Root cause: the simulation always assigned strategies to fixed seat indices, so each strategy always had the same neighbors for passing and trick order. This created systematic positional advantages from the interaction patterns (who you pass to/receive from, your position in trick play order relative to stronger/weaker players).

Fix: `sim.runGame` now randomly permutes the strategy-to-seat assignment each game using `rng.Perm(4)`, then maps seat-based results (wins, moon shots) back to strategy indices. The game mechanics themselves are fair (confirmed: 4 identical strategies produce ~25% each), so the permutation fully eliminates the bias.
