---
# hearts-h9d0
title: Research dependency update automation for Go and Node dependencies
status: completed
type: task
priority: normal
created_at: 2026-03-18T06:55:54Z
updated_at: 2026-03-30T09:46:58Z
---

## Context

The project has Go module dependencies and a single Node devDependency (Tailwind CSS) that
are updated manually. The indirect `golang.org/x/*` packages in go.mod are pinned to 2021
commit hashes. There is no `.github` directory and no automation in place.

The repo is currently **private**, which may affect tool availability, pricing, or feature
access. This needs to be confirmed before any implementation work begins.

## Open Questions (resolved)

- ~~Does the repo's private status restrict access to Dependabot or Renovate's free tier?~~ Repo is now public; both tools are fully free.
- ~~Is there a plan/timeline to open-source the repo, and should this work wait for that?~~ Repo is public as of 2026-03-30.

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

## Summary of Changes

Evaluated Dependabot vs Renovate for a public GitHub repo with Go modules and one npm devDependency. Both are free for public repos. Chose **Dependabot** for simplicity — GitHub-native, single YAML config, supports grouping. Renovate is more powerful (built-in auto-merge, richer scheduling) but overkill for this repo's small dependency surface.

**Decision:** Dependabot with weekly grouped updates.
**Follow-up:** hearts-yvvb (implementation ticket)
