---
# hearts-v1pc
title: Tighten moon-shot abort logic to reduce failed attempts
status: scrapped
type: task
priority: normal
created_at: 2026-03-27T13:03:29Z
updated_at: 2026-03-27T15:59:52Z
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


## Reasons for Scrapping

Exhaustive testing (15+ approaches across 50k-run simulations) failed to produce the required 1pp improvement. Baseline: hard bot at 36.9% win rate (250k run).

### Approaches tested

**Threshold tightening (entry reduction):**
- Remove relaxed 2h+7t path entirely → 36.7% (neutral)
- Tighten trailing-by-30 from 1h+6t to 2h+6t/2h+7t → 37.0%, 36.9% (noise)
- Score-aware thresholds (strict when leading, relaxed when trailing) → 36.5% (worse)
- Lower near-game-over from 85 to 75 → 36.7% (neutral)

**Abort logic improvements:**
- Mid-game coverage abort (safe+near-safe < remaining tricks after 4 tricks) → 36.4% (worse)
- Combined coverage abort + threshold tightening → 36.4% (worse)
- Remove soft re-activation → 36.5% (worse)

**Moon-shot play improvements:**
- Void-aware leads (avoid suits where opponents are void) → 37.1%, 36.7% (noise)
- Hearts-first leads (collect hearts early) → 36.9% (neutral)
- Aggressive follow (play highest when not last) → 37.1% (noise when combined)
- Conservative early leads (defensive when no safe cards) → 36.6% (neutral)

**Non-moon-shot improvements:**
- Lower MC activation threshold (hand ≤ 9) → 33.7% (much worse — 3 MC samples too few)
- Disable moon blocking → 36.9% (neutral — blocking is net-zero)
- Void-aware defensive follow → 36.8% (neutral)

### Key findings

1. **The moon-shot system is near-optimal for this architecture.** Reducing failed attempts also reduces successful ones proportionally. Successful moons (26→0 pts) are extremely valuable and offset the cost of failures.
2. **Mid-round abort doesn't save enough.** By the time the bot could abort (3-4 tricks in), it already holds AH/KH/QS and will win penalty tricks even in defensive mode.
3. **The 50k confidence interval is ~±0.43pp** (SE ≈ 0.22% at 37%). All results fell within noise of the 36.9% baseline. A genuine 1pp improvement would be detectable, but no approach produced one.
4. **The trailing-by-30 path (1h+6t) appears net-positive** despite being very relaxed — removing it hurts.
5. **Moon blocking is net-neutral** — disabling it doesn't change win rate, suggesting the blocking code's benefit and cost are balanced.
