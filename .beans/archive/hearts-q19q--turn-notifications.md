---
# hearts-q19q
title: Turn Notifications
status: completed
type: feature
priority: normal
created_at: 2026-03-25T09:47:49Z
updated_at: 2026-03-25T10:44:03Z
parent: hearts-5ceo
---

## Context
In multiplayer games, players often tab away while waiting for their turn. There's no way to know it's your turn without checking the tab manually.

## Higher Goal
Part of the In-Game User Settings epic — reducing friction in multiplayer by notifying players when action is needed.

## Acceptance Criteria
- [x] Player can enable/disable browser notifications for their turn
- [x] Disabled by default
- [x] When enabled, a browser notification fires when it becomes the player's turn to play or pass
- [x] Browser permission is requested only when the player enables the setting
- [x] Setting persists in localStorage across sessions
- [x] Setting is accessible from an in-game settings UI element
- [x] No notification if the tab is already focused

## Out of Scope
- Push notifications (service worker / offline)
- Sound-only notification (covered by sound toggle)
- Notifications for non-turn events (e.g. player joined, game over)


## Summary of Changes

Added a "Turn notifications" toggle to the in-game settings panel. When enabled, browser notifications fire on `your_turn` (play phase) and `pass_submitted` with count=0 (start of pass phase). Notifications are disabled by default, permission is requested only on enable, and no notification fires if the tab is focused. Setting persists via `hearts.notifications.enabled` in localStorage.
