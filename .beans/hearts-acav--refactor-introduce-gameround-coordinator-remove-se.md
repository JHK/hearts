---
# hearts-acav
title: 'Refactor: introduce game.Round coordinator, remove Seat/Player interfaces'
status: completed
type: task
priority: normal
created_at: 2026-03-24T15:22:10Z
updated_at: 2026-03-24T15:30:11Z
---

Introduce a game.Round step-at-a-time state machine that owns per-round state (hands, tricks, passes, scoring). Remove the Seat interface (bots get direct ChoosePlay/ChoosePass methods) and the Player interface (Round tracks state internally). Session and sim become thinner orchestrators.

## Tasks
- [x] Create game/round.go with Round coordinator
- [x] Update game/rules.go: Play struct uses seat index, TrickWinner returns seat index
- [x] Update game/seat.go: remove Seat interface, keep TurnInput/PassInput
- [x] Delete game/player.go
- [x] Update bot package: Bot interface with ChoosePlay/ChoosePass, remove callback pattern
- [x] Rewrite sim package to use Round
- [x] Rewrite session/table.go to use Round
- [x] Update all tests
- [x] Update architecture docs


## Summary of Changes

Introduced `game.Round` as a step-at-a-time state machine that owns all per-round state (hands, tricks, passes, scoring). Removed the `Seat` interface and `Player` interface from the game package. Bot strategies now implement `ChoosePlay`/`ChoosePass` directly instead of using a callback pattern. The session layer drives the Round coordinator and tracks only cumulative scoring and identity. The sim package was dramatically simplified from parallel Player/Seat arrays to simple Round method calls.
