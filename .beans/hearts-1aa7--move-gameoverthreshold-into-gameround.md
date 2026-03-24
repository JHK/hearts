---
# hearts-1aa7
title: Move gameOverThreshold into game.Round
status: todo
type: task
created_at: 2026-03-24T17:49:12Z
updated_at: 2026-03-24T17:49:12Z
---

## Context

`gameOverThreshold` (100 points) is currently defined as a **duplicate constant** in two places:
- `internal/session/table.go:991`
- `internal/sim/sim.go:11`

Both `Table` and `Sim` independently accumulate cumulative scores across rounds and check whether any player has breached the threshold. `game.Round` has no awareness of cumulative scores or game-over conditions.

## Higher goal

The game-over threshold is a **domain rule** — it belongs in `game`, not in the runtime (`session`) or simulation (`sim`) layers. Centralizing it in `Round` (or a new `Game` coordinator) eliminates duplication, keeps domain logic in the domain layer, and makes it easier to change the threshold (e.g. for short games).

## Acceptance criteria

- [ ] `gameOverThreshold` is defined once, in `internal/game`
- [ ] `game.Round` (or a thin game-level coordinator) tracks cumulative scores and exposes a method/phase indicating game-over + winners
- [ ] `session.Table` delegates game-over detection to the domain layer instead of doing its own check
- [ ] `sim.Sim` delegates game-over detection to the domain layer instead of doing its own check
- [ ] No duplicate winner-computation logic — single source of truth in `internal/game`
- [ ] Existing tests pass; add unit tests for the new game-over logic in `internal/game`

## Out of scope

- Configurable threshold (nice-to-have, but not this ticket)
- Persisting cumulative scores across restarts
