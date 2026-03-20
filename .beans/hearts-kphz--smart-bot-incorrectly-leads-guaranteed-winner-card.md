---
# hearts-kphz
title: Smart bot incorrectly leads guaranteed-winner cards in defensive mode
status: todo
type: bug
priority: normal
created_at: 2026-03-20T11:02:57Z
updated_at: 2026-03-20T11:03:07Z
---

## Context

`smartChooseLead` builds a `safePool` (line 368, `smart.go`) by keeping cards where `rank < 11 || isSafeHighCard(c, hand, playedCards)`. `isSafeHighCard` returns `true` when all higher ranks in the suit are either played or held by the bot — so holding J–A makes even a 10 a guaranteed trick winner. The filter therefore *keeps* guaranteed winners in the defensive lead pool instead of excluding them, inverting the intended safety logic. In defensive mode a card that guarantees winning the trick is a liability, not a safe choice: void opponents will dump the Queen of Spades onto it.

The existing `detectSuitVoids` guard only avoids suits where a void has already been observed, offering no protection on the first lead into an undetected void.

## Current Behavior

In defensive mode the bot may lead a guaranteed-winner card (any rank, not just an ace) early in a round into a suit with an undetected void. Opponents who cannot follow discard the Queen of Spades (13 pts) onto the bot.

## Desired Behavior

In defensive mode the bot should prefer leading cards where at least one higher rank still exists in opponents' hands (i.e. NOT a safe high card per the current definition). Guaranteed-winner leads should be chosen only when no non-winning alternative exists, or when actively pursuing a moon shot.

## Acceptance Criteria
- [ ] Unit tests exist for `smart.ChoosePlay` covering: defensive lead avoids guaranteed-winner cards when lower-ranked alternatives exist; moon-shot lead still chooses the highest guaranteed winner; at least one follow and one discard scenario
- [ ] A baseline sim result (50 k games, current code) is recorded before the fix is applied
- [ ] After the fix, a 50 k sim run shows bot win rate and moon-shot rate no worse than the baseline (within noise)
- [ ] The bot does not lead a guaranteed-winner card in defensive mode when non-winning legal alternatives exist
- [ ] Moon-shot lead and follow/discard behavior is unaffected

## Out of Scope
- Changes to moon-shot evaluation logic or pass strategy
- Improving void detection beyond what `detectSuitVoids` already does
- Other strategic improvements to the smart bot
