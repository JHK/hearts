---
# hearts-ax6n
title: Rotate player names vertically in scoreboard header
status: completed
type: feature
priority: normal
created_at: 2026-03-17T07:41:57Z
updated_at: 2026-03-17T08:52:32Z
---

## Context

The scoreboard takes up unnecessary horizontal space due to wide player name columns and verbose row labels. With four players (especially bots with the ` [bot]` suffix), the table triggers a horizontal scrollbar. The scoreboard already has `overflow-x: auto` as a fallback, but scrolling is friction — the score should be readable at a glance.

## Higher Goal

Keep the scoreboard readable without horizontal scrolling, even with longer or more player names. Lay groundwork for i18n by replacing text labels with symbols where possible.

## Acceptance Criteria

- [ ] Player name headers are always rotated vertically (`writing-mode: vertical-rl`)
- [ ] "Round N" row labels are replaced with just the number (`3` instead of `Round 3`)
- [ ] "Current" row label is replaced with `►`
- [ ] "Sum" row label is replaced with `Σ`
- [ ] No horizontal scrollbar appears in a standard 4-player game

## Out of Scope

- Conditional/responsive rotation based on overflow detection
- Truncating or abbreviating player names
- Full i18n implementation

## Summary of Changes

- Player name headers rotate vertically at ≥1280px (when scoreboard is side-by-side with the game table)
- Round labels shortened to just the number (e.g. `3` instead of `Round 3`)
- Row labels replaced with symbols: Current → `►`, Sum → `Σ`, Round header → empty
- Rotation only applies at wide breakpoints; stacked layout keeps horizontal names
