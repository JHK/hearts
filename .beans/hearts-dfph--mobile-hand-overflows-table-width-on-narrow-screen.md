---
# hearts-dfph
title: Mobile hand overflows table width on narrow screens
status: completed
type: bug
priority: normal
created_at: 2026-03-17T17:23:59Z
updated_at: 2026-03-18T09:50:44Z
---

Full 13-card hand overflows table width on narrow phones, causing layout glitches until cards are played

## Context
On narrow mobile viewports (≤700px), the bottom-seat hand uses 68px cards with -38px overlap. At 13 cards (game start), that's ~428px — wider than most phones (e.g. 360px). The table board uses a single-column layout on mobile but doesn't constrain or adapt the hand width, causing overflow and layout glitches. The issue self-resolves as cards are played and the hand shrinks.

## Current Behavior
At game start on narrow phones, the 13-card hand overflows the table width, breaking the layout visually. The glitches disappear once enough cards have been played to bring the hand within bounds.

**Steps to reproduce:** Open on a phone (or DevTools at ~360px width), start a game, observe the initial full hand.

## Desired Behavior
The hand fits within the table width at all viewport sizes, with no overflow.

## Implementation Options
Two approaches are viable — left to the implementer to decide:

1. **Scale cards down** — calculate available width and shrink card/overlap dimensions to fit. Simple, but cards may become hard to tap on very narrow screens.
2. **Two-row wrap by suit** — overflow into a second row, grouping by suit. Better tap targets, more structured visual, but more complex layout and interaction code.

## Acceptance Criteria
- [ ] A full 13-card hand renders without overflowing table width at 360px viewport
- [ ] No layout glitches at game start on mobile
- [ ] Cards remain tappable and interactive (play, pass-select)

## Out of Scope
- Opponent card backs (already small, unlikely to overflow)
- Desktop layout changes
