---
# hearts-ftkk
title: Clean up interface boundaries between game, bot, and table layers
status: done
type: task
priority: normal
created_at: 2026-03-24T10:51:37Z
updated_at: 2026-03-24T10:51:44Z
---

Remove table-layer concerns from bot.Bot, clarify game.Player contract, separate seated identity from web transport, rename Runtime to Table, consolidate Player/Participant

## Context
The interfaces between the three main layers have accumulated concerns that don't belong:

- `bot.Bot` carries two table-layer concerns: `BotName() string` (only called by `table` to populate
  a player's name — the sim never uses it) and `Unwrap() *game.Player` (a reconnection mechanism
  that is purely a table/web concern and should not be part of a strategy interface)
- `playerState` in `table/runtime.go` mixes seated identity (Name, Seat, Participant) with
  web-transport state (Token, protocol.PlayerID), making the general concept hard to see
- `game.Participant` has no doc comment establishing its contract, so its intentional narrowness
  is invisible to readers
- `table.Runtime` is a misleading name; in Go, the idiomatic name for the main type in a package
  is the package name itself (`table.Table`)

## Higher Goal
Clean layer boundaries make it easier to reason about each layer in isolation, extend the sim
without web concerns bleeding in, and onboard contributors without needing to trace where each
concern actually lives. The sim, table, and game packages should each have an obvious, minimal
interface contract.

## Acceptance Criteria
- [x] `game.Participant` has a doc comment that explicitly states it represents pure game-mechanical
      state: hand, points, pass state, and game operations — no identity, no I/O, no naming
- [x] `bot.Bot` contains only strategy/decision methods: `ChoosePlay`, `ChoosePass`, `Kind`;
      `BotName()` and `Unwrap()` are removed from the interface
- [x] Bot names are assigned at the session layer via `StrategyKind.BotName()`, consistent with
      how human names are assigned — bots and humans use the same path
- [x] The `Unwrap()` concern is handled without polluting the `bot.Bot` interface (a separate
      `unwrappable` interface scoped to the session package)
- [x] `playerState` makes the distinction between seated identity and web-transport state visible —
      Token and id are clearly separate from Name, position, and Participant
- [x] `table.Runtime` renamed to `session.Table`; package renamed from `table` to `session`
- [x] `game.Seat` interface added (`RequestPlay`/`RequestPass` with callback); `game.TurnInput`
      and `game.PassInput` moved from `bot` to `game` as pure domain types
- [x] Bot types implement `game.Seat`; `bot.Bot` embeds `game.Seat`
- [x] Session runtime uses `game.Seat` for bot scheduling; `player.seat != nil` replaces
      `bot.Bot` type assertions for bot detection
- [x] `game.Participant` renamed to `game.Player` (the exported interface); `Player` struct made
      private (`player`) — consumers only see the interface, resolving naming confusion
- [x] Bots hold no player state: `*game.Player` embedding removed from all bot structs;
      `Unwrap()` eliminated entirely (not just moved to a separate interface)
- [x] `WrapPlayer` removed; human→bot conversion in leave just sets `seat`, no participant swap
- [x] Human reconnection just clears `seat` — no unwrapping needed
- [x] Sim separates `[4]game.Player` (state) from `[4]game.Seat` (strategy), mirroring session pattern
- [x] `bot.Bot` slimmed to `game.Seat` + `Kind()`; `ChoosePlay`/`ChoosePass` removed from interface,
      replaced by `bot.Play()`/`bot.Pass()` sync helper functions that wrap any `game.Seat`
- [x] All existing tests pass

## Out of Scope
- Changes to game rules or strategy logic
- New bot strategies
- Protocol or wire format changes
- Giving the sim named/seated players (it works with array indices and that is sufficient)
