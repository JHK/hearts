---
# hearts-h9d0
title: Research dependency update automation for Go and Node dependencies
status: draft
type: task
priority: normal
created_at: 2026-03-18T06:55:54Z
updated_at: 2026-03-18T06:56:03Z
---

## Context

The project has Go module dependencies and a single Node devDependency (Tailwind CSS) that
are updated manually. The indirect `golang.org/x/*` packages in go.mod are pinned to 2021
commit hashes. There is no `.github` directory and no automation in place.

The repo is currently **private**, which may affect tool availability, pricing, or feature
access. This needs to be confirmed before any implementation work begins.

## Open Questions (resolve before moving to todo)

- Does the repo's private status restrict access to Dependabot or Renovate's free tier?
- Is there a plan/timeline to open-source the repo, and should this work wait for that?

## Options to Evaluate

- **Dependabot** — built into GitHub, minimal config (`.github/dependabot.yml`). Free for
  public repos; available on private repos but subject to GitHub plan limits.
- **Renovate** — more configurable (grouping, auto-merge rules, scheduling). Free tier
  available; self-hosted option exists if needed for private repos.

## Acceptance Criteria

- [ ] Both options are evaluated against the repo's current GitHub plan and access level
- [ ] Open questions above are answered and documented
- [ ] A follow-up task ticket is created to implement the chosen approach

## Out of Scope

- Auto-merging PRs
- Updating the Go toolchain version
- Docker / ko base image updates
- GitHub Actions workflow dependency updates (no workflows exist yet)
