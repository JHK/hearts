---
# hearts-1aa7
title: Move gameOverThreshold into game.Round
status: completed
type: task
priority: normal
created_at: 2026-03-24T17:49:12Z
updated_at: 2026-03-24T18:06:35Z
---

## Context

`gameOverThreshold` (100 points) is currently defined as a **duplicate constant** in two places:
- `internal/session/table.go:991`
- `internal/sim/sim.go:11`

Both `Table` and `Sim` independently accumulate cumulative scores across rounds and check whether any player has breached the threshold. `game.Round` has no awareness of cumulative scores or game-over conditions.

## Higher goal

The game-over threshold is a **domain rule** — it belongs in `game`, not in the runtime (`session`) or simulation (`sim`) layers. Centralizing it in `Round` (or a new `Game` coordinator) eliminates duplication, keeps domain logic in the domain layer, and makes it easier to change the threshold (e.g. for short games).

## Acceptance criteria

- [x] `gameOverThreshold` is defined once, in `internal/game`
- [x] `game.Round` (or a thin game-level coordinator) tracks cumulative scores and exposes a method/phase indicating game-over + winners
- [x] `session.Table` delegates game-over detection to the domain layer instead of doing its own check
- [x] `sim.Sim` delegates game-over detection to the domain layer instead of doing its own check
- [x] No duplicate winner-computation logic — single source of truth in `internal/game`
- [x] Existing tests pass; add unit tests for the new game-over logic in `internal/game`

## Out of scope

- Configurable threshold (nice-to-have, but not this ticket)
- Persisting cumulative scores across restarts

## Summary of Changes

Introduced `game.Game` type in `internal/game/game.go` as a thin game-level coordinator that tracks cumulative scores across rounds, detects game-over (threshold = 100), computes winners, and provides `NextPassDirection()`. Both `session.Table` and `sim.Sim` now delegate all game-over and winner logic to this domain type. Removed duplicate `gameOverThreshold` constants and `computeWinners`/`winners` functions from session and sim packages. Removed `cumulativePoints` from `playerState` and `roundsStarted` from `tableState`. Added comprehensive unit tests for the new `Game` type.
