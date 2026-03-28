---
# hearts-o4y0
title: Move Take Seat into header as icon button with dropdown
status: completed
type: task
priority: normal
created_at: 2026-03-27T09:12:15Z
updated_at: 2026-03-28T15:47:31Z
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

- [x] 🪑 icon button in the header, visually consistent with the settings gear and 🤖 icon
- [x] Clicking opens a dropdown listing bot-occupied seats (seat name or position)
- [x] Selecting a seat claims it (same behavior as current inline "Take seat" buttons)
- [x] Icon only visible when the player is an observer and there are bot seats to claim
- [x] Old inline "Take seat" buttons removed from seat displays
- [x] Follows design system (hearts-8ivt)

## Out of Scope

- Add Bot button (separate bean hearts-gfmc)
- Seat assignment logic changes
- Start button changes

## Summary of Changes

- Added 🪑 icon button in the page header (between back arrow and settings gear) with a dropdown panel
- Dropdown dynamically lists bot-occupied seats by name; clicking one claims that seat
- Button only visible when the viewer is an observer and at least one bot seat exists
- Removed old inline `claim-seat-btn` buttons from seat name labels in render.js
- Removed `claimSeat` parameter from `createRenderer` (no longer needed by renderer)
- Added `.claim-seat-panel` and `.claim-seat-option` CSS classes matching the bot-strength-panel pattern
- Reuses `initSettingsPopover` for consistent open/close/escape behavior
