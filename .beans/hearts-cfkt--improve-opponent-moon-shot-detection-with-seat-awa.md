---
# hearts-cfkt
title: Improve opponent moon-shot detection with seat-aware TurnInput
status: completed
type: task
priority: normal
created_at: 2026-03-26T06:48:09Z
updated_at: 2026-03-26T07:45:54Z
parent: hearts-8j8z
---

Use the new Play-based TurnInput.Trick and TurnInput.PlayedCards (with seat info) to improve opponent moon-shot detection in the hard bot. Add earlier detection via trick-winner analysis, block-aware follow, and block-aware discard. Update strategies.md accordingly.

## Summary of Changes

- **Enhanced `detectMoonShooter`**: Changed signature from `[]Card` to `[]Play` to use seat info. Added early detection path (3+ tricks, 3+ penalty points, consistent trick-winner pattern) alongside existing strong signal (4+ tricks, 14+ points). Both paths confirm against `roundPoints`.
- **Added `trickWinnerSeat` helper**: Determines which seat won a completed trick from `[]Play` data.
- **Added `hardBlockMoonFollow`**: When a shooter is detected and winning a penalty trick, overtakes with lowest over-card to steal penalty points.
- **Added `hardBlockMoonDiscard`**: When void and a shooter is winning the current trick, holds penalty cards instead of dumping them.
- **Wired blocking into `ChoosePlay`**: Block-aware follow and discard branches added alongside existing block-aware lead.
- **Updated `strategies.md`**: Documented TurnInput seat info, detection algorithm, block follow/discard, trickWinnerSeat helper. Removed opponent moon-shot detection from 'Not Yet Implemented'.

## Tuning Results

Benchmarked 13 configurations at 50k–100k games. Score-aware discard-only blocking (margin=0) was the clear winner: maintains Hard's win rate (37.1% vs 37.2% baseline) while reducing Random's moon shots by ~5%. Follow-blocking was removed — overtaking penalty tricks costs ~1.5pp in win rate for limited benefit.
