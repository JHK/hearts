---
# hearts-3w3n
title: Remove [bot] suffix from bot display names
status: completed
type: task
priority: normal
tags:
    - frontend
    - ux
created_at: 2026-03-28T15:16:37Z
updated_at: 2026-03-28T15:24:00Z
---

## Context

Bot names are displayed with a ` [bot]` suffix appended in the client-side JS (`render.js`). This adds visual noise — the design system already states that bots are placeholders for humans and the UI should not treat them as a distinct class.

## Higher Goal

Leaner, less cluttered game interface that treats all players uniformly.

## Acceptance Criteria

- [x] Bot player names render without the `[bot]` suffix in all UI locations (player labels, scoreboard, chart legend, game-over screen)
- [x] `is_bot` field is still available in the client for any future logic that needs it

## Out of Scope

- Removing `is_bot` from the protocol/server — only the display suffix is removed
- Any other bot-related UI differentiation (e.g. avatars, colors)

## Summary of Changes

Removed the `[bot]` suffix from four display locations in `render.js`: the `displayName()` helper, scoreboard headers, game-over table, and chart legend labels. The `is_bot` field remains available for claim-seat logic.
