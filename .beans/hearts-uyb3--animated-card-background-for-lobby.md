---
# hearts-uyb3
title: Animated card background for lobby
status: todo
type: task
priority: normal
tags:
    - frontend
created_at: 2026-03-26T11:46:24Z
updated_at: 2026-03-26T11:46:40Z
parent: hearts-14cx
---

Add a background of spread-out, slowly drifting, dimmed playing cards using existing SVG card assets

## Context
The lobby currently has a plain radial gradient background (white → light blue). A background of spread-out, slowly drifting, dimmed playing cards would give the first screen players see more visual character.

## Higher Goal
Part of the lobby beautification effort (hearts-14cx) to make the lobby feel more like a card game.

## Acceptance Criteria
- [ ] The lobby background shows spread-out playing cards using the existing SVG card assets from `assets/cards/`
- [ ] Cards drift slowly (CSS or JS animation) — subtle, not distracting
- [ ] Cards are dimmed/faded so lobby content remains clearly readable
- [ ] Animation performance is acceptable (no jank on typical hardware)
- [ ] Background doesn't interfere with the responsive layout

## Out of Scope
- New card artwork — reuses existing SVGs
- Interactive/clickable background cards
- Background on the table/game page (lobby only)
