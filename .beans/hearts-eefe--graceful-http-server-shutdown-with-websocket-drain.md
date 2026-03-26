---
# hearts-eefe
title: Graceful HTTP server shutdown with WebSocket drain
status: todo
type: feature
created_at: 2026-03-26T09:10:25Z
updated_at: 2026-03-26T09:10:25Z
parent: hearts-p6hh
---

Replace http.ListenAndServe with http.Server + srv.Shutdown(ctx) to support graceful shutdown. On SIGINT/SIGTERM, signal active tables, drain WebSocket connections, and exit cleanly. Currently all in-memory game state is lost instantly on restart with no warning to connected players.
