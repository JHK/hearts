---
# hearts-gisl
title: Consolidate bot integration under game layer with a Participant interface
status: completed
type: task
priority: normal
created_at: 2026-03-20T13:37:00Z
updated_at: 2026-03-20T14:09:38Z
---

Move bots to game/bot, introduce game.Participant interface, replace IsBot flag and bots map with type assertions

## Context

After hearts-0jmb, `game.Player` is a concrete struct owning per-player game state. Bots live in a separate package (`internal/player/bot`) and are tracked through a parallel, bot-specific data structure in the runtime:

- `playerState.IsBot bool` — a flag on the player struct
- `tableState.bots map[game.PlayerID]bot.Strategy` — a second map alongside `playersByID`

Every bot-aware code path must check both: look up the strategy in the map and branch on `IsBot`. The `bot.Strategy` interface has no relationship to `game.Player` at the type level, so there is nothing stopping them from drifting out of sync.

Moving bots under `game/bot` and introducing a `game.Participant` interface would make bots and human players the same kind of thing at the type level, eliminating the dual-tracking and replacing `IsBot` flag checks with type assertions.

## Higher Goal

Completing the ownership picture started in hearts-0jmb: `internal/game` becomes the single source of truth for what a player *is*, whether human or bot. The runtime shrinks further — it schedules and dispatches, but does not maintain a parallel map of strategies.

## Acceptance Criteria
- [ ] `game.Participant` interface exists, implemented by both human players and bots; covers at minimum the methods needed by the runtime and sim to interact with a seated player
- [ ] Bot implementations (random, dumb, first-legal, smart) move to `internal/game/bot`, importing `game` but not vice-versa
- [ ] `tableState.bots map[game.PlayerID]bot.Strategy` is removed; the strategy is reachable via type assertion on the participant
- [ ] `playerState.IsBot bool` is removed; bot detection is a type check (`game.Bot` or similar)
- [ ] `runtime.playerState` holds a `game.Participant` instead of embedding `*game.Player` directly
- [ ] `sim` uses `game.Participant` (or `game.Bot`) instances directly — no separate strategy slice
- [ ] All existing tests pass; no change to the command/event protocol or WebSocket API

## Out of Scope
- Changing bot strategy logic or adding new strategies
- UI changes or changes to the snapshot shape visible to clients
- Persistence or bot configuration beyond what already exists
- Moving `bot.TurnInput` / `bot.PassInput` into `game` (may follow naturally but not required)
