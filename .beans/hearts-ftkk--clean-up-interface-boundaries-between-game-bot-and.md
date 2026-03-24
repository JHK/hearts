---
# hearts-ftkk
title: Clean up interface boundaries between game, bot, and table layers
status: todo
type: task
priority: normal
created_at: 2026-03-24T10:51:37Z
updated_at: 2026-03-24T10:51:44Z
---

Remove table-layer concerns from bot.Bot, clarify game.Participant contract, separate seated identity from web transport, rename Runtime to Table

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
- [ ] `game.Participant` has a doc comment that explicitly states it represents pure game-mechanical
      state: hand, points, pass state, and game operations — no identity, no I/O, no naming
- [ ] `bot.Bot` contains only strategy/decision methods: `ChoosePlay`, `ChoosePass`, `Kind`;
      `BotName()` and `Unwrap()` are removed from the interface
- [ ] Bot names are assigned at the table layer (e.g. via factory or config), consistent with
      how human names are assigned — bots and humans use the same path
- [ ] The `Unwrap()` concern is handled without polluting the `bot.Bot` interface (type assertion
      at call site or a separate `unwrappable` interface scoped to the table package)
- [ ] `playerState` (or a named interface for it) makes the distinction between seated identity
      and web-transport state visible — Token and protocol.PlayerID are clearly separate from
      Name, Seat, and Participant
- [ ] `table.Runtime` is renamed to `table.Table`
- [ ] All existing tests pass

## Out of Scope
- Changes to game rules or strategy logic
- New bot strategies
- Protocol or wire format changes
- Giving the sim named/seated players (it works with array indices and that is sufficient)
