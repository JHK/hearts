---
# hearts-gev6
title: Centralize card size with target/minimum constants
status: todo
type: task
priority: normal
tags:
    - frontend
    - ux
created_at: 2026-04-27T13:21:29Z
updated_at: 2026-04-27T13:21:37Z
---

Replace hardcoded card pixel sizes with two CSS custom properties (target + minimum) so cards scale up on desktop while preserving mobile sizing.

## Context

Card sizes in `internal/webui/assets/styles.input.css` are hardcoded as literal pixel values scattered across multiple selectors:

- `.hand-card` — 56px (non-bottom seats), 68px (bottom seat, with a `clamp()` rule under 490px viewport)
- `.play-card` (trick center) — 74px
- `.trick-slot` — 76×110px
- `.back-card` — 22×32px

On desktop these look noticeably too small relative to the available table space. Mobile sizing is acceptable and must be preserved.

## Higher Goal

A single source of truth for card sizing makes future visual tuning (animations, layouts, new viewports) a one-line change instead of a hunt across the stylesheet, and unblocks making the table feel appropriately sized on desktop.

## Acceptance Criteria

- [ ] Two CSS custom properties define card sizing: a target size (used on desktop / wide viewports) and a minimum size (used on narrow / mobile viewports).
- [ ] All card-related dimensions (`.hand-card`, `.play-card`, `.trick-slot`, `.back-card`, related margin/overlap values) are derived from those two constants — no hardcoded card pixel values remain.
- [ ] Cards are visibly larger on desktop than today.
- [ ] Mobile (≤490px and the intermediate breakpoints) renders at sizes equivalent to today — no visual regression on small viewports.
- [ ] `design-system.md` updated to mention the card-size constants as the canonical knob.

## Out of Scope

- Redesigning card visuals (faces, backs, shadows, animations).
- Changing seat layout or trick-center geometry beyond what falls out of larger cards.
- Making card size user-configurable in settings.
