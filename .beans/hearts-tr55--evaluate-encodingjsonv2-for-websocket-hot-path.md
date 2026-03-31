---
# hearts-tr55
title: Evaluate encoding/json/v2 for WebSocket hot path
status: completed
type: task
priority: normal
tags:
    - backend
created_at: 2026-03-31T15:37:18Z
updated_at: 2026-03-31T15:51:47Z
parent: hearts-u20m
---

## Context

The codebase uses `encoding/json` extensively: WebSocket message envelopes (`ReadJSON`/`WriteJSON`), protocol contract structs, i18n locale loading, and API responses. All JSON usage is standard struct-based with no custom marshalers. Go 1.25 introduced `encoding/json/v2` (experimental via `GOEXPERIMENT=jsonv2`) with substantially faster decoding and stricter semantics (rejects duplicate keys, case-sensitive matching).

## Higher Goal

Reduce latency on the WebSocket message hot path and prepare for eventual json/v2 graduation to stable.

## Acceptance Criteria

- [x] Benchmark comparing json v1 vs v2 for representative message types (wsMessage, wsCommand, protocol events)
- [x] Document any semantic differences that would affect our structs (e.g. case sensitivity, duplicate key behavior)
- [x] Decision recorded: adopt now (behind GOEXPERIMENT), wait for stable, or skip
- [x] gorilla/websocket compatibility verified (it uses encoding/json internally) — works transparently, no code changes needed

## Out of Scope

- Custom marshalers or jsontext low-level API
- Migrating non-JSON serialization

## References

- [Go 1.25 release notes — encoding/json/v2](https://go.dev/doc/go1.25): experimental json/v2 package
- [encoding/json/v2 proposal](https://github.com/golang/go/discussions/63397): design rationale and semantic changes

## Summary of Changes

Benchmarked json v1 vs v2 across 8 representative message types (10 iterations each, benchstat comparison). Key findings:

- **Marshal (server→client, dominant direction): 30-112% slower** due to v1 compat shim overhead
- **Unmarshal (client→server): 40-49% faster** with dramatically fewer allocations
- **gorilla/websocket**: works transparently, all tests pass
- **Semantic differences**: non-issue since all structs use explicit json tags

**Decision: wait for stable.** The marshal regression on the broadcast-heavy hot path outweighs the unmarshal win. Absolute latencies are already sub-10µs.

Artifacts:
- `internal/protocol/json_bench_test.go` — reusable benchmarks
- `docs/json-v2-evaluation.md` — full analysis with data
