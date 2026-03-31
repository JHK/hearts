# encoding/json/v2 Evaluation

Evaluated 2026-03-31 on Go 1.26.1 with `GOEXPERIMENT=jsonv2`.

## Decision: Wait for stable

Do not adopt json/v2 now. The marshal regression (the dominant hot path direction)
outweighs the unmarshal improvement. Revisit when v2 graduates from experimental.

## Benchmark Results (benchstat, 10 iterations)

Benchmarks cover representative WebSocket message types: `wsCommand` (client→server
envelope), `CardPlayedData` (frequent game event), `TrickCompletedData`,
`RoundCompletedData` (map-heavy), and `Snapshot` (largest payload, full game state).

### Latency (sec/op)

| Benchmark                 | v1 (baseline) | v2 (GOEXPERIMENT) | Change  |
|---------------------------|---------------|--------------------|---------|
| Marshal_WsCommand         | 170 ns        | 230 ns             | +35%    |
| Marshal_CardPlayed        | 172 ns        | 365 ns             | +112%   |
| Marshal_TrickCompleted    | 192 ns        | 374 ns             | +95%    |
| Marshal_RoundCompleted    | 965 ns        | 1244 ns            | +29%    |
| Marshal_Snapshot          | 3.0 µs        | 4.1 µs             | +37%    |
| **Unmarshal_WsCommand**   | **595 ns**    | **317 ns**         | **-47%**|
| **Unmarshal_CardPlayed**  | **455 ns**    | **231 ns**         | **-49%**|
| **Unmarshal_Snapshot**    | **9.8 µs**    | **5.9 µs**         | **-40%**|
| **geomean**               | **675 ns**    | **716 ns**         | **+6%** |

### Allocations (B/op)

| Benchmark                 | v1        | v2       | Change |
|---------------------------|-----------|----------|--------|
| Marshal_WsCommand         | 48 B      | 48 B     | ~      |
| Marshal_CardPlayed        | 64 B      | 160 B    | +150%  |
| Marshal_TrickCompleted    | 96 B      | 160 B    | +67%   |
| Marshal_RoundCompleted    | 656 B     | 288 B    | -56%   |
| Marshal_Snapshot          | 1985 B    | 1626 B   | -18%   |
| Unmarshal_WsCommand       | 368 B     | 128 B    | -65%   |
| Unmarshal_CardPlayed      | 272 B     | 48 B     | -82%   |
| Unmarshal_Snapshot        | 2968 B    | 2321 B   | -22%   |
| **geomean**               | **322 B** | **231 B**| **-28%**|

## Analysis

### Marshal is slower (30-112%)

With `GOEXPERIMENT=jsonv2`, `encoding/json.Marshal` calls `jsonv2.Marshal` with
`DefaultOptionsV1()` — the v1 compatibility shim adds overhead. Small structs
marshaled through an `any` interface (like `wsMessage.Data`) are hit hardest because
v2 does more reflection work to discover the concrete type.

This matters: the server broadcasts far more messages than it receives (every card
play, trick completion, turn change goes to all 4 clients). Marshal is the dominant
direction on the hot path.

### Unmarshal is faster (40-49%)

v2's decoder is genuinely faster with dramatically fewer allocations (1 alloc vs 7-8
for small types, 32 vs 94 for Snapshot). This is the headlining v2 improvement.

However, unmarshal only happens once per client command (play a card, pass cards),
which is far less frequent than the broadcast marshal path.

### gorilla/websocket compatibility

gorilla/websocket v1.5.3 hardcodes `encoding/json` in `ReadJSON`/`WriteJSON`. With
`GOEXPERIMENT=jsonv2`, this transparently uses v2 internals with v1 semantics — no
code changes needed, no breakage observed. All tests pass.

However, this also means we cannot use v2 directly (without the compat layer) through
gorilla's API. To get pure v2 marshal performance, we would need to marshal manually
and use `WriteMessage` instead of `WriteJSON`, which couples us to an experimental API.

### Semantic differences

With `DefaultOptionsV1()` (what gorilla/websocket uses implicitly), v1 semantics are
preserved: case-insensitive field matching, duplicate keys allowed, lenient behavior.
No struct changes needed.

If using `encoding/json/v2` directly with v2 defaults:
- **Case-sensitive matching**: our structs use explicit `json:"..."` tags on all fields,
  so this is a non-issue.
- **Duplicate key rejection**: not relevant (we don't produce duplicate keys).
- **Stricter error reporting**: would surface any currently-silent issues (positive).

### Why not adopt behind GOEXPERIMENT now?

1. Marshal regression is the wrong direction for our workload (broadcast-heavy).
2. `GOEXPERIMENT` is explicitly "not subject to Go 1 compatibility promise."
3. CI/deployment complexity of setting build flags for an experimental feature.
4. The win (unmarshal) is on the lighter code path.
5. Absolute latencies are already sub-10µs — not a bottleneck.

### When to revisit

- When `encoding/json/v2` graduates to stable (removes GOEXPERIMENT requirement).
- If gorilla/websocket adds a v2-native JSON option.
- If marshal performance improves in a future Go release (the v1 compat shim is a
  known overhead that the Go team is working to reduce).
