---
# hearts-hukr
title: Introduce sentinel errors at package boundaries
status: completed
type: task
priority: normal
created_at: 2026-03-24T17:36:21Z
updated_at: 2026-03-24T17:57:41Z
---

## Context

All errors in the codebase are ad-hoc `errors.New()` or `fmt.Errorf()` strings. There are no custom error types, no sentinel errors, and no `errors.Is()`/`errors.As()` calls. Errors flow across package boundaries as opaque strings — callers never branch on *which* error occurred, only on `err != nil`.

This works today because most errors are forwarded straight to the WebSocket client as `.Error()` strings. But the implicit contract between packages (e.g. what errors can `game.Round.Play()` return?) is invisible, and any future branching logic would require fragile string matching.

## Higher Goal

Make cross-package error contracts explicit and caller code more readable. Typed/sentinel errors document what can go wrong and let callers branch on error *kind* without string comparison.

## Acceptance Criteria

- [x] `internal/game`: Export sentinel errors for the `Round` state machine — wrong phase, not your turn, card not in hand, illegal play (hearts not broken, must follow suit, etc.). Callers in `session/table.go` that check `err != nil` from `Play()`, `SubmitPass()`, etc. can use `errors.Is()` if needed.
- [x] `internal/game`: Export sentinel errors for `ParseCard()` / `ParseCards()` — invalid format, unknown rank, unknown suit.
- [x] `internal/game/bot`: Export sentinel errors for `ChoosePlay()` / `ChoosePass()` — no legal plays, not enough cards.
- [x] `internal/session`: Export sentinel errors for table state — game over, round in progress, table full, table stopping.
- [x] Existing error messages preserved (sentinel `.Error()` returns the same user-facing strings).
- [x] No `errors.Is()` / `errors.As()` calls are *required* yet — this is about defining the contract, not forcing callers to switch. Adopt `errors.Is()` where it improves readability.
- [x] All tests pass (`mise run test`).

## Out of Scope

- Changing error messages shown to WebSocket clients.
- Adding error codes or structured error responses on the wire protocol.
- Wrapping every internal error — only errors that cross package boundaries or are used for branching.


## Summary of Changes

Added exported sentinel errors at three package boundaries (`internal/game`, `internal/game/bot`, `internal/session`). Parameterized errors wrap sentinels with `fmt.Errorf("%w ...")` so `errors.Is()` works while preserving user-facing message strings. No callers were forced to switch to `errors.Is()` — this defines the contract for future use.
