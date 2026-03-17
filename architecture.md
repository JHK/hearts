# Architecture

This document defines the architecture for the web-only Hearts application.

## Core principles

- Web-only interaction model: players use the browser UI.
- No NATS transport: server and table runtime are in-process.
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
  - Browser sends commands (join/start/play/add_bot).
  - Server pushes table events and private player updates.

Each browser instance represents one player identity for a table.

## Responsibilities

### Web app (`cmd/hearts`, `internal/webui`)

- Starts the HTTP server.
- Serves embedded HTML assets and generated CSS/JS assets.
- Hosts lobby and table routes.
- Manages WebSocket lifecycle and message IO.
- Does not decide card legality or game outcomes.

### Table manager (`internal/table`)

- Creates and tracks in-memory table runtimes.
- Provides table discovery/listing for lobby UX.
- Routes commands to the correct table runtime.
- Owns no Hearts rule decisions directly.

### Table runtime (authoritative agent)

- Owns table, round, trick, and score mutable state.
- Enforces all gameplay rules using `internal/game`.
- Assigns canonical `player_id` values and seats.
- Emits public table events and private player events.
- Accepts commands from browser players and bots.

### Bot automation (`internal/player/bot`)

- Strategy-only logic.
- Runs as local server-side agents.
- Uses the same command/event contract as human players.
- Never bypasses table authority.

### Domain logic (`internal/game`)

- Owns pure Hearts rules and scoring.
- Contains no networking, persistence, or websocket concerns.

## Concurrency model

- Each table runtime runs in a dedicated goroutine.
- Commands are serialized through the table runtime loop.
- Websocket clients and bots communicate with table runtimes via messages.
- Mutable game state stays inside the owning runtime.

## State model

- In-memory only.
- No persistence across restarts.
- Reconnects can recover identity within a running process via browser token/cookie, but not after restart.

## Authority rules

- Only the table runtime validates and commits game actions.
- Only the table runtime assigns canonical player IDs and seats.
- Clients may pre-validate for UX, but server validation is final.

## Logging

All logs are emitted via `log/slog` with a JSON handler to stdout.

**Log level** is controlled by the `-log-level` CLI flag (`debug`, `info`, `warn`, `error`). If the flag is absent or empty, the `LOG_LEVEL` environment variable is used. The default level is `info`. The container image sets `LOG_LEVEL=warn` as the default.

**Event levels:**

| Level | Events |
|-------|--------|
| `info` | table created, table destroyed, table started (each round), player connected (WebSocket), player disconnected, player joined table, player left table |
| `warn` | table orphaned (all human players disconnected, pending cleanup) |
| `debug` | bot added to table |

**Structured fields** included on each log entry where applicable:

- `event` — machine-readable event name (e.g. `table_created`, `player_joined`)
- `table_id` — the table identifier
- `player_id` — the player identifier
- `name` — player display name (join/leave events)
- `addr` — remote address (WebSocket connect/disconnect)
- `round` — round number (table started event)
