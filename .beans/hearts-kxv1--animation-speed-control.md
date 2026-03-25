---
# hearts-kxv1
title: Animation Speed Control
status: completed
type: feature
priority: normal
created_at: 2026-03-25T09:42:38Z
updated_at: 2026-03-25T09:59:22Z
parent: hearts-5ceo
---

## Context
All animation timings are hardcoded: 520ms card play, 1400ms trick capture with 90ms stagger, 1200ms winner pulse, 400ms scoreboard FLIP. For experienced players this feels sluggish, especially over many rounds.

## Higher Goal
Part of the In-Game User Settings epic — giving players control over their gameplay pacing.

## Acceptance Criteria
- [x] Player can toggle between normal and fast animation speed
- [x] Speed setting affects all gameplay animations (card play, trick capture, winner pulse, scoreboard)
- [x] Setting persists in localStorage across sessions
- [x] `prefers-reduced-motion` still takes precedence — animations stay disabled regardless of speed setting
- [x] Setting is accessible from an in-game settings UI element

## Out of Scope
- Lobby or non-gameplay animations
- Per-animation granular control (one setting controls all)
- Slider or more than two presets


## Summary of Changes

Added animation speed control with a gear icon settings panel in the table header. Players can toggle between normal and fast (2x) animation speeds via a switch. Uses CSS custom properties (`--anim-card-in`, `--anim-trick-capture`, `--anim-winner-pulse`, `--anim-scoreboard-flip`) with a `[data-speed="fast"]` attribute on body to halve all durations. JS-driven timings in main.js and render.js also read the speed setting dynamically. Setting persists in localStorage under `hearts.animation.speed`. The existing `prefers-reduced-motion` media query and JS checks remain untouched and take precedence.

Files changed:
- `internal/webui/assets/table.html` — settings gear button + dropdown panel
- `internal/webui/assets/styles.input.css` — CSS custom properties for animation durations, fast mode overrides, settings UI styles
- `internal/webui/assets/js/table/dom.js` — new DOM element references
- `internal/webui/assets/js/table/main.js` — speed setting init, localStorage persistence, toggle handler
- `internal/webui/assets/js/table/render.js` — dynamic trick capture and scoreboard FLIP timings
