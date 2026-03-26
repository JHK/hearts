---
# hearts-aazx
title: Decompose NewHandler into per-concern route registration files
status: completed
type: epic
priority: low
created_at: 2026-03-26T09:10:56Z
updated_at: 2026-03-26T12:22:25Z
parent: hearts-p6hh
---

## Vision

`server.go` (732 lines) is a monolith mixing routing, page handlers, API handlers, WebSocket protocol, presence tracking, dev tools, and helpers. The goal is to extract each concern into its own file (or sub-package where warranted), leaving `server.go` as pure wiring: create router, apply middleware, call per-concern registration functions.

## Context

After the Chi router migration (hearts-1sxq), the routing code is well-structured with route groups, but everything still lives in one file. This makes navigation hard and creates merge conflicts when multiple changes touch different concerns.

## Guidance for Each Child Task

- **Re-evaluate before extracting.** Earlier tasks will change the code. Each task must read the current state of `server.go` and the package before deciding what to move and where.
- **Consider sub-package extraction.** If a concern has a narrow interface (few calls/interactions with the rest of the package) but substantial implementation beneath it, extract it into a sub-package under `internal/webui/` rather than just a separate file. Example: the connection/presence tracker (see hearts-eefe) is a good candidate — small API surface, significant internal logic.
- **Don't over-extract.** If something is only a few lines or tightly coupled to the handler wiring, leave it in `server.go`.

## Out of Scope

- Changing behavior or public API — this is purely structural
- Refactoring the WebSocket protocol or message types
- Extracting `fingerprint.go` or `lobby_hub.go` (already well-separated)

## Summary of Changes

All 7 subtasks completed. server.go reduced from 732 to 123 lines of pure wiring (embed, Config, Run, NewHandler, writeJSON). Concerns extracted to: ws.go (WebSocket), routes_pages.go (page handlers), routes_api.go (API handlers), routes_dev.go (dev-only routes), and tracker/ sub-package (presence + connection tracking). Cross-concern wsMessage dependency also cleaned up.
