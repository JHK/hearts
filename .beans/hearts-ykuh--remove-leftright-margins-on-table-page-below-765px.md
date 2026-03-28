---
# hearts-ykuh
title: Remove left/right margins on table page below 765px
status: completed
type: bug
priority: normal
tags:
    - frontend
created_at: 2026-03-28T16:17:10Z
updated_at: 2026-03-28T16:21:54Z
parent: hearts-dfll
---

## Context

The in-game table view has layout issues at viewport widths below 765px. Elements with left/right margins cause content to overflow or compress awkwardly at narrow widths.

## Current Behavior

At viewports narrower than ~765px, the table page UI breaks — elements with horizontal margins consume too much space, causing layout problems.

## Desired Behavior

Below 765px, left/right margins on table page elements are removed so the layout uses the full viewport width.

## Acceptance Criteria

- [x] At 765px and below, table page elements have zero left/right margins
- [x] Layout remains visually correct at common narrow widths (e.g. 375px, 414px, 480px, 768px)
- [x] No regressions at wider viewports — existing margins preserved above the breakpoint
- [x] Design system updated if this introduces a new responsive breakpoint or changes documented spacing

## Out of Scope

- Redesigning the table layout itself — this is just margin removal
- Other responsive issues unrelated to horizontal margins

## Summary of Changes

Split the `max-width: 490px` media query into two:
- **765px**: Removes `main` padding, section border-radius, and table-board horizontal padding — making the table page go edge-to-edge on narrow viewports.
- **490px**: Retains card sizing adjustments only (hand card width clamp, tighter card overlap).

No new CSS properties were added; the existing rules were simply moved to a wider breakpoint.
