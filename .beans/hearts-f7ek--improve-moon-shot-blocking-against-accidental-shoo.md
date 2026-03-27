---
# hearts-f7ek
title: Improve moon-shot blocking against accidental shooters
status: todo
type: task
created_at: 2026-03-27T13:03:39Z
updated_at: 2026-03-27T13:03:39Z
parent: hearts-8j8z
---

In a 1000-game sample, the random bot successfully shot the moon 158 times against the hard bot in lost games. The hard bot's detectMoonShooter logic requires 14+ penalty points AND 4+ tricks for a strong signal, which is too late against accidental shooters who stumble into moon runs.

## Problem

The current detection thresholds (hard.go detectMoonShooter) are tuned for deliberate moon-shot attempts. Random and easy bots can accidentally accumulate all penalty points without triggering detection until it is too late to block.

Early signal (3+ penalties, 3+ tricks, one opponent won every penalty trick) should catch some cases, but the "won every penalty trick" requirement is strict — an accidental shooter may split penalty wins across tricks in ways that don't trigger this.

## Proposed Fix

1. Lower the strong signal threshold: if one opponent has 10+ penalty points and 3+ tricks completed, start blocking.
2. Add a proportional signal: if one opponent holds more than 75% of all scored penalty points, treat as a potential shooter regardless of trick count.
3. When blocking, prioritize leading hearts to force penalty distribution rather than waiting for safe-high-card leads.

## Validation

You MUST follow the Validation section in the parent epic (hearts-8j8z):
- Establish a baseline with a single 250k run first.
- Measure each change with a 50k run. An improvement must be at least 1pp to keep.
- Run simulations sequentially — never in parallel. Wait for each run to complete before proceeding.
- If a change does not meet the threshold, revert it and try something else.
