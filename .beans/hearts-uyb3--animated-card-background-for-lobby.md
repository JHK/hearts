---
# hearts-uyb3
title: Animated card background for lobby
status: completed
type: task
priority: normal
tags:
    - frontend
created_at: 2026-03-26T11:46:24Z
updated_at: 2026-03-27T13:56:10Z
parent: hearts-14cx
---

Add a background of spread-out, slowly drifting, dimmed playing cards using existing SVG card assets

## Context
The lobby currently has a plain radial gradient background (white → light blue). A background of spread-out, slowly drifting, dimmed playing cards would give the first screen players see more visual character.

## Higher Goal
Part of the lobby beautification effort (hearts-14cx) to make the lobby feel more like a card game.

## Acceptance Criteria
- [x] The lobby background shows spread-out playing cards using the existing SVG card assets from `assets/cards/`
- [x] Cards drift slowly (CSS or JS animation) — subtle, not distracting
- [x] Cards are dimmed/faded so lobby content remains clearly readable
- [x] Animation performance is acceptable (no jank on typical hardware)
- [x] Background doesn't interfere with the responsive layout

## Out of Scope
- New card artwork — reuses existing SVGs
- Interactive/clickable background cards
- Background on the table/game page (lobby only)

## Summary of Changes

- Added a `#cardBg` container div to `index.html` with `aria-hidden="true"`
- Added CSS: fixed-position background layer at z-index 0, 80px card images at 7% opacity with 40% grayscale, `card-drift` keyframe animation using CSS custom properties for per-card drift vectors
- Added JS: on load, shuffles the full deck and places 22 randomly positioned/rotated cards with alternating drift animations (30-70s duration) for a gentle, non-distracting effect
- `pointer-events: none` ensures the background never interferes with interaction
- Uses `will-change: transform` and CSS animations (not JS rAF) for GPU-composited, jank-free performance
