---
# hearts-iij0
title: Extend felt to section edges and square up the table
status: todo
type: task
priority: normal
tags:
    - frontend
    - ux
created_at: 2026-03-28T15:21:08Z
updated_at: 2026-03-28T15:21:21Z
parent: hearts-dfll
---

Remove padding/margins between felt and outer chrome so felt fills the full section, and adjust aspect ratio to be squarer like a real table. Simplifies HTML structure and CSS.

## Context

The table felt (`table-board`) sits inside `trickSection` with 12px section padding plus 20px main padding, creating visible gaps between the felt and the outer chrome. The felt also has its own border and border-radius, adding redundant visual layers. The result is a padded, rounded rectangle that feels heavy rather than lean.

## Higher Goal

Part of the table UI redesign (hearts-dfll). The table should feel like a real card table — a clean, edge-to-edge felt surface with a squarer aspect ratio.

## Acceptance Criteria

- [ ] Felt background extends to the edges of the outer visual element (no visible gap between felt and section/page chrome)
- [ ] Redundant borders, padding, and border-radius between section and felt are removed or consolidated
- [ ] Table aspect ratio is closer to square (real table proportions) — adjust grid column ratios or min-height as needed
- [ ] HTML structure is simplified where wrappers become unnecessary
- [ ] Mobile responsive behavior preserved (single-column layout at ≤700px, full-bleed at ≤490px)
- [ ] Design system doc updated if visual constants change

## Out of Scope

- Seat layout or card rendering changes
- Header/controls redesign (separate ticket)
- Trick center controls integration
