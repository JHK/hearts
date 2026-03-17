---
# hearts-uet2
title: Fix trick collection animation positioning and scaling
status: done
type: bug
priority: normal
created_at: 2026-03-15T16:38:50Z
updated_at: 2026-03-17T13:08:20Z
---

## Context

The "get the trick" animation runs after every completed trick, animating card clones from the trick center slots to the winner's seat. Affects all players every hand.

## Current Behavior

Three positioning bugs, all consistent and happening every trick:

1. **Wrong start position** — the animated card clones don't originate from where the cards visually rest in the trick center slots; they appear to start from a different position.
2. **Wrong end position** — cards don't fly to the winner's name label; they land somewhere nearby but not on it.
3. **Cards spread into a vertical row at the destination** — all four cards should stack on the same spot (with a fan/rotation effect), but instead they are offset from each other at the end.

## Desired Behavior

- Each card clone starts at the exact pixel position where its trick-slot card is rendered.
- All cards converge on the center of the trick winner's name label.
- Cards stack at the same destination point; only rotation differs between cards (fan effect).
- Cards shrink as they travel, arriving at the same size as opponent hand cards.

## Acceptance Criteria

- [x] Animated card clones visually originate from their corresponding trick slot card positions
- [x] All cards converge on the center of the winner's seat name element
- [x] Cards stack at the same destination point; only rotation differs between cards
- [x] Cards arrive at a size visually matching opponent hand cards (consistent with the `scale()` value used for hand cards)

## Out of Scope

- Changing animation duration, easing, or stagger timing
- The winner seat pulse animation (`trick-winner-pulse`)
- `prefers-reduced-motion` behavior (cards are skipped entirely there)
