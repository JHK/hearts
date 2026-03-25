---
# hearts-l7h4
title: Split table.go into focused files
status: completed
type: task
priority: normal
created_at: 2026-03-25T12:22:22Z
updated_at: 2026-03-25T13:17:53Z
---

table.go is 1230 lines and mixes types, public API, actor loop, handlers, snapshot building, and utilities.

## Proposed Split

**table_snapshot.go** (~120 lines) — Pure read-only state serialization:
- buildSnapshot, buildBotHands, roundPhaseString
- copyPoints, copyRoundHistory

**table_handlers.go** (~760 lines) — All behavioral logic inside the actor goroutine:
- All handle* methods (join, leave, start, play, pass, ready, bot turn/pass, resume)
- Validation helpers (validateStartPreconditions, validateRoundCommandPreconditions, botForPhase)
- Bot scheduling (scheduleBotTurn, schedulePassingBots, scheduleBotPass)
- Pause/resume (publishGameResumed, resumeAfterPause)
- Round lifecycle (completeRound, buildTotals, maybeEndGame, initializeRound)
- Event publishing helpers (publishRoundStart, publishTurn, publishPlayPhaseStart, startPassReview)

**table.go** (~350 lines) — Skeleton:
- Types (Table, tableState, playerState, command types, snapshot types)
- Constructor, Close, ID
- Public API methods (Join, Start, Play, etc.) — thin wrappers around submit
- Subscription management (Subscribe)
- emit, publishPublic, publishPrivate
- Actor loop (run) with dispatch switch

## Out of Scope

Further splitting handlers by concern (play vs pass, bots vs humans) — these cross-reference freely within the actor goroutine and splitting would create artificial boundaries.

## Summary of Changes

Split `table.go` (1230 lines) into three focused files:
- **table.go** (422 lines): Types, constructor, public API, subscription management, emit helpers, actor loop with dispatch switch
- **table_handlers.go** (671 lines): All handler methods, validation helpers, bot scheduling, pause/resume, round lifecycle, event publishing
- **table_snapshot.go** (153 lines): buildSnapshot, buildBotHands, roundPhaseString, copyPoints, copyRoundHistory

Pure refactor — no behavioral changes. All existing tests pass.
