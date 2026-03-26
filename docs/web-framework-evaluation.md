# Web Framework Evaluation

Evaluating options to replace hand-rolled HTTP plumbing in `internal/webui/server.go`.

**Current state:** Go 1.24, `http.ServeMux` with manual routing, gorilla/websocket,
embedded assets via `embed.FS`, custom fingerprinting in `fingerprint.go`, pre-rendered
HTML templates with ETag support. ~713 lines in server.go.

## Options Evaluated

### 1. Go 1.22+ `net/http` (enhanced routing)

**What's new:** Method-based patterns (`"GET /table/{tableID}"`), path parameters via
`r.PathValue()`, automatic 405 responses.

**What it eliminates:**
- Manual method checking (`if r.Method != "GET"`)
- Manual path parsing (`strings.Split`, `strings.TrimPrefix`)

**What remains manual:**
- Route groups / scoped middleware (no built-in concept)
- Cache header middleware (wrap `FileServer` yourself)
- Template asset hashing, ETag computation
- WebSocket (still needs gorilla/websocket)

**Assessment:** The routing improvements are significant but this project already uses
Go 1.24 patterns (e.g. `"GET /table/{tableID}"` is available). The gap is middleware
composition and route groups — you'd write a small `chain()` helper and manually scope
middleware per handler.

### 2. Chi

**What it adds over net/http 1.22+:**
- `r.Route("/prefix", func(r chi.Router){...})` — declarative route groups
- `r.Use(mw)` scoped to groups — the single biggest ergonomic win
- Built-in middleware: Logger, Recoverer, RequestID, Compress, etc.
- `r.Mount()` for sub-routers

**What remains manual:** Static file serving, fingerprinting, templates, ETag — all
identical to net/http.

**Compatibility:** Full `net/http` compatibility. Handlers are plain `http.HandlerFunc`.
Middleware uses standard `func(http.Handler) http.Handler` signature. gorilla/websocket
works unchanged. `httptest` works unchanged.

**Community:** ~18k GitHub stars, actively maintained (v5.2.1, Dec 2024), MIT license.

**Assessment:** Chi's value is **route groups with scoped middleware**. For this codebase
— with distinct asset/page/WebSocket/API routes needing different cache/middleware — this
provides a real cleanup. Zero migration friction since it wraps net/http.

### 3. Echo

**What it adds over Chi:**
- Response helpers (`c.JSON()`, `c.String()`, `c.Render()`)
- Centralized error handling (handlers return `error`)
- `Renderer` interface for templates
- Richer built-in middleware (CORS, Gzip, CSRF, etc.)

**Trade-offs:**
- **Own `echo.Context` type** — handlers use `func(echo.Context) error`, not
  `http.HandlerFunc`. All net/http middleware needs wrapping via `echo.WrapMiddleware()`.
- Testing requires constructing `echo.Context` instead of `httptest.NewRequest`.
- WebSocket: still uses gorilla/websocket directly — zero value-add.
- Static serving: `StaticFS()` is a thin wrapper, no fingerprinting support.
- Template rendering: `Renderer` interface is 4 lines — near-zero value-add.
- v5 has been in limbo for years.

**Assessment:** Echo's response helpers save ~3 lines per handler but are mostly
irrelevant for a WebSocket-heavy game server. The lock-in to `echo.Context` is a
significant cost for minimal benefit. Most communication flows through WebSockets, not
REST handlers.

### 4. Fiber

**What it adds:**
- Express-like API, good ergonomics
- Large middleware ecosystem
- High raw HTTP throughput (fasthttp)
- Built-in WebSocket support (fasthttp/websocket fork)

**Critical issues:**
- **Built on fasthttp, not net/http.** gorilla/websocket is incompatible — all WebSocket
  code must be rewritten. `httptest` doesn't work. Standard net/http middleware doesn't
  work.
- **No HTTP/2 support** in fasthttp.
- **Pooled context lifecycle** — `*fiber.Ctx` is reused after handler returns. Holding
  references in goroutines (which the table actor model does heavily) causes data races.
- v2→v3 migration ahead with breaking changes.
- The `adaptor` package for net/http compat negates performance benefits.

**Assessment:** Disqualified. The fasthttp foundation creates meaningful friction that
outweighs any benefit for a LAN card game. The gorilla/websocket incompatibility alone
means rewriting the entire WebSocket layer, and the pooled context lifecycle conflicts
with the goroutine-per-table actor model.

## Comparison Matrix

| Concern | net/http 1.22+ | Chi | Echo | Fiber |
|---|---|---|---|---|
| Route groups | Manual | Built-in | Built-in | Built-in |
| Scoped middleware | Manual | Built-in | Built-in | Built-in |
| net/http compat | Native | Full | Partial (wrapping) | Incompatible |
| gorilla/websocket | Works | Works | Works | Incompatible |
| httptest | Works | Works | Needs echo.Context | Needs app.Test() |
| Static + embed.FS | Manual | Manual | Thin wrapper | Middleware |
| Fingerprinting | Manual | Manual | Manual | Manual |
| Template ETag | Manual | Manual | Manual | Manual |
| Migration effort | Minimal | Minimal | Moderate | Major |
| Dependency weight | Zero | Light (~1 pkg) | Medium (~3 pkgs) | Heavy (fasthttp) |

## Recommendation: Chi

**Chi is the right choice** for this codebase.

**Why Chi over net/http 1.22+:** Route groups with scoped middleware eliminate real
boilerplate. This codebase has distinct route categories (static assets with immutable
cache, HTML pages with ETag/no-cache, WebSocket endpoints, API endpoints, dev-only
routes) that naturally map to Chi groups with different middleware stacks.

**Why Chi over Echo:** Full net/http compatibility means zero migration friction for
existing handlers, middleware, WebSocket code, and tests. Echo's value-add (response
helpers) is marginal for a WebSocket-heavy app where most handlers just upgrade to WS.

**Why Chi over Fiber:** Fiber is disqualified due to fasthttp incompatibility with
gorilla/websocket and the pooled context lifecycle conflicting with the actor model.

**What Chi doesn't solve:** Fingerprinting, template ETag, and static asset serving
remain manual regardless of framework choice. Chi cleans up the routing/middleware layer
while leaving the asset pipeline untouched — which is fine, since the existing
fingerprinting code in `fingerprint.go` works well.

### Expected impact

Adopting Chi would:
1. Replace ~40 lines of manual route registration with grouped routes
2. Eliminate per-handler cache header repetition via scoped middleware
3. Add panic recovery and structured request logging as one-liners
4. Keep all existing handler, WebSocket, and test code working unchanged
5. Add one lightweight dependency (~1 package, MIT license)
