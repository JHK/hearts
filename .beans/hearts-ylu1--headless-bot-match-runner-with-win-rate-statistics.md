---
# hearts-ylu1
title: Headless bot match runner with win-rate statistics
status: todo
type: feature
priority: normal
created_at: 2026-03-19T07:16:47Z
updated_at: 2026-03-19T07:16:54Z
---

Standalone binary to run N games between configurable bot strategies and report win counts per slot

## Context
We have three bot strategies (smart, random, first-legal) but no way to measure how they perform against each other. Evaluating strategy quality currently requires manual observation through the live UI.

## Higher Goal
Enable empirical comparison of bot strategies so that improvements to smart bots can be validated quantitatively rather than by intuition.

## Acceptance Criteria
- [ ] A `Simulation` (or similar) type accepts exactly 4 `bot.Strategy` instances and an iteration count
- [ ] Running a simulation plays that many complete games sequentially, each starting fresh with a shuffled deck
- [ ] A game is considered complete when any player's cumulative score reaches or exceeds 100 points
- [ ] The return value is a win-count per player/strategy slot (lowest score when the threshold is crossed wins)
- [ ] Shoot-the-moon is handled correctly (existing game logic applies)
- [ ] No HTTP, WebSocket, or table runtime involved — pure in-process function calls
- [ ] A standalone `cmd/sim/main.go` binary demonstrates the runner with hardcoded strategies and a configurable iteration count

## Out of Scope
- Parallel/concurrent game execution
- Persistence of results
- Strategy configuration via flags or config files
- Any UI or reporting beyond plain text stdout output
