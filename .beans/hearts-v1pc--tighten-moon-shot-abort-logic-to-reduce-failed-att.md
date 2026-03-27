---
# hearts-v1pc
title: Tighten moon-shot abort logic to reduce failed attempts
status: todo
type: task
created_at: 2026-03-27T13:03:29Z
updated_at: 2026-03-27T13:03:29Z
parent: hearts-8j8z
---

The hard bot commits to moon shots too aggressively and fails to abort early enough. In a 1000-game sample, 202 failed moon attempts (20-25 raw points) appeared in 650 lost games — making this the single largest source of catastrophic rounds.

## Problem

The current abort logic (hard.go) only aborts when RoundPoints[mySeat] < totalPenalty — i.e., when the bot has lost penalty points to other players. But this check happens too late: by the time another player wins a penalty trick, the bot has already committed heavily and often eats 20+ points.

## Proposed Fix

Add earlier abort signals:
1. Trick-count heuristic: If the bot hasn't won the last N consecutive tricks (e.g. 2), and the moon-shot was based on relaxed thresholds, abort early.
2. Point-gap check: If the bot's round points are high (e.g. 15+) but not 26, and there are few tricks remaining (3 or fewer), check whether completing the moon is still achievable given remaining cards.
3. Opponent void awareness: If an opponent is known to be void in a suit the bot needs to lead, the moon-shot is likely doomed.

## Validation

You MUST follow the Validation section in the parent epic (hearts-8j8z):
- Establish a baseline with a single 250k run first.
- Measure each change with a 50k run. An improvement must be at least 1pp to keep.
- Run simulations sequentially — never in parallel. Wait for each run to complete before proceeding.
- If a change does not meet the threshold, revert it and try something else.
