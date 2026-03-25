---
# hearts-svyn
title: Improve hard bot moonshot strategies for measurable win-rate gain
status: completed
type: task
created_at: 2026-03-25T18:49:56Z
updated_at: 2026-03-25T18:49:56Z
parent: hearts-8j8z
blocked_by:
    - hearts-2nme
---

Implement multiple moonshot strategy improvements in the hard bot to differentiate it from medium.

## Acceptance Criteria
- [x] Hard bot has distinct moonshot evaluation, pass, lead, and follow strategies
- [x] Moonshot success rate measurably increases vs medium baseline
- [x] Win rate does not decrease vs baseline
- [x] 500k+ iteration sim confirms improvement
- [x] Comprehensive tests cover all new strategies

## Summary of Changes

Five moonshot improvements in the hard bot:

1. **Relaxed evaluation threshold** (`hardEvaluateMoonShot`): Triggers moonshot with 2+ guaranteed heart tricks and 7+ total (medium requires 3 hearts / 8 total). EV analysis shows moonshot is profitable even at ~31% success rate.

2. **Void-aware moonshot pass** (`hardChooseMoonShotPass`): Prioritizes passing all cards from short non-run off-suits to create voids, giving the bot discard flexibility during play.

3. **Non-heart-first moonshot lead** (`hardMoonShotLead`): Exhausts off-suit safe leads before hearts. Opponents who become void in off-suits discard hearts into the bot's winning tricks. Prefers the longest suit to drain opponents' holdings.

4. **Position-aware moonshot follow** (`hardMoonShotFollow`): Plays highest card when penalty is at stake and not last to play (maximizes win chance). Plays lowest winning card when last (guaranteed win, save resources).

5. **Soft dynamic re-activation** (`countNearSafeCards`): Re-activates moonshot when safe cards are one short of remaining tricks but near-safe cards (one gap from top) fill the gap.

### Sim results (500k games)
- Moonshots: 9108 vs 6420 medium (+42%)
- Win rate: 37.1% vs 36.6% medium (+0.5pp)
- Baseline hard was 37.0% vs 36.7% medium (+0.3pp)

### Blocked by
hearts-2nme (penalty ownership abort fix, already completed)
