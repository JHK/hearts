---
# hearts-5v74
title: Bot auto-submits pass immediately on disconnect, leaving no window to reconnect
status: completed
type: bug
priority: normal
tags:
    - backend
created_at: 2026-03-28T14:50:04Z
updated_at: 2026-03-28T14:56:57Z
---

## Context

When a player disconnects during the passing phase, `handleLeave` immediately converts them
to a bot and the bot auto-submits a random pass in the same handler call. Even a brief
network hiccup or accidental navigation means the player loses control of their pass with
zero grace period.

Related: hearts-9nsa covers UX polish for when a takeover *has* happened (showing passed
cards, better button state). This bean addresses the root cause — the bot shouldn't act
instantly.

## Current Behavior

1. Player disconnects during passing phase (hasn't submitted yet)
2. `handleLeave` converts seat to bot and immediately calls `bot.ChoosePass` + `SubmitPass`
3. Player reconnects moments later — pass is locked in, random cards were sent

## Desired Behavior

- The bot conversion still happens on disconnect (needed for game pause logic), but the bot
  should not immediately submit a pass
- The returning human should find the game paused with their pass still unsubmitted
- If the game is resumed with a bot still in the seat (no human returns), the bot submits
  at that point

## Acceptance Criteria

- [x] Disconnecting during PhasePassing does not auto-submit a pass for the departing player
- [x] Reconnecting player can still submit their own pass
- [x] When game resumes with bot still in seat, bot submits its pass normally
- [x] Existing tests pass; new test covers disconnect-reconnect during passing

## Out of Scope

- UX improvements for showing bot-passed cards (hearts-9nsa)
- Changing pass finality — once submitted, a pass cannot be undone

## Summary of Changes

Removed the immediate `ChoosePass`/`SubmitPass` call from `handleLeave` during `PhasePassing`. The bot conversion and game pause still happen on disconnect, but the bot no longer auto-submits a pass. When the game resumes (either via reconnection or manual resume), `resumeAfterPause` → `schedulePassingBots` handles bot pass submission through the normal async path.

Files changed:
- `internal/session/table_handlers.go`: Removed `PhasePassing` case from the switch in `handleLeave`
- `internal/session/table.go`: Added `Drain()` helper for test synchronization
- `internal/session/table_test.go`: Added two tests covering disconnect-reconnect and disconnect-resume-with-bot during passing
