---
# hearts-p4vl
title: Pause as inline center button instead of full overlay
status: completed
type: task
priority: normal
created_at: 2026-03-27T09:27:01Z
updated_at: 2026-03-28T14:06:01Z
parent: hearts-dfll
blocked_by:
    - hearts-6vk2
---

Replace the full-screen pause overlay with a simple button in the trick center, visually consistent with Start/Continue. Only "Game Over" remains a full-page overlay.

## Context

Currently `.game-paused-overlay` is a fixed, full-screen dark backdrop with a centered white panel. This is heavy-handed for a pause state. The new direction (per the trick center bean hearts-6vk2) is that all in-game state buttons live in the trick center without overlays. Pause should follow the same pattern — just a "Resume" or "Paused" button in the center, matching the visual weight of Start and Continue.

## Higher Goal

Consistent control patterns — overlays reserved for terminal states (Game Over) only.

## Acceptance Criteria

- [x] Pause state shows a button in the trick center (e.g. "Resume" or "Game Paused") instead of a full-screen overlay
- [x] No backdrop blur or screen-dimming for pause
- [x] Button visually consistent with Start/Continue buttons in the trick center
- [x] Game board remains visible behind the pause button
- [x] Game Over overlay unchanged — still a full-page overlay
- [x] Follows design system (hearts-8ivt)

## Out of Scope

- Game Over overlay changes
- Pause logic or state machine changes
- Trick center layout (handled by hearts-6vk2)

## Summary of Changes

Replaced the full-screen pause overlay with an inline control inside the trick center (`centerControls`). The pause state now shows a message and "Continue with Bot" button using the same `felt-btn` style as Start/Continue, consistent with the pattern established in hearts-6vk2. Removed the `.game-paused-panel`, `.game-paused-message`, and `.game-paused-actions` CSS classes. Game Over overlay is unchanged.
