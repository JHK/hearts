---
# hearts-0jmb
title: Introduce game.Player as authoritative owner of player game state
status: in_progress
type: task
priority: normal
created_at: 2026-03-20T10:44:39Z
updated_at: 2026-03-20T11:30:00Z
---

## Context

Player game state is currently scattered across multiple structs in `internal/table/runtime.go`:
- `playerState.Hand []game.Card` — managed in 3 places (`initializeRound`, `applyPassSubmissions`, `applyValidatedPlay`)
- `roundState.PassSubmissions/Sent/Received` — maps keyed by `PlayerID`
- `roundState.RoundPoints` and `tableState.totals` — more maps keyed by `PlayerID`

The `sim` package duplicates all of this independently with `[4][]game.Card` arrays and inline point tracking. The `internal/game/` layer has no `Player` concept — only pure functions operating on raw slices passed in by the caller.

## Higher Goal

Centralizing player game state in a `game.Player` struct makes `internal/game/` the single source of truth for what a player *is* in Hearts. The `sim` and `runtime` layers become thinner orchestrators — they delegate state management to `game.Player` and focus on what they own: bot scheduling + event streaming (`sim`/`runtime`), and server-side input validation (`runtime`). The UI snapshot fields that mirror player state (hand, points, pass info) become straightforward reads from `game.Player`.

## Acceptance Criteria
- [x] A `game.Player` struct exists in `internal/game/` with fields for hand, round points, cumulative points, and pass state (submitted, sent, received)
- [x] `game.Player` exposes methods for: receiving dealt cards, submitting a pass, receiving passed cards, playing a card (removes from hand), accumulating trick points
- [x] `runtime.playerState` no longer manages hand or points directly — delegates to an embedded or referenced `game.Player`
- [x] `roundState` pass maps and round-points maps are replaced by per-player state on `game.Player`
- [x] `sim` uses `game.Player` instances instead of raw `[4][]game.Card` arrays and ad-hoc point tracking
- [x] All existing tests pass; unit tests added for `game.Player` methods
- [x] Runtime continues to validate all player inputs (card legality, pass card count, etc.) before delegating to `game.Player`

## Out of Scope
- Changing the command/event protocol or WebSocket API
- Moving bot strategy logic into `game.Player`
- UI changes
- Persistence beyond what already exists
