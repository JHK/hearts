---
# hearts-3lkc
title: Add structured slog logging for lifecycle events
status: completed
type: feature
priority: normal
created_at: 2026-03-17T13:24:27Z
updated_at: 2026-03-17T16:36:26Z
---

Add log/slog structured logging for player/table lifecycle events with configurable log level

## Context

The server currently uses Go's unstructured `log` package with a single `Printf` call. There's no visibility into player or table lifecycle events, making it hard to debug issues in production or during development. Adding structured, level-aware logging to key events will make the system observable without requiring a full OTel SDK integration.

## Higher Goal

Enable operators and developers to understand what's happening in a running server — who connected, which tables were created or destroyed, when games started — at a glance and with machine-parseable output that can be ingested by log aggregators (Loki, Datadog, etc.) without transformation.

## Acceptance Criteria

- [ ] All logs emitted via `log/slog` with JSON handler to stdout
- [ ] Log level configurable via `-log-level` CLI flag and `LOG_LEVEL` env var (flag takes precedence); accepts `debug`, `info`, `warn`, `error`
- [ ] Default log level is `info`; container image default is `warn` (set in `.ko.yaml` or `Dockerfile` env)
- [ ] The following events are logged at the indicated levels:
  - `info`: table created, table started, table destroyed
  - `info`: player connected (WebSocket), player disconnected
  - `info`: player joined table, player left table
  - `warn`: table orphaned (all players gone, pending cleanup)
  - `debug`: any other lower-level events useful for tracing state transitions
- [ ] Each log entry includes structured fields: `table_id`, `player_id`, `event`, and any other relevant context (e.g. `addr` for connect/disconnect)
- [ ] All existing `log.Printf` / `log.Fatalf` calls audited and migrated to `slog` at appropriate levels
- [ ] `architecture.md` updated to document the logging approach: `slog`, JSON to stdout, level convention, key field names

## Out of Scope

- OTel SDK, traces, spans, or metrics
- Log shipping / sidecar configuration
- Log rotation or file output
- Changing any game logic or table behavior

## Summary of Changes

Replaced unstructured log package with log/slog JSON handler. Added lifecycle events (table created/started/destroyed, player connected/disconnected/joined/left, table orphaned, bot added) at appropriate levels. Log level configurable via -log-level flag or LOG_LEVEL env var; container default set to warn in .ko.yaml. Documented logging approach in architecture.md.
