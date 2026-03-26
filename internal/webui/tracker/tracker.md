# tracker package

This package groups self-contained state trackers used by the webui
server layer. Each file exports its own API with no inner-package
dependencies — you can read and reason about each tracker in isolation.

The package exists to keep server.go focused on wiring and route
registration, while giving each tracker an explicit, exported interface
that documents the contract between the tracker and its consumers.

## Files

- **conn.go** — ConnTracker: tracks active WebSocket connections for
  graceful shutdown (actor pattern, single goroutine).
- **presence.go** — HumanPresence / PlayerPresence: ref-counted
  presence tracking per table and per player, used for leave detection
  across multi-tab sessions (actor pattern, single goroutine each).
