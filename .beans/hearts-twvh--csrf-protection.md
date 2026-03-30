---
# hearts-twvh
title: Origin-check WebSocket upgrades in production
status: todo
type: task
priority: normal
tags:
    - security
created_at: 2026-03-25T09:20:46Z
updated_at: 2026-03-30T09:55:50Z
---

## Context

The WebSocket upgrader in `ws.go` has `CheckOrigin: func(r *http.Request) bool { return true }`,
accepting connections from any origin. There are no POST/mutation HTTP endpoints, so the
WebSocket upgrade is the only cross-origin attack surface. A malicious page could open a
WebSocket to the production server and interact with the game API on behalf of a visiting user.

## Higher Goal

Prevent unauthorized frontends from using the game's WebSocket API when deployed.

## Acceptance Criteria

- [ ] In production, WebSocket upgrades are rejected unless the `Origin` header matches the request's `Host` header
- [ ] In dev mode (`-dev` flag), all origins are accepted (preserves hot-reload workflows)
- [ ] Rejected upgrades return HTTP 403 with a log line at `warn` level

## Out of Scope

- CORS headers for regular HTTP responses (no cross-origin HTTP API exists)
- CSRF tokens or double-submit cookies (no POST handlers exist)
- Authentication or authorization of WebSocket sessions
