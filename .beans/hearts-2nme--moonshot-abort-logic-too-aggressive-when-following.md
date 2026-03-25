---
# hearts-2nme
title: Moonshot abort logic too aggressive when following
status: completed
type: bug
priority: normal
created_at: 2026-03-25T17:20:17Z
updated_at: 2026-03-25T18:07:34Z
parent: hearts-8j8z
blocked_by:
    - hearts-gavv
---

Bot kills moonshot when another player leads, even when it holds all penalty points and strong cards

## Context
The smart bot's moonshot tracking in `internal/game/bot/smart.go` aborts the moonshot whenever another player leads a trick (lines 79-86). The dynamic re-activation check (lines 99-108) only runs when the bot is leading (`len(input.Trick) == 0`), so a bot following in a trick can never re-activate moonshot mid-trick.

## Current Behavior
When another player leads a trick, `moonShotAborted` is set to `true` unconditionally (line 83-84). The re-activation path requires `len(input.Trick) == 0`, so it never fires when the bot is following. This means a bot with 13 round points (e.g. from QS) and AH+KH in hand has its moonshot killed just because someone else led — even though it's in a strong position to collect all remaining penalty cards.

Observed in game: Edsger had 13 round points, held AC/JD/3H/KH/AH with 5 tricks remaining, and `moonShotActive` was `false` because Margaret led 6H.

## Desired Behavior
The bot should maintain or re-activate moonshot pursuit when it has captured all penalty points so far and still holds cards capable of winning remaining penalty-card tricks. The abort-on-follow heuristic should be refined — either:
1. Run the dynamic re-activation check when following too (not just when leading), or
2. Factor in current penalty-point ownership before aborting, or
3. Both

## Acceptance Criteria
- [x] A bot that has collected all penalty points so far and holds safe high cards for remaining tricks keeps moonshot active when following
- [x] The fix does not cause false-positive moonshot pursuit (bot chasing moon when it's clearly lost penalty cards to others)
- [x] A simulation run (`cmd/sim`, 50k+ iterations) shows no decreased win rates compared to baseline
- [x] Existing `smart_test.go` moonshot tests pass; new test covers the follow-abort scenario

## Out of Scope
- Reworking the initial `evaluateMoonShot` hand-evaluation heuristic
- Changing moonshot pass strategy

## Summary of Changes

Added `MyRoundPoints` field to `game.TurnInput` so bots know how many penalty points they have captured.

Refined the hard bot's moonshot abort logic in two ways:
1. When following (someone else leads), only abort if `MyRoundPoints < totalPenaltyPlayed` — i.e., penalty points have leaked to opponents. Previously aborted unconditionally.
2. The dynamic re-activation check (safe high cards >= remaining tricks) now also runs when following, not just when leading, but is gated on penalty ownership to prevent false-positive pursuit.

Added `penaltyPointsInCards` helper. Added test `TestHardKeepsMoonShotWhenFollowingWithAllPenalties`. 50k sim confirms no win-rate regression.
