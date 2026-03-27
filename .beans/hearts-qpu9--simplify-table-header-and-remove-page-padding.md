---
# hearts-qpu9
title: Simplify table header and remove page padding
status: completed
type: task
priority: normal
created_at: 2026-03-27T09:06:44Z
updated_at: 2026-03-27T13:53:38Z
parent: hearts-dfll
blocked_by:
    - hearts-8ivt
---

Strip the table page down to a leaner layout: remove outer padding so the board extends closer to the edges, and simplify the header to just a title that blends into the background without a visible panel or border.

## Context

The lobby redesign (hearts-c65d) established a clean, minimal aesthetic. The table page still has generous padding and a distinct header panel that feels heavier than necessary. This is the first bean applying the design system (hearts-8ivt) to the table page — it sets the visual tone for the remaining beans in the epic.

## Higher Goal

Make the table page feel as modern and lean as the redesigned lobby.

## Acceptance Criteria

- [x] Table page outer padding removed or significantly reduced — the board fills more of the viewport
- [x] Header has no visible background, border, or shadow — title text sits directly on the page background
- [x] Settings gear and back-to-lobby link remain functional and accessible in the header
- [x] Styling follows the design system documented in hearts-8ivt
- [x] Responsive: still usable on mobile viewports
- [x] Visual regression check: no layout breakage on the board, seats, or scoreboard

## Out of Scope

- Moving buttons into the header (that's separate beans)
- Trick center or control changes
- Scoreboard styling


## Summary of Changes

- Removed `<section>` wrapper from table header so it no longer inherits panel background, border, or shadow
- Reduced `.table-page main` padding from 16px to 6px so the board fills more of the viewport
- Added light padding directly to `.page-header` for breathing room without a visible panel
- Preserved 490px breakpoint behavior (zero main padding, header still has minimal padding)
- Settings gear and back-to-lobby link unchanged and functional
