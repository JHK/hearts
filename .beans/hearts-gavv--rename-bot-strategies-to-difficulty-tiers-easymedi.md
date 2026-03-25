---
# hearts-gavv
title: Rename bot strategies to difficulty tiers (easy/medium/hard)
status: completed
type: task
priority: normal
tags:
    - backend
    - ux
created_at: 2026-03-25T17:25:54Z
updated_at: 2026-03-25T17:37:04Z
parent: hearts-8j8z
---

Rename dumb→easy, smart→medium, clone smart as hard — establishes difficulty tiers for future bot improvements

## Context
The current bot strategies are named "dumb" and "smart", which are internal/developer labels rather than user-friendly difficulty levels. As the epic improves bot play, we want a clear tier system where "hard" is the target for improvements while "medium" preserves current smart bot behavior as a baseline.

## Higher Goal
Establish a clean difficulty tier system (easy/medium/hard) so that future bot improvements under hearts-8j8z land on the "hard" tier without regressing the existing smart bot behavior, which becomes the "medium" baseline.

## Acceptance Criteria
- [x] `StrategyDumb` renamed to `StrategyEasy` across code, tests, and UI (struct, file, constants)
- [x] `StrategySmart` renamed to `StrategyMedium` across code, tests, and UI
- [x] New `StrategyHard` added as a clone of medium (same logic, separate struct/file)
- [x] Hard is the default bot strategy in the UI (replacing smart)
- [x] Sim runner works with all three tiers; baseline win rates unchanged for easy and medium
- [x] Documentation updated (CLAUDE.md bot references, architecture.md if applicable)

## Out of Scope
- Behavioral changes to any tier (hard gets improvements in separate tickets)
- Removing the random or first-legal strategies

## Summary of Changes

Renamed bot strategies from developer labels to difficulty tiers:
- `StrategyDumb` → `StrategyEasy` (file: `easy.go`)
- `StrategySmart` → `StrategyMedium` (file: `medium.go`)
- New `StrategyHard` added as identical clone of Medium (file: `hard.go`), ready for future improvements
- Hard is now the default in UI dropdowns and rematch bot fill
- Sim runner updated to use hard/medium/easy/random
- `table_snapshot.go` uses a `moonShotter` interface instead of concrete type assertion, supporting both Medium and Hard
- No behavioral changes to any strategy
