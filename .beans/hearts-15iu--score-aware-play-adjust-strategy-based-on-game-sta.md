---
# hearts-15iu
title: 'Score-aware play: adjust strategy based on game standings'
status: todo
type: task
priority: normal
tags:
    - backend
created_at: 2026-03-25T19:02:10Z
updated_at: 2026-03-26T09:21:27Z
parent: hearts-8j8z
---

Adjust aggression based on current scores — play riskier when trailing, more conservatively when leading, consider enabling opponent moon-shots to hurt the leader


## Context
The hard bot plays each round in isolation — it doesn't consider cumulative game scores when making decisions. In Hearts, the game ends when someone hits 100+ points, and the lowest score wins. This means optimal strategy shifts based on standings:

- **When leading:** Play conservatively. Avoid risky plays. Don't need to shoot the moon.
- **When trailing badly:** Take more risks. Moon-shots become worth attempting even with marginal hands. Feeding another trailing player a moon-shot can hurt the leader.
- **When mid-pack:** Balance risk and defense depending on proximity to the leader/trailer.

## Higher Goal
Make the hard bot play strategically across the full game arc, not just optimize each round independently.

## Implementation Notes
**Score-aware adjustments:**

1. **Moon-shot threshold adjustment:**
   - When trailing by 30+: lower the threshold for attempting a moon-shot (accept riskier hands)
   - When leading: raise the threshold (only attempt with near-perfect hands)

2. **Defensive intensity:**
   - When leading: maximize avoidance of *any* penalty points, even at cost of giving points to mid-pack players
   - When trailing: accept small penalties if it means avoiding large ones or setting up a future moon-shot

3. **Opponent targeting (advanced):**
   - When one opponent is close to 100: consider feeding them points to end the game (if you're in a winning position)
   - When you're close to 100: maximize avoidance at all costs

4. **Moon-shot enabling:**
   - If a trailing opponent appears to be shooting and you're in the lead, consider *not* blocking — letting them shoot hurts everyone equally (+26) but may keep you in first place while the trailing player catches up to others

**Data needed:** `ChoosePlay`/`ChoosePass` would need access to cumulative game scores, not just the current round state. Check whether `TurnInput`/`PassInput` already includes this or if it needs to be added.

## Acceptance Criteria
- [ ] Hard bot has access to cumulative game scores during decision-making
- [ ] Moon-shot attempt threshold adjusts based on score differential
- [ ] Defensive play intensity adjusts based on standings (leading = more cautious)
- [ ] Benchmark: 50k+ sim iterations before/after; win-rate must not decrease
- Note: existing code/tests may be freely rewritten or removed if a 250k sim shows ≥0.3pp improvement over the previous baseline
- [ ] Edge case: bot near 100 points plays maximally defensively

## Out of Scope
- Complex opponent modeling (predicting what opponents will do based on their scores)
- Cooperative play between bots
- Score display changes in the UI

## References
- [VIP Hearts Tips](https://viphearts.com/blog/hearts-tips/): track who's close to winning, adjust strategy accordingly
- [Mark's Advanced Hearts](https://mark.random-article.com/hearts/advanced.html): pass moon-shooting materials to trailing players to level scores
- [Solitaired Strategy Guide](https://solitaired.com/guides/hearts-strategies-and-tips-to-win): long-game strategy awareness
