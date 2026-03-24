---
# hearts-woy2
title: Ensure bots in the same game get unique names
status: completed
type: bug
priority: normal
created_at: 2026-03-20T11:06:38Z
updated_at: 2026-03-24T17:08:32Z
---

## Context

When a bot is added to a table, its name is chosen randomly from a fixed per-strategy list (e.g. `randomBotNames`, `smartBotNames`). There is no check against names already taken by other players in the game. If two bots of the same strategy type join, they can end up with the same display name, causing confusing UI and potentially breaking assumptions elsewhere.

The collision point is `handleAddBot` in `internal/table/runtime.go:518` — `strategy.BotName()` returns a random name with no awareness of existing players.

## Desired Behavior

Every player in a game (bot or human) has a unique name. If a bot's chosen name is already taken, a different name should be selected.

## Acceptance Criteria

- [x] Two bots added to the same game never share a display name
- [x] A bot's name does not collide with a human player's name either

## Out of Scope

- Enforcing uniqueness of human-chosen names (humans pick their own names; that's a separate concern)
- Persisting or reserving names across tables

## Summary of Changes

Changed `StrategyKind.BotName()` to accept a `taken map[string]bool` parameter. The caller in `handleAddBot` now collects existing player names and passes them in. `BotName` shuffles the pool and picks the first untaken name; if the entire pool is exhausted, it appends a numeric suffix (e.g. "Fritz 2"). Added three tests covering collision avoidance, suffix fallback, and human name avoidance.
