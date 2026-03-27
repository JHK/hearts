# Architecture

This document defines the architecture for the web-only Hearts application.

## Core principles

- Web-only interaction model: players use the browser UI.
- No NATS transport: server and session runtime are in-process.
- The table is the single authoritative owner of Hearts rules, game loop, and scoring.
- Human players and bots use the same command path and validation rules.
- Concurrency uses actor-style goroutines for owned mutable state.
- Runtime state is in-memory only and is lost on process restart.

## Routes and UX model

- `/` is the lobby page.
  - Choose player name.
  - Create a new table or join an existing table.
- `/table/<table_id>` is the gameplay page.
  - Join the table as the current browser player.
  - Add bots, start round, and play cards.
  - Render game events and per-player hand updates.
- `/ws/table/<table_id>` is the real-time channel.
  - Browser sends commands (join/start/play/pass/rename/add_bot/claim_seat/rematch/resume_game).
  - Server pushes table events and private player updates.

Each browser instance represents one player identity for a table.

## Responsibilities

### Web app (`cmd/hearts`, `internal/webui`)

- Starts the HTTP server.
- Serves embedded HTML assets and generated CSS/JS assets with content-hash fingerprinting for cache busting.
- Hosts lobby and table routes.
- Manages WebSocket lifecycle and message IO.
- Does not decide card legality or game outcomes.

### Session manager (`internal/session`)

- `session.Table` is the authoritative runtime for one game session.
- `session.Manager` creates and tracks in-memory sessions; provides discovery/listing for lobby UX.
- Routes commands to the correct `Table` via its command channel.
- Owns no Hearts rule decisions directly.

### Table runtime (authoritative agent)

- Owns table-level mutable state; delegates round state to `game.Round`.
- Validates player inputs by calling `game.Round` methods, which apply game rules internally.
- Assigns canonical `player_id` values and seat positions. A `player_id` is the stable string identity used across all protocol events; it is owned by the session and lives in `internal/protocol`.
- Each seat is represented by `playerState`, which carries: web-transport state (`id`, `Token`), seated identity (`Name`, `position`), cumulative scoring, and a `bot.Bot` for bot seats (nil for humans).
- Emits public table events and private player events.
- Accepts commands from browser players; drives bot seats via `bot.Bot.ChoosePlay`/`ChoosePass`.

### Bot automation (`internal/game/bot`)

- `bot.Bot` is the decision interface: `ChoosePlay(TurnInput)` and `ChoosePass(PassInput)` return decisions directly (no callbacks). `Kind()` provides strategy metadata. Bots hold no player state — game state is managed by `game.Round`.
- Bot names are assigned at the table layer via `StrategyKind.BotName()`; name pools live in each strategy file.
- Runs as local server-side agents; never bypasses table authority.
- Bot detection in the session is `player.bot != nil` (set in `playerState` when a bot occupies the seat).

### Domain logic (`internal/game`)

- Owns Hearts rules, scoring, and per-round game state.
- `game.Round` is a step-at-a-time state machine for one round: owns hands, trick state, pass state, and scoring. Callers drive phase transitions (`SubmitPass` → `ApplyPasses` → `MarkReady` → `StartPlaying` → `Play`), giving the session hooks for event emission.
- `game.TurnInput`/`game.PassInput` are pure data types that carry decision context (hand, trick, flags) to bots and UIs.
- Pure functions handle card validation, trick resolution, and pass direction.
- Contains no networking, persistence, or websocket concerns.

## Concurrency model

- Each session runtime runs in a dedicated goroutine.
- Commands are serialized through the session runtime loop.
- Websocket clients and bots communicate with session runtimes via messages.
- Mutable game state stays inside the owning runtime.

## State model

- In-memory only.
- No persistence across restarts.
- Reconnects can recover identity within a running process via browser token/cookie, but not after restart.

## Frontend resilience

The browser client should retry on transient, potentially recoverable failures (network blips, brief server unavailability) before giving up. Terminal conditions (server explicitly rejects the request) should fail fast. The goal is to avoid bouncing users out of a game due to a momentary glitch while still redirecting promptly when the resource is genuinely gone.

## Authority rules

- Only the session runtime validates and commits game actions.
- Only the session runtime assigns canonical player IDs and seats.
- Clients may pre-validate for UX, but server validation is final.

## Logging

All logs are emitted via `log/slog` with a JSON handler to stdout.

**Log level** is controlled by the `-log-level` CLI flag (`debug`, `info`, `warn`, `error`). If the flag is absent or empty, the `LOG_LEVEL` environment variable is used. The default level is `info`. The container image sets `LOG_LEVEL=warn` as the default.

**Event levels:**

| Level | Events |
|-------|--------|
| `info` | table created, table destroyed, table started (each round), player connected (WebSocket), player disconnected, player joined table, player left table, player renamed |
| `warn` | table orphaned (all human players disconnected, pending cleanup) |
| `debug` | bot added to table |

**Structured fields** included on each log entry where applicable:

- `event` — machine-readable event name (e.g. `table_created`, `player_joined`)
- `table_id` — the table identifier
- `player_id` — the player identifier
- `name` — player display name (join/leave events)
- `addr` — remote address (WebSocket connect/disconnect)
- `round` — round number (table started event)
