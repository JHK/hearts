---
# hearts-rm3i
title: Evaluate switching from Dependabot to Renovate for dependency updates
status: completed
type: task
priority: normal
tags:
    - infrastructure
created_at: 2026-03-31T14:30:43Z
updated_at: 2026-03-31T14:35:03Z
---

## Context

We chose Dependabot in `hearts-h9d0` for its simplicity. After the first round of automated updates, we discovered a gap: Dependabot's `gomod` ecosystem only updates module dependencies in the `require` block — it does **not** bump the `go` or `toolchain` directives in `go.mod`. The Go version is currently pinned at 1.24.0 while 1.24.3 is available, with no automated path to update it.

Renovate handles this out of the box: it updates the `toolchain` directive by default, and the `go` directive with an opt-in `rangeStrategy: "bump"` rule.

## Higher Goal

Keep all dependencies — including the Go toolchain — automatically updated without manual tracking.

## Acceptance Criteria

- [x] Renovate is configured and tested on the repo (can coexist with Dependabot temporarily)
- [x] Confirmed that Renovate creates PRs for Go toolchain version bumps
- [x] Confirmed that Renovate covers the same scope Dependabot currently handles (Go modules, npm)
- [x] Dependabot config removed after Renovate is validated
- [x] Decision and rationale documented on this ticket

## Out of Scope

- Auto-merging PRs (can be explored separately)
- Docker / container image updates
- GitHub Actions dependency updates

## Decision & Rationale

Switching from Dependabot to Renovate because:

1. **Go toolchain gap**: Dependabot's gomod ecosystem only updates require block dependencies. It cannot bump the go or toolchain directives in go.mod. Renovate handles both out of the box — toolchain by default, and go with rangeStrategy bump.
2. **Same coverage**: Renovate covers Go modules and npm with the same grouping and weekly schedule.
3. **Better extensibility**: If we later want auto-merge, Docker, or GitHub Actions updates, Renovate supports all of these without a separate tool.

### Configuration

- renovate.json at repo root with config:recommended base
- Weekly schedule (Monday mornings) matching previous Dependabot cadence
- chore: commit message prefix preserved
- Go deps grouped as go-dependencies, npm deps grouped as npm-dependencies
- Go toolchain directive bumped via rangeStrategy bump on golang dep type

## Summary of Changes

- Added renovate.json with Renovate configuration
- Removed .github/dependabot.yml and empty .github/ directory
