---
# hearts-x0fa
title: Extract shared helpers into helpers.go
status: completed
type: task
priority: low
created_at: 2026-03-26T09:55:38Z
updated_at: 2026-03-26T11:58:24Z
parent: hearts-aazx
---

## Context

server.go contains several utility functions (writeJSON, truncateUTF8, mustReadAsset, mustRenderTemplate, contentETag, serveHTMLWithETag) used across multiple handlers. After the other extractions, some of these may have moved with their primary consumer, leaving fewer behind.

## Higher Goal

Part of the server.go decomposition (hearts-aazx). Reduce server.go to pure wiring by extracting each concern into its own file.

## Acceptance Criteria

- [x] Re-evaluated what helpers remain in server.go after prior extractions
- [ ] If 3+ helpers remain, extracted to `helpers.go`
- [x] If fewer remain, decided to keep them in place (and documented why in commit message)
- [ ] All existing tests pass without modification

## Guidance

- This task should run **last** — its scope depends entirely on what the other extractions left behind.
- Some helpers (e.g. `writeJSON`) may be generic enough for a shared utility sub-package if the codebase grows, but don't over-extract for a handful of functions.

## Out of Scope

- Changing helper behavior
- Creating a general-purpose utility package

## Summary of Changes

No extraction needed. After prior decompositions (routes_pages.go, routes_api.go, ws.go), only one helper (`writeJSON`) remains in server.go. Since fewer than 3 helpers remain, keeping it in place per the guidance in this ticket.
