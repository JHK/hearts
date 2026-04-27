---
# hearts-9mss
title: Add subtle drop shadow to cards in hand for better distinguishability
status: completed
type: task
priority: normal
tags:
    - frontend
    - ux
created_at: 2026-04-27T14:43:33Z
updated_at: 2026-04-27T14:51:09Z
---

Cards in the player's hand blend into each other; a slight drop shadow would visually separate adjacent cards.

## Context
Cards in the player's hand sit close together (overlapping or near-flush), and the white card faces with thin borders make adjacent cards visually merge. This makes it harder to scan the hand quickly and distinguish individual cards — especially within the same suit.

## Higher Goal
Improve readability of the hand so players can parse their cards at a glance. Part of the broader goal of polishing the in-game visual experience.

## Acceptance Criteria
- [x] Cards in the hand render with a subtle drop shadow that visually separates them from neighbors
- [x] Shadow is subtle — does not dominate the card or clash with the table felt background
- [x] Shadow remains consistent across hover/selected/disabled states (or is intentionally varied per state)
- [x] `design-system.md` updated if a new shadow token is introduced

## Out of Scope
- Redesigning card faces, borders, or spacing
- Shadows on cards in the trick area or pass tray (unless trivially shared via the same component)
- Animation changes to card lift/select

## Summary of Changes

- `.hand-card-image` drop-shadow: added a small leftward x-offset (`-1px`) and slightly stronger opacity (`0.22` → `0.32`), darker tone (`rgba(15, 36, 60, …)` → `rgba(8, 22, 40, …)`). The leftward bias matters because hand cards stack right-over-left in the fan, so the exposed left edge of each card now casts a visible shadow onto the card behind it.
- `.hand-card:hover .hand-card-image` updated to keep the same directional bias and proportionally deeper shadow on hover.
- `.hand-card.selected` blue glow left as-is — it stacks on top of the base image shadow and remains the dominant signal for selection.
- `design-system.md` Play Cards section updated to describe the intent of the shadow (separating adjacent cards in the fan) and how hover/selection vary it. No new CSS token introduced; tweak is inline on existing rules.
