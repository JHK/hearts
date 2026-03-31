---
# hearts-tr55
title: Evaluate encoding/json/v2 for WebSocket hot path
status: todo
type: task
priority: normal
tags:
    - backend
created_at: 2026-03-31T15:37:18Z
updated_at: 2026-03-31T15:37:46Z
parent: hearts-u20m
---

## Context

The codebase uses `encoding/json` extensively: WebSocket message envelopes (`ReadJSON`/`WriteJSON`), protocol contract structs, i18n locale loading, and API responses. All JSON usage is standard struct-based with no custom marshalers. Go 1.25 introduced `encoding/json/v2` (experimental via `GOEXPERIMENT=jsonv2`) with substantially faster decoding and stricter semantics (rejects duplicate keys, case-sensitive matching).

## Higher Goal

Reduce latency on the WebSocket message hot path and prepare for eventual json/v2 graduation to stable.

## Acceptance Criteria

- [ ] Benchmark comparing json v1 vs v2 for representative message types (wsMessage, wsCommand, protocol events)
- [ ] Document any semantic differences that would affect our structs (e.g. case sensitivity, duplicate key behavior)
- [ ] Decision recorded: adopt now (behind GOEXPERIMENT), wait for stable, or skip
- [ ] If adopting: gorilla/websocket compatibility verified (it uses encoding/json internally)

## Out of Scope

- Custom marshalers or jsontext low-level API
- Migrating non-JSON serialization

## References

- [Go 1.25 release notes — encoding/json/v2](https://go.dev/doc/go1.25): experimental json/v2 package
- [encoding/json/v2 proposal](https://github.com/golang/go/discussions/63397): design rationale and semantic changes
