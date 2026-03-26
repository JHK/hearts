---
# hearts-c4hi
title: Structured request logging middleware with slog
status: completed
type: feature
priority: low
created_at: 2026-03-26T09:10:41Z
updated_at: 2026-03-26T12:33:55Z
parent: hearts-p6hh
---

Add a custom request logging middleware using slog for consistent structured JSON logs (method, path, status, duration). Scope to API and page groups. Replaces ad-hoc slog.Info calls in WebSocket handlers with a uniform pattern.

## Summary of Changes

Added `requestLoggingMiddleware` in `internal/webui/middleware_logging.go` that logs each HTTP request with method, path, status, and duration_ms using slog. Applied to the API route group and a new page route group (HTML pages only, not static assets or WebSocket endpoints). The middleware uses a `statusWriter` wrapper with `Unwrap()` for proper interface propagation.
