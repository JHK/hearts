---
# hearts-o4y0
title: Move Take Seat into header as icon button with dropdown
status: todo
type: task
created_at: 2026-03-27T09:12:15Z
updated_at: 2026-03-27T09:12:15Z
parent: hearts-dfll
blocked_by:
    - hearts-qpu9
---

Replace the inline "Take seat" buttons on individual bot seats with a compact 🪑 icon button in the header bar. Clicking it opens a dropdown listing seats currently occupied by bots (i.e. claimable). Selecting a seat claims it.

## Context

Currently each bot seat renders a small inline "Take seat" button next to the player name. This clutters the seat display and is inconsistent with the header-icon pattern being established for Add Bot (hearts-gfmc) and settings. Moving it into the header as a matching icon button keeps the board clean and the controls consistent.

## Higher Goal

Consistent, compact header controls — all table-level actions live in the header as icon buttons.

## Acceptance Criteria

- [ ] 🪑 icon button in the header, visually consistent with the settings gear and 🤖 icon
- [ ] Clicking opens a dropdown listing bot-occupied seats (seat name or position)
- [ ] Selecting a seat claims it (same behavior as current inline "Take seat" buttons)
- [ ] Icon only visible when the player is an observer and there are bot seats to claim
- [ ] Old inline "Take seat" buttons removed from seat displays
- [ ] Follows design system (hearts-8ivt)

## Out of Scope

- Add Bot button (separate bean hearts-gfmc)
- Seat assignment logic changes
- Start button changes
