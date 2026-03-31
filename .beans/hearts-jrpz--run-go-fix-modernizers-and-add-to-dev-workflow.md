---
# hearts-jrpz
title: Run go fix modernizers and add to dev workflow
status: todo
type: task
priority: normal
tags:
    - backend
created_at: 2026-03-31T15:37:17Z
updated_at: 2026-03-31T15:37:44Z
parent: hearts-u20m
---

## Context

Go 1.26 expanded `go fix` with dozens of modernization fixers that auto-apply idiomatic updates (e.g. modern loop variable semantics, stdlib `slices`/`maps` usage). Currently we run `gofmt -w` via `mise run fmt` but don't run `go fix`.

## Higher Goal

Keep the codebase idiomatic as Go evolves, with minimal manual effort — similar to how `go fmt` keeps formatting consistent.

## Acceptance Criteria

- [ ] `go fix ./...` has been run and all applicable modernizations reviewed and applied
- [ ] `mise run fmt` (or a new `mise run fix` task) includes `go fix ./...` so it runs routinely alongside `gofmt`
- [ ] All tests pass after applying fixes

## Out of Scope

- Manual refactoring beyond what `go fix` suggests
- Upgrading Go version (already on 1.26.1)

## References

- [Go 1.26 release notes — go fix](https://go.dev/doc/go1.26): expanded modernizers in go fix
