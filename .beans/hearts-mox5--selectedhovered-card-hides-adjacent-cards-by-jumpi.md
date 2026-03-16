---
# hearts-mox5
title: Selected/hovered card hides adjacent cards by jumping to front of stack
status: todo
type: bug
priority: normal
created_at: 2026-03-16T08:17:01Z
updated_at: 2026-03-16T08:17:12Z
---

## Context
When a card is hovered or selected, its `z-index` is raised (`z-index: 3` on hover, `z-index: 4` on selected). Because cards overlap with negative margins, this causes the active card to render on top of its neighbors, hiding parts of adjacent cards. The intent is just to lift the card slightly upward — not to bring it to the front of the visual stack.

## Higher Goal
Players should be able to see their full hand at all times. A selected card should feel "picked up" vertically, not pulled forward out of the deck.

## Acceptance Criteria
- [ ] Hovering a card lifts it upward without covering adjacent cards
- [ ] Selecting a card lifts it upward without covering adjacent cards
- [ ] The natural left-to-right stacking order of the hand is preserved regardless of hover/selection state

## Out of Scope
- Changing the card overlap amount or hand layout
- Changing the lift distance (`translateY` values)
