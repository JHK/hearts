---
# hearts-mgme
title: Replace Makefile with mise tasks
status: in-progress
type: task
priority: normal
created_at: 2026-03-15T16:41:37Z
updated_at: 2026-03-15T16:43:12Z
---

Replace the Makefile with mise task targets, so all dev commands are defined in mise.toml instead.

## Tasks

- [x] Migrate all Makefile targets to mise tasks (setup, run, fmt, test, css, css-watch)
- [x] Delete the Makefile
- [x] Update README.md (Quick start + Frontend styling workflow sections reference make)
- [x] Update CLAUDE.md (Commands section references make)
