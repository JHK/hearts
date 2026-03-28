---
# hearts-9nsa
title: Improve takeover UX during passing phase
status: completed
type: feature
priority: normal
created_at: 2026-03-28T14:30:00Z
updated_at: 2026-03-28T14:45:38Z
---

When a human takes over a bot seat during the passing phase **after the bot has already submitted passes**, the UX is confusing:

1. **No visibility into what was passed.** The bot chose cards on the player's behalf, but the player has no way to see which cards were passed until the pass-review phase begins. The `pass_sent` data exists in the snapshot but isn't surfaced in the passing-phase UI.

2. **Stale button label.** The player sees a greyed-out "Pass Left" (or equivalent direction) button. Since they can't act on it, this is misleading. It would be better to show the "Continue" button (greyed out / with a waiting message) so the player understands they're waiting for other players to finish passing — not that something is broken.

## Current behavior

- Bot submits pass → human takes over → sees disabled "Pass Left" button and their (post-pass) hand
- No indication of which cards were passed or that they're waiting on others
- Phase transitions to pass_review → then they see Continue + pass_sent/pass_received

## Desired behavior

- When `pass_submitted` is true during passing phase:
  - Show which cards were passed on their behalf (e.g. highlight or list the `pass_sent` cards)
  - Replace the "Pass Left" button with a disabled "Continue" button or a "Waiting for other players…" indicator
  - This makes the transition to pass_review feel seamless rather than jarring

## Implementation notes

- `snapshot.PassSent` is already populated via `round.PassSent(seat)` — it just needs to be surfaced in the passing-phase UI, not only during pass_review
- The button swap logic is in `render.js` lines ~582-595 — condition on `passSubmitted` to show the waiting state instead of the disabled pass button
- Verify that `PassSent` is included in the snapshot during passing phase (not just pass_review) — if not, `table_snapshot.go` needs a small update

## Tasks

- [x] Ensure `PassSent` is populated in snapshot during passing phase when pass is already submitted (already works)
- [x] Update `render.js` to show passed cards and waiting state when `passSubmitted` is true
- [x] Test: take over bot seat after bot has passed, verify new UX

## Summary of Changes

Updated render.js to improve the takeover UX when a human takes over a bot seat during the passing phase after the bot has already submitted passes:
- Show "Waiting for other players…" message instead of a stale disabled "Pass Left" button
- Highlight the cards that were passed on the player's behalf
- Pass all arguments in the recursive renderYourHand call for robustness
