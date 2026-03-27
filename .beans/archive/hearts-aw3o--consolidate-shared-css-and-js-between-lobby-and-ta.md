---
# hearts-aw3o
title: Consolidate shared CSS and JS between lobby and table
status: completed
type: task
created_at: 2026-03-27T14:23:42Z
updated_at: 2026-03-27T14:23:42Z
parent: hearts-dfll
---

Extract shared overlay, settings, and fingerprint support for cross-directory JS imports.

## Summary of Changes

- CSS: Extracted .overlay and .overlay-panel base classes from duplicate game-over/game-paused overlay styles
- JS: Created js/shared/settings.js with localStorage keys, ensureToken(), and initSettingsPopover()
- JS: Converted lobby from IIFE to ES module to enable shared imports
- Go: Extended fingerprint.go import rewriting to handle ../ relative paths with correct path resolution
- Tests: Added cross-directory fingerprinted import coverage
