---
# hearts-qpu9
title: Simplify table header and remove page padding
status: todo
type: task
created_at: 2026-03-27T09:06:44Z
updated_at: 2026-03-27T09:06:44Z
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

- [ ] Table page outer padding removed or significantly reduced — the board fills more of the viewport
- [ ] Header has no visible background, border, or shadow — title text sits directly on the page background
- [ ] Settings gear and back-to-lobby link remain functional and accessible in the header
- [ ] Styling follows the design system documented in hearts-8ivt
- [ ] Responsive: still usable on mobile viewports
- [ ] Visual regression check: no layout breakage on the board, seats, or scoreboard

## Out of Scope

- Moving buttons into the header (that's separate beans)
- Trick center or control changes
- Scoreboard styling
