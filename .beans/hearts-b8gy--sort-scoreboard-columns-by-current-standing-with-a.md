---
# hearts-b8gy
title: Sort scoreboard columns by current standing, with animated reordering
status: todo
type: feature
priority: normal
created_at: 2026-03-17T07:55:28Z
updated_at: 2026-03-17T07:56:43Z
---

Sort scoreboard columns by ascending live total points (fewest = winning = leftmost), updating on every snapshot with animated column transitions

## Context
The scoreboard renders player columns in fixed, server-arrival order. There is no visual cue about who is winning or losing. In Hearts, the player with the fewest total points is winning, so column order carries meaningful rank information.

## Higher Goal
Give players an at-a-glance read of the current standings without having to mentally compare numbers across columns.

## Acceptance Criteria
- [ ] Scoreboard columns are ordered by ascending live total points (fewest = leftmost = winning)
- [ ] The ordering updates on every snapshot render (i.e. after each trick or round ends)
- [ ] Column reordering is animated so position changes are visually obvious (CSS transition or similar)
- [ ] Tied players maintain a stable relative order (no jitter)
- [ ] The column representing the local player is still visually distinguishable (existing highlighting, if any, is preserved)

## Out of Scope
- Adding a rank badge or trophy icon to the leading player
- Reordering seat positions in the play area (only the scoreboard)
- Persisting column order across page reloads
