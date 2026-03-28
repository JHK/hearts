---
# hearts-bgxb
title: Adjust lobby floating card opacity and background
status: todo
type: task
priority: normal
tags:
    - frontend
created_at: 2026-03-28T13:20:35Z
updated_at: 2026-03-28T13:20:43Z
---

Make card backing solid white (no opacity) and bump card image opacity to 0.25

## Context

The floating card background on the lobby page (`.card-bg-card` in `styles.input.css`) currently uses a translucent card image (`opacity: 0.13`) over the page background. The visual intent has shifted — the backing should be solid white and the card image more visible.

## Higher Goal

Refine the lobby visual polish established in the animated card background work.

## Acceptance Criteria

- [ ] `.card-bg-card .card-bg-backing` background is solid white with no opacity
- [ ] `.card-bg-card img` opacity is `0.25`
- [ ] Tailwind CSS recompiled and verified in browser

## Out of Scope

- Changing card size, drift animation, or grayscale filter
- Dark mode considerations
