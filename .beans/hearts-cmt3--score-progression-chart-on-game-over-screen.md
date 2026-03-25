---
# hearts-cmt3
title: Score progression chart on game-over screen
status: completed
type: feature
priority: normal
tags:
    - frontend
created_at: 2026-03-25T10:08:22Z
updated_at: 2026-03-25T13:14:05Z
---

Show a line chart of cumulative scores per player on the game-over overlay, using Chart.js and the existing round history data

## Context

When a game ends, the game-over overlay shows a winner announcement and a table of final scores. The per-round score history (`snapshot.round_history`) is already available on the client but isn't visualized — players can't see how the game unfolded.

## Higher Goal

Make the end-of-game moment more engaging and give players a quick visual sense of how each player's score progressed, who was leading when, and where the decisive rounds were.

## Acceptance Criteria

- [x] Chart.js added as a dependency and served as a fingerprinted asset
- [x] Game-over overlay includes a line chart showing cumulative score per player across rounds
- [x] Each player's line is visually distinguishable (color or label)
- [x] The winner(s) line is highlighted or called out
- [x] Chart renders correctly for games of varying length (2–20+ rounds)
- [x] Works on mobile viewports (responsive sizing)

## Out of Scope

- Interactive tooltips or hover effects on chart points beyond Chart.js defaults
- Per-trick or per-card granularity (chart is per round only)
- Exporting or sharing the chart

## Summary of Changes

Added a Chart.js line chart to the game-over overlay showing cumulative score progression per player across rounds. Replaced the winner text with medal emoji rankings (🏆🥈🥉) in the scores table. Each player row has a colored dot swatch matching their chart line for clear identification. Chart.js is vendored and served as a fingerprinted asset with defer loading.
