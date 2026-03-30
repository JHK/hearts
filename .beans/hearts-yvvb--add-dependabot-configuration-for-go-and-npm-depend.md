---
# hearts-yvvb
title: Add Dependabot configuration for Go and npm dependencies
status: todo
type: task
priority: normal
tags:
    - infrastructure
created_at: 2026-03-30T09:45:38Z
updated_at: 2026-03-30T09:46:46Z
---

Add .github/dependabot.yml with weekly grouped updates for gomod and npm ecosystems

## Context

The project has Go module dependencies (including indirect `golang.org/x/*` packages pinned to 2021 hashes) and a single Node devDependency (Tailwind CSS). There is no automation for dependency updates. Research in hearts-h9d0 evaluated Dependabot vs Renovate and chose Dependabot for its simplicity and zero-install setup.

## Higher Goal

Keep dependencies current with minimal manual effort, reducing security exposure from stale transitive dependencies.

## Acceptance Criteria

- [ ] `.github/dependabot.yml` added with `gomod` and `npm` ecosystems, weekly schedule
- [ ] Go dependencies grouped into a single PR
- [ ] npm dependencies grouped into a single PR (future-proofing if more devDeps are added)
- [ ] First Dependabot PRs appear after merge (verified manually)

## Out of Scope

- Auto-merge setup (Actions workflow or branch rulesets)
- Go toolchain version updates
- GitHub Actions dependency updates (no workflows exist yet)
