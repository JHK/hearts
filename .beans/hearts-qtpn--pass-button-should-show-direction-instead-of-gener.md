---
# hearts-qtpn
title: Pass button should show direction instead of generic label
status: completed
type: feature
priority: normal
tags:
    - frontend
    - ux
created_at: 2026-03-26T13:40:39Z
updated_at: 2026-03-27T09:20:34Z
---

## Context

During the passing phase, the submit button reads **"Pass 3 Cards"** as static text. The pass direction (left, right, across) is only shown in the turn indicator above the board. Players have to look in two places — the turn indicator for *which way* and the button for *the action*.

## Higher Goal

Reduce cognitive load during the passing phase by putting the most important information — the direction — right on the action button itself.

## Acceptance Criteria

- [x] Submit button text includes the pass direction (e.g. "Pass Left", "Pass Right", "Pass Across")
- [x] Turn indicator still shows direction as before (no regression)
- [x] Button text updates correctly across rounds when direction rotates

## Out of Scope

- Redesigning the overall pass panel layout
- Translating direction labels (i18n)

## Summary of Changes

Updated `renderPassPanel` in `render.js` to set the submit button text dynamically based on `snapshot.pass_direction`. The button now reads "Pass Left", "Pass Right", or "Pass Across" instead of the static "Pass 3 Cards". The turn indicator remains unchanged. The button text is set on every render call, so it updates correctly as the direction rotates across rounds.
