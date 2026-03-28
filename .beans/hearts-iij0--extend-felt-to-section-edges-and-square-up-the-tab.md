---
# hearts-iij0
title: Extend felt to section edges and square up the table
status: completed
type: task
priority: normal
tags:
    - frontend
    - ux
created_at: 2026-03-28T15:21:08Z
updated_at: 2026-03-28T15:41:44Z
parent: hearts-dfll
---

Remove padding/margins between felt and outer chrome so felt fills the full section, and adjust aspect ratio to be squarer like a real table. Simplifies HTML structure and CSS.

## Context

The table felt (`table-board`) sits inside `trickSection` with 12px section padding plus 20px main padding, creating visible gaps between the felt and the outer chrome. The felt also has its own border and border-radius, adding redundant visual layers. The result is a padded, rounded rectangle that feels heavy rather than lean.

## Higher Goal

Part of the table UI redesign (hearts-dfll). The table should feel like a real card table — a clean, edge-to-edge felt surface with a squarer aspect ratio.

## Acceptance Criteria

- [x] Felt background extends to the edges of the outer visual element (no visible gap between felt and section/page chrome)
- [x] Redundant borders, padding, and border-radius between section and felt are removed or consolidated
- [x] Table aspect ratio is closer to square (real table proportions) — adjust grid column ratios or min-height as needed
- [x] HTML structure is simplified where wrappers become unnecessary
- [x] Mobile responsive behavior preserved (single-column layout at ≤700px, full-bleed at ≤490px)
- [x] Design system doc updated if visual constants change

## Out of Scope

- Seat layout or card rendering changes
- Header/controls redesign (separate ticket)
- Trick center controls integration

## Summary of Changes

- Removed section chrome (padding, border, background, box-shadow) from `#trickSection` so the felt gradient extends edge-to-edge
- Removed redundant `1px solid #2d7e6f` border from `.table-board` — the felt's border-radius now serves as the only visual boundary
- Squarer grid proportions: reduced center column from `minmax(300px, 1fr)` to `minmax(260px, 1fr)` and side columns from `220px` to `200px`
- Reduced large-screen game column minimum from `860px` to `780px`
- Simplified ≤490px breakpoint — removed now-redundant `#trickSection` override and border-side resets
