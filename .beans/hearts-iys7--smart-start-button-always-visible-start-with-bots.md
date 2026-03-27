---
# hearts-iys7
title: Smart Start button — always visible, Start with bots when seats open
status: completed
type: task
priority: normal
created_at: 2026-03-27T09:16:30Z
updated_at: 2026-03-27T15:43:30Z
parent: hearts-dfll
blocked_by:
    - hearts-gfmc
---

Make the Start button visible from the moment the table is created. When seats are still open, the label changes to "Start with bots" — clicking it fills remaining seats with hard bots and starts the game.

## Context

Currently the Start button only appears once all 4 seats are filled. This creates a dead zone where a solo player sees no clear call to action. Showing the button immediately with a smart label makes the path to playing obvious — especially for someone who just wants to jump into a game against bots.

## Higher Goal

Reduce friction to start playing. One click from table creation to game start.

## Acceptance Criteria

- [x] Start button visible in the trick center as soon as the table exists (for the table owner)
- [x] When all 4 seats are filled: label is "Start"
- [x] When seats are still open: label is "Start with bots"
- [x] "Start with bots" fills remaining empty seats with hard bots, then starts the game
- [x] If bots were already added at different strengths, those are preserved — only truly empty seats get hard bots
- [x] Button styling follows design system (hearts-8ivt) and trick center pattern

## Out of Scope

- Letting the user choose bot strength from the Start button (use the 🤖 header icon for that)
- Trick center layout changes (separate bean)
- Non-owner players seeing the Start button

## Summary of Changes

Made the Start button visible immediately when a player is seated (instead of waiting for all 4 seats to fill). When seats are open, the label shows "Start with bots" (or "Continue with bots" after a completed round). Clicking it auto-fills empty seats with hard bots, preserving any bots already added at other strengths, then starts the game. The server-side `validateStartPreconditions` no longer requires exactly 4 players — instead, `handleStart` calls `fillEmptySeatsWithBots` to fill remaining seats before initializing the round.
