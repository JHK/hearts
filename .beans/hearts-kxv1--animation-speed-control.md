---
# hearts-kxv1
title: Animation Speed Control
status: todo
type: feature
priority: normal
created_at: 2026-03-25T09:42:38Z
updated_at: 2026-03-25T09:42:48Z
parent: hearts-5ceo
---

## Context
All animation timings are hardcoded: 520ms card play, 1400ms trick capture with 90ms stagger, 1200ms winner pulse, 400ms scoreboard FLIP. For experienced players this feels sluggish, especially over many rounds.

## Higher Goal
Part of the In-Game User Settings epic — giving players control over their gameplay pacing.

## Acceptance Criteria
- [ ] Player can toggle between normal and fast animation speed
- [ ] Speed setting affects all gameplay animations (card play, trick capture, winner pulse, scoreboard)
- [ ] Setting persists in localStorage across sessions
- [ ] `prefers-reduced-motion` still takes precedence — animations stay disabled regardless of speed setting
- [ ] Setting is accessible from an in-game settings UI element

## Out of Scope
- Lobby or non-gameplay animations
- Per-animation granular control (one setting controls all)
- Slider or more than two presets
