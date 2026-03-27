---
# hearts-6vk2
title: Trick center becomes the button background
status: todo
type: task
created_at: 2026-03-27T09:21:44Z
updated_at: 2026-03-27T09:21:44Z
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

- [ ] `.trick-center-controls` panel styling removed (no separate background, border, shadow)
- [ ] The trick center area itself becomes the visual host for buttons
- [ ] Start, Continue, and Pass 3 Cards buttons render cleanly in the center without a floating panel
- [ ] Trick animation layer still works correctly (cards played to center)
- [ ] Follows design system (hearts-8ivt)

## Out of Scope

- Pause button changes (separate bean)
- Start button label logic (separate bean hearts-iys7)
- Scoreboard or seat changes
