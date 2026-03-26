---
# hearts-c65d
title: Hearts logo on lobby table cards
status: completed
type: task
priority: normal
tags:
    - frontend
created_at: 2026-03-26T11:46:24Z
updated_at: 2026-03-26T14:46:36Z
parent: hearts-14cx
---

Display the existing icon.svg on each table entry in the lobby list for visual branding

## Context
The lobby table list shows table name, status badge, and player count — but has no visual branding. Adding the hearts logo to each table entry gives the lobby a more polished, game-like feel.

## Higher Goal
Part of the lobby beautification effort (hearts-14cx) to give the game a more finished, branded appearance.

## Acceptance Criteria
- [x] The existing `icon.svg` (favicon source) is displayed on each table entry in the lobby list
- [x] Table cards use a 3D flip animation — tap to flip, back side shows Join button
- [x] The logo is sized and positioned so it doesn't crowd the table name or metadata
- [x] Looks good on mobile (responsive layout)

## Out of Scope
- Creating a new logo asset — reuses existing `icon.svg`
- Changes to the table/game page itself (lobby list only)

## Summary of Changes

Redesigned the lobby into a unified "Tables" view:
- Square card tiles with the hearts logo, 3D flip animation to join
- "+" card (matching back-face style) to create tables without naming
- Player name moved to a settings gear popover in the header
- Floating presence names drift around the tables area (casino feel)
- Removed separate Create Table, player name, and presence sections
