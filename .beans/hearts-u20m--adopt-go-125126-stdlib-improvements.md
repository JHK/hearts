---
# hearts-u20m
title: Adopt Go 1.25/1.26 stdlib improvements
status: todo
type: epic
priority: normal
tags:
    - backend
created_at: 2026-03-31T15:36:01Z
updated_at: 2026-03-31T15:37:13Z
---

## Vision

Adopt valuable Go 1.25/1.26 standard library additions to improve performance, test reliability, and code quality. The codebase is already on Go 1.26.1 so all features are available now.

## Context

Go 1.25 and 1.26 introduced several features directly relevant to this codebase: modernization fixers, experimental json/v2 with better performance, virtualized time for testing concurrent code, and generic error unwrapping. Each child ticket addresses one adoption.

## Out of Scope

- Crypto/TLS changes (no relevant use case)
- CrossOriginProtection (doesn't cover WebSocket upgrades)
- Platform-specific or WASM changes
