---
# hearts-9mss
title: Add subtle drop shadow to cards in hand for better distinguishability
status: todo
type: task
priority: normal
tags:
    - frontend
    - ux
created_at: 2026-04-27T14:43:33Z
updated_at: 2026-04-27T14:43:43Z
---

Cards in the player's hand blend into each other; a slight drop shadow would visually separate adjacent cards.

## Context
Cards in the player's hand sit close together (overlapping or near-flush), and the white card faces with thin borders make adjacent cards visually merge. This makes it harder to scan the hand quickly and distinguish individual cards — especially within the same suit.

## Higher Goal
Improve readability of the hand so players can parse their cards at a glance. Part of the broader goal of polishing the in-game visual experience.

## Acceptance Criteria
- [ ] Cards in the hand render with a subtle drop shadow that visually separates them from neighbors
- [ ] Shadow is subtle — does not dominate the card or clash with the table felt background
- [ ] Shadow remains consistent across hover/selected/disabled states (or is intentionally varied per state)
- [ ] `design-system.md` updated if a new shadow token is introduced

## Out of Scope
- Redesigning card faces, borders, or spacing
- Shadows on cards in the trick area or pass tray (unless trivially shared via the same component)
- Animation changes to card lift/select
