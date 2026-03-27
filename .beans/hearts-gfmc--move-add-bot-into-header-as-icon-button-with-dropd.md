---
# hearts-gfmc
title: Move Add Bot into header as icon button with dropdown
status: completed
type: task
priority: normal
created_at: 2026-03-27T09:09:08Z
updated_at: 2026-03-27T15:15:31Z
parent: hearts-dfll
blocked_by:
    - hearts-qpu9
---

Replace the current top-right bot control panel with a compact 🤖 icon button in the header bar, next to the settings gear. Clicking it opens a dropdown to choose bot strength (easy/medium/hard), then adds a bot at that level.

## Context

Currently the bot control is a green gradient panel positioned absolutely at the top-right of the board, with a button and a select dropdown. It's visually heavy and inconsistent with the lean header direction. Moving it into the header as an icon button with dropdown matches the settings gear pattern and declutters the board.

## Higher Goal

Consistent, compact header controls — all table-level actions live in the header as icon buttons.

## Acceptance Criteria

- [x] 🤖 icon button in the header, visually consistent with the settings gear
- [x] Clicking opens a dropdown with bot strength options (easy, medium, hard)
- [x] Selecting a strength adds a bot at that level (same behavior as current "Add Bot")
- [x] Icon visible when bots can be added: empty seats exist, or a player left and their seat can be filled
- [x] Icon hidden when no seats are available for bots
- [x] Old bot control panel removed from the board
- [x] Follows design system (hearts-8ivt)

## Out of Scope

- Changing bot behavior or strategies
- "Take Seat" button (separate bean)
- Start button changes

## Summary of Changes

Moved the bot control from a green gradient panel on the game board to a compact robot icon button in the page header, next to the settings gear. Clicking the icon opens a dropdown with Easy/Medium/Hard options. Selecting one sends the same `add_bot` WebSocket command. The button auto-hides when no seats are available. Reuses the existing `icon-btn`, `settings-container`, and `initSettingsPopover` patterns for visual and behavioral consistency.
