---
# hearts-bwkd
title: Auto-Play Single Legal Card
status: todo
type: feature
priority: normal
created_at: 2026-03-25T09:46:55Z
updated_at: 2026-03-25T09:47:10Z
parent: hearts-5ceo
---

## Context
When a player has only one legal card to play, they still must click it manually. This adds unnecessary friction, especially in late tricks and endgame when hands are small.

## Higher Goal
Part of the In-Game User Settings epic — an opt-in convenience feature to reduce tedious clicks.

## Acceptance Criteria
- [ ] Player can enable/disable auto-play of the only legal card
- [ ] Disabled by default
- [ ] When enabled, the card is played automatically after a brief visual delay (so the player sees what happened)
- [ ] Setting persists in localStorage across sessions
- [ ] Setting is accessible from an in-game settings UI element

## Out of Scope
- Auto-play when multiple legal cards exist
- Auto-pass (pass phase automation)
