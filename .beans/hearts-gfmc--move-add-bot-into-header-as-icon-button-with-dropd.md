---
# hearts-gfmc
title: Move Add Bot into header as icon button with dropdown
status: todo
type: task
created_at: 2026-03-27T09:09:08Z
updated_at: 2026-03-27T09:09:08Z
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

- [ ] 🤖 icon button in the header, visually consistent with the settings gear
- [ ] Clicking opens a dropdown with bot strength options (easy, medium, hard)
- [ ] Selecting a strength adds a bot at that level (same behavior as current "Add Bot")
- [ ] Icon visible when bots can be added: empty seats exist, or a player left and their seat can be filled
- [ ] Icon hidden when no seats are available for bots
- [ ] Old bot control panel removed from the board
- [ ] Follows design system (hearts-8ivt)

## Out of Scope

- Changing bot behavior or strategies
- "Take Seat" button (separate bean)
- Start button changes
