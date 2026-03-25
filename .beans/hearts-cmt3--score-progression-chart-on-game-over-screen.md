---
# hearts-cmt3
title: Score progression chart on game-over screen
status: todo
type: feature
priority: normal
tags:
    - frontend
created_at: 2026-03-25T10:08:22Z
updated_at: 2026-03-25T10:08:42Z
---

Show a line chart of cumulative scores per player on the game-over overlay, using Chart.js and the existing round history data

## Context

When a game ends, the game-over overlay shows a winner announcement and a table of final scores. The per-round score history (`snapshot.round_history`) is already available on the client but isn't visualized — players can't see how the game unfolded.

## Higher Goal

Make the end-of-game moment more engaging and give players a quick visual sense of how each player's score progressed, who was leading when, and where the decisive rounds were.

## Acceptance Criteria

- [ ] Chart.js added as a dependency and served as a fingerprinted asset
- [ ] Game-over overlay includes a line chart showing cumulative score per player across rounds
- [ ] Each player's line is visually distinguishable (color or label)
- [ ] The winner(s) line is highlighted or called out
- [ ] Chart renders correctly for games of varying length (2–20+ rounds)
- [ ] Works on mobile viewports (responsive sizing)

## Out of Scope

- Interactive tooltips or hover effects on chart points beyond Chart.js defaults
- Per-trick or per-card granularity (chart is per round only)
- Exporting or sharing the chart
