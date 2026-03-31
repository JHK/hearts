---
# hearts-rm3i
title: Evaluate switching from Dependabot to Renovate for dependency updates
status: todo
type: task
priority: normal
tags:
    - infrastructure
created_at: 2026-03-31T14:30:43Z
updated_at: 2026-03-31T14:32:38Z
---

## Context

We chose Dependabot in `hearts-h9d0` for its simplicity. After the first round of automated updates, we discovered a gap: Dependabot's `gomod` ecosystem only updates module dependencies in the `require` block — it does **not** bump the `go` or `toolchain` directives in `go.mod`. The Go version is currently pinned at 1.24.0 while 1.24.3 is available, with no automated path to update it.

Renovate handles this out of the box: it updates the `toolchain` directive by default, and the `go` directive with an opt-in `rangeStrategy: "bump"` rule.

## Higher Goal

Keep all dependencies — including the Go toolchain — automatically updated without manual tracking.

## Acceptance Criteria

- [ ] Renovate is configured and tested on the repo (can coexist with Dependabot temporarily)
- [ ] Confirmed that Renovate creates PRs for Go toolchain version bumps
- [ ] Confirmed that Renovate covers the same scope Dependabot currently handles (Go modules, npm)
- [ ] Dependabot config removed after Renovate is validated
- [ ] Decision and rationale documented on this ticket

## Out of Scope

- Auto-merging PRs (can be explored separately)
- Docker / container image updates
- GitHub Actions dependency updates
