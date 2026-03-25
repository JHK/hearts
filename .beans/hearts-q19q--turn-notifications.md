---
# hearts-q19q
title: Turn Notifications
status: todo
type: feature
priority: normal
created_at: 2026-03-25T09:47:49Z
updated_at: 2026-03-25T09:48:56Z
parent: hearts-5ceo
---

## Context
In multiplayer games, players often tab away while waiting for their turn. There's no way to know it's your turn without checking the tab manually.

## Higher Goal
Part of the In-Game User Settings epic — reducing friction in multiplayer by notifying players when action is needed.

## Acceptance Criteria
- [ ] Player can enable/disable browser notifications for their turn
- [ ] Disabled by default
- [ ] When enabled, a browser notification fires when it becomes the player's turn to play or pass
- [ ] Browser permission is requested only when the player enables the setting
- [ ] Setting persists in localStorage across sessions
- [ ] Setting is accessible from an in-game settings UI element
- [ ] No notification if the tab is already focused

## Out of Scope
- Push notifications (service worker / offline)
- Sound-only notification (covered by sound toggle)
- Notifications for non-turn events (e.g. player joined, game over)
