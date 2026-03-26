---
# hearts-bzvb
title: Sound Toggle
status: completed
type: feature
priority: normal
created_at: 2026-03-25T09:45:25Z
updated_at: 2026-03-25T10:31:14Z
parent: hearts-5ceo
---

## Context
The game plays sound effects (hearts breaking, queen of spades) but there is no way to mute or control volume. Players in shared spaces or who simply prefer silence have to mute the browser tab entirely.

## Higher Goal
Part of the In-Game User Settings epic — basic audio control is expected in any game with sound.

## Acceptance Criteria
- [x] Player can mute/unmute game sounds
- [x] Sounds are on by default
- [x] Setting persists in localStorage across sessions
- [x] Setting is accessible from an in-game settings UI element

## Out of Scope
- Volume slider (mute toggle only)
- Per-sound-effect controls
- Background music

## Summary of Changes

Added a Sound toggle to the existing settings panel. Sound is on by default and the preference persists in `localStorage` under key `hearts.sound.enabled`. The `audio.js` module gained a `setMuted()` export that gates both `playHeartsBreaking()` and `playQueenOfSpades()`. Files changed: `audio.js`, `main.js`, `dom.js`, `table.html`.
