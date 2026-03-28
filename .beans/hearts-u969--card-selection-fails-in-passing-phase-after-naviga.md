---
# hearts-u969
title: Card selection fails in passing phase after navigating away and back
status: completed
type: bug
priority: normal
tags:
    - frontend
created_at: 2026-03-28T14:18:57Z
updated_at: 2026-03-28T14:34:15Z
---

Clicking cards to select for passing stops working after leaving the table page and returning. Desktop/mouse. Likely related to stale render state or event handler rebinding on reconnect.

## Context

Affects the passing phase on desktop. Triggered by navigating away from the table page and returning (e.g., going to the lobby and back). Card clicks stop registering for pass selection.

## Current Behavior

After leaving a table page and navigating back, clicking cards during the passing phase does not toggle their selection. The cards appear in the hand but are unresponsive to clicks.

## Desired Behavior

Card selection should work reliably in the passing phase regardless of navigation history. Returning to a table mid-pass should restore full interactivity.

## Acceptance Criteria

- [x] Cards are clickable for pass selection after navigating away and back to the table
- [x] Investigate whether bfcache restoration or `handRenderKey` staleness causes the DOM/event handler mismatch
- [ ] Fix verified in Chrome and Firefox on desktop (manual testing needed)

## Out of Scope

- Playing phase card interaction
- Mobile/touch-specific issues
- WebSocket reconnection reliability

## Summary of Changes

**Root cause:** After navigating away from the table page and back, `state.lastHandRenderKey` retained its old value. When the server sent the same game state on reconnect, `renderYourHand()` matched the stale key and skipped re-rendering — so no click handlers were attached to the card buttons.

**Fix (in `main.js`):**
1. Clear `lastHandRenderKey` and `lastTrickSignature` on WebSocket open, forcing a full re-render when state arrives after reconnection.
2. Add a `pageshow` event listener to handle bfcache restoration — when the browser restores the page from bfcache with a dead WebSocket, reset render keys and reconnect.
