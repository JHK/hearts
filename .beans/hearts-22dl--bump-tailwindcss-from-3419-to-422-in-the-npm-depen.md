---
# hearts-22dl
title: Bump tailwindcss from 3.4.19 to 4.2.2 in the npm-dependencies group
status: completed
type: task
priority: normal
created_at: 2026-03-30T15:04:47Z
updated_at: 2026-03-30T15:10:30Z
---

Update tailwindcss to v4.2.2 with new features and improvements

## Context

Imported from PR #3 by app/dependabot.

Update tailwindcss to v4.2.2 with new features and improvements

## Acceptance Criteria

- [x] Changes reviewed and validated
- [x] Build passes with changes applied
- [x] Decision made: merge, modify, or close the PR

## References

- PR #3: dependabot/npm_and_yarn/npm-dependencies-0ec2bb0bc8

## Summary of Changes

Migrated from Tailwind CSS v3 to v4:
- Replaced `@tailwind` directives with `@import "tailwindcss"` and `@source` directives in `styles.input.css`
- Added `@tailwindcss/cli` package (v4 moved CLI to a separate package)
- Removed `-c tailwind.config.js` flag from npm scripts (v4 auto-detects sources from CSS)
- Deleted `tailwind.config.js` (empty config, not needed in v4)
- Removed stale `tailwind.config.js` reference from `mise.toml` sources
- Restored unrelated bean file that Dependabot branch had reverted
