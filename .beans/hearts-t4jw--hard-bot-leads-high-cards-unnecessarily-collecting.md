---
# hearts-t4jw
title: Hard bot leads high cards unnecessarily, collecting avoidable penalty points
status: todo
type: bug
priority: normal
created_at: 2026-03-30T09:28:25Z
updated_at: 2026-03-30T09:39:53Z
parent: hearts-8j8z
---

Defensive lead tiebreaker prefers highest rank, guaranteeing trick wins and penalty exposure when only high cards remain

## Context
The hard bot's defensive leading logic (`smartChooseLead` in `medium.go`) sometimes leads with high cards (J, Q, K, A), which almost guarantees winning the trick and exposes it to penalty dumps from void opponents.

## Current Behavior
When the filtering cascade in `smartChooseLead` can't find low cards (rank < 11 / < 9), the pool falls back to high cards. The `compareDefensiveLeadCard` tiebreaker then prefers the **highest** rank ("shed first" logic at `helpers.go:190`), maximizing trick-win probability and penalty exposure.

Sim data (50-game sample, 960 hard-bot leads):
- 16.9% of leads are J+ rank
- 85% win rate on these leads (vs 15% for low-card leads)
- 1.8 avg penalty pts per won trick (vs 1.3 for low leads, +38%)
- Includes cases like leading Aces into heart dumps and QS onto self for 15pts

## Desired Behavior
When forced to lead a high card, the bot should minimize trick-winning probability by preferring the **lowest** available high card (a Jack can be beaten by Q/K/A; an Ace guarantees winning). The "shed first" rank preference should only apply within the safe low-card pool, not when the pool has fallen through to dangerous high cards.

## Acceptance Criteria
- [ ] When all filtering produces only high cards (rank ≥ 11), lead selection prefers the lowest rank (least likely to win the trick)
- [ ] "Shed first" (highest rank) tiebreaker still applies when pool contains low cards
- [ ] Sim confirms ≥1pp improvement on 50k run over epic baseline (hard: 39.1%)
- [ ] `strategies.md` updated if leading heuristics change

## Out of Scope
- Expanding MC evaluator to earlier tricks (separate improvement)
- Reworking the entire leading filter cascade
- Changes to moon-shot or blocking lead logic


## Approach
Follow the validation approach from the parent epic (hearts-8j8z):
1. **One change at a time**: Make a single improvement, then measure with a 50k sim
2. **Sequential only**: Do NOT run multiple sims in parallel — wait for each to complete
3. **Threshold**: ≥1pp improvement on 50k run to be considered significant
4. **Iterate**: Keep improvements that meet the threshold, revert those that don't
5. **Update baseline**: If a 50k run shows a significant improvement, re-run the 250k baseline and update the table in the epic
