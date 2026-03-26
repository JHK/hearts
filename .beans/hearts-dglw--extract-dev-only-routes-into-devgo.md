---
# hearts-dglw
title: Extract dev-only routes into dev.go
status: completed
type: task
priority: low
created_at: 2026-03-26T09:55:33Z
updated_at: 2026-03-26T12:08:48Z
parent: hearts-aazx
---

## Context

Dev-only routes (/dev.js, /api/debug/bots, registerDevAssetHandlers) are conditionally registered when the server runs in dev mode. They're a cohesive concern that's only relevant during development.

## Higher Goal

Part of the server.go decomposition (hearts-aazx). Reduce server.go to pure wiring by extracting each concern into its own file.

## Acceptance Criteria

- [x] Dev-only routes and helpers live in `dev.go`
- [x] `server.go` calls a registration function from the new file, gated on dev mode
- [x] All existing tests pass without modification
- [x] Re-evaluated current state of server.go before extracting

## Guidance

- Currently ~50 lines. Straightforward extraction.
- A build tag (`//go:build dev`) is tempting but would change the build model — out of scope here. Just a separate file is fine.

## Out of Scope

- Build tag gating
- Changing dev route behavior

## Summary of Changes

Extracted three dev-only functions into `internal/webui/dev.go`:
- `registerDevRoutes` — serves `/dev.js` script (moved from routes_pages.go)
- `registerDevAssetHandlers` — serves CSS/JS without fingerprinting for hot-reload (moved from routes_pages.go)
- `registerDevAPIRoutes` — serves `/api/debug/bots` endpoint (extracted from routes_api.go)

Call sites in routes_pages.go and routes_api.go now delegate to dev.go, still gated on `cfg.Dev`.
