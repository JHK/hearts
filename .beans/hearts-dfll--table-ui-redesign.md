---
# hearts-dfll
title: Table UI redesign
status: completed
type: epic
priority: normal
created_at: 2026-03-27T08:52:48Z
updated_at: 2026-03-28T19:22:34Z
---

Visual overhaul of the game table page to match the modern, lean aesthetic established by the lobby redesign (hearts-c65d). The table should feel lighter: less padding, simplified header, integrated controls, and consistent use of icon buttons with dropdowns.

## Goals

- Establish and document a shared visual design system for the whole game
- Simplify the table chrome: minimal header, no extra padding
- Consolidate action buttons (Add Bot, Take Seat) into compact header icons with dropdowns
- Make the Start button smarter and always visible
- Integrate trick center controls into the center area itself (no separate floating panel)
- Replace the pause overlay with an inline center button

## Out of Scope

- Lobby changes (already done in hearts-c65d)
- Game Over overlay (already correct behavior)
- Gameplay logic changes
