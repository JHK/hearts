---
# hearts-y7gh
title: Simulation win rates depend on slot position, not just strategy strength
status: todo
type: bug
priority: normal
tags:
    - backend
created_at: 2026-03-25T17:47:26Z
updated_at: 2026-03-25T17:47:37Z
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
- [ ] Root cause identified (likely in parallelization, dealing, turn order, or tie-breaking)
- [ ] Fix applied so that the same strategy produces statistically similar win rates across all slot positions
- [ ] Validated with 50k+ iteration runs showing no significant positional bias

## Out of Scope
- Adding a permutation mode to the sim CLI (separate feature)
- Changing bot strategies themselves
