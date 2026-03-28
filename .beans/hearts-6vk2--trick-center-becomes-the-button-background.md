---
# hearts-6vk2
title: Trick center becomes the button background
status: completed
type: task
priority: normal
created_at: 2026-03-27T09:21:44Z
updated_at: 2026-03-28T11:20:28Z
parent: hearts-dfll
blocked_by:
    - hearts-qpu9
---

Remove the separate floating panel for trick center controls. The entire trick center area itself serves as the background — buttons (Start, Continue, Pass 3 Cards, etc.) fill the center without their own card/panel element.

## Context

Currently `.trick-center-controls` is a dark semi-transparent panel with backdrop blur, positioned absolutely at the center of the board. It's a visual element layered on top of the trick center. The goal is to eliminate this layering: the trick center *is* the control surface. When a button needs to show, the trick center background hosts it directly — no extra panel, no extra border radius, no extra shadow.

## Higher Goal

Reduce visual clutter in the game board. Controls feel integrated, not stacked.

## Acceptance Criteria

- [x] `.trick-center-controls` panel styling removed (no separate background, border, shadow)
- [x] The trick center area itself becomes the visual host for buttons
- [x] Start, Continue, and Pass 3 Cards buttons render cleanly in the center without a floating panel
- [x] Trick animation layer still works correctly (cards played to center)
- [x] Follows design system (hearts-8ivt)

## Out of Scope

- Pause button changes (separate bean)
- Start button label logic (separate bean hearts-iys7)
- Scoreboard or seat changes

## Summary of Changes

Removed panel chrome (background, box-shadow, backdrop-filter, border-radius, fixed width) from `.trick-center-controls`. The controls now fill the trick center with `inset: 0` and use flexbox centering, making the trick center itself the visual host for buttons.

Refined button colors: teal is the standard button color across lobby, table, and overlays. Felt buttons (trick center) use a `.felt-btn` class with a darkened gold three-stop gradient for contrast against the green felt and legibility with white text. Design system updated with structured button type documentation.
