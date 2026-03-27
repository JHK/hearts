---
# hearts-5ceo
title: In-Game User Settings
status: completed
type: epic
priority: normal
created_at: 2026-03-25T09:40:37Z
updated_at: 2026-03-27T12:28:31Z
---

## Vision
Players can customize their gameplay experience through an in-game settings panel. Settings persist in the browser across sessions and cover animation pacing, audio, automation shortcuts, and notifications — the controls that make the difference between a pleasant and a tedious experience during repeated play.

## Context
Currently the game has no user-configurable settings. Animation timings are hardcoded (e.g. 1400ms trick capture), audio plays without a mute option, and there's no way to opt into convenience features like auto-play or turn notifications. The only user preference stored today is the player name. Power users and repeat players feel the friction most.

## Out of Scope
- Visual theming (dark mode, card backs, felt color)
- Server-side preference storage (settings live in localStorage only)
- Card sort order customization
- Confirm-before-play toggle
