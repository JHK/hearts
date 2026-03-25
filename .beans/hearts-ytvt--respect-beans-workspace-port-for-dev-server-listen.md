---
# hearts-ytvt
title: Respect BEANS_WORKSPACE_PORT for dev server listen address
status: completed
type: task
priority: normal
created_at: 2026-03-25T07:11:22Z
updated_at: 2026-03-25T07:21:28Z
---

## Context
Beans allocates a dynamic port per workspace and passes it to the run command via the `BEANS_WORKSPACE_PORT` environment variable. The "Open" button in the beans UI opens `http://localhost:<allocated-port>/`. The Hearts server ignores this and always binds to `127.0.0.1:8080`, so the button opens the wrong URL.

## Higher Goal
Seamless dev workflow when using beans workspaces — clicking "Open" should just work.

## Acceptance Criteria
- [x] Server reads `BEANS_WORKSPACE_PORT` and uses it as the default listen port when set
- [x] Explicit `-addr` flag still takes precedence
- [x] Default remains `127.0.0.1:8080` when the env var is absent

## Out of Scope
- Supporting multiple concurrent workspaces on different ports
- Changing beans itself to support a fixed-port config

## Summary of Changes

Changed `cmd/hearts/app/run.go` to use `BEANS_WORKSPACE_PORT` as the default listen port when set, falling back to `127.0.0.1:8080`. The env var is read before flag parsing so `-addr` still takes precedence.
