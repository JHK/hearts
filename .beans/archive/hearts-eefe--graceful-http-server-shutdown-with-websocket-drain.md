---
# hearts-eefe
title: Graceful HTTP server shutdown with WebSocket drain
status: completed
type: feature
priority: normal
created_at: 2026-03-26T09:10:25Z
updated_at: 2026-03-26T09:16:43Z
parent: hearts-p6hh
---

Replace http.ListenAndServe with http.Server + srv.Shutdown(ctx) to support graceful shutdown. On SIGINT/SIGTERM, signal active tables, drain WebSocket connections, and exit cleanly. Currently all in-memory game state is lost instantly on restart with no warning to connected players.

## Summary of Changes

Replaced `http.ListenAndServe` in `webui.Run()` with `http.Server` + `srv.Shutdown(ctx)`. On SIGINT/SIGTERM the server now:
1. Stops accepting new connections
2. Waits up to 5 seconds for in-flight HTTP requests and WebSocket connections to drain
3. Closes all active tables (which stops table goroutines and closes subscriber channels)
4. Exits cleanly

The session manager is now created in `Run()` rather than lazily in `NewHandler()`, so it's available for the shutdown sequence.
