---
# hearts-qtpn
title: Pass button should show direction instead of generic label
status: todo
type: feature
priority: normal
tags:
    - frontend
    - ux
created_at: 2026-03-26T13:40:39Z
updated_at: 2026-03-26T13:40:46Z
---

## Context

During the passing phase, the submit button reads **"Pass 3 Cards"** as static text. The pass direction (left, right, across) is only shown in the turn indicator above the board. Players have to look in two places — the turn indicator for *which way* and the button for *the action*.

## Higher Goal

Reduce cognitive load during the passing phase by putting the most important information — the direction — right on the action button itself.

## Acceptance Criteria

- [ ] Submit button text includes the pass direction (e.g. "Pass Left", "Pass Right", "Pass Across")
- [ ] Turn indicator still shows direction as before (no regression)
- [ ] Button text updates correctly across rounds when direction rotates

## Out of Scope

- Redesigning the overall pass panel layout
- Translating direction labels (i18n)
