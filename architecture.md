# Architecture

This project follows an agent-oriented architecture with strict separation of concerns.

## Core principles

- The CLI only steers high-level actions and does not own game rules or state.
- The table is the single authoritative owner of Hearts rules, game loop, events, and scoring.
- A player is an entity that interacts solely with a table through commands/events.
- Human and bot players use the same transport command/event contract.
- Concurrency is handled through agents (actor-style goroutines), not shared mutable state.
- NATS parallelism and protocol marshalling are isolated in transport/protocol layers.

## Responsibilities

### CLI

- Starts a server.
- Connects to a server.
- Discovers tables.
- Joins a table.
- Adds bots to a table.
- Starts a game.
- Performs moves.
- Requests table stats.

The CLI is orchestration-only and remains intentionally thin.
Ephemeral bots are local player agents spawned by a CLI process and live for that process lifetime.

### Server

- Hosts NATS connectivity and table lifecycles.
- Routes discovery and table-level command traffic.
- Does not contain Hearts rule decisions.
- Does not own bot control APIs.

### Table (authoritative agent)

- Owns full table and round state.
- Enforces legal actions and turn progression.
- Assigns canonical player IDs when players join.
- Runs the game loop and scoring.
- Publishes public table events and private player events.
- Computes and serves table/game stats.

The table is the source of truth. No other component may finalize rule outcomes.

### Player agent

Players (human or bot) follow the same table command/event contract and remain table-type agnostic.

### Bot

- Implements the same player contract as humans.
- Contains strategy-only logic.
- Never bypasses table authority.

## Concurrency model (agents pattern)

- Each table runs in its own goroutine and processes messages sequentially.
- Each player runs in its own goroutine and processes its own incoming events/turn prompts.
- Agents communicate through command/event messages.
- Mutable game state stays inside the owning agent loop.
- This avoids shared-state races and keeps synchronization localized.

## Separation of concerns

### Orchestration layer (`cmd/hearts`, `cmd/hearts/app`)

- Owns CLI parsing and high-level user flows (connect/discover/join/addbot/start/play/stats).
- Keeps only client/session-facing state.
- Does not implement Hearts rules, table state transitions, or transport internals.

### Server hosting layer (`internal/server`)

- Owns embedded NATS lifecycle and hosted table runtime lifecycle.
- Owns hosted table runtimes.
- Does not decide game outcomes or card legality.

### Table authority layer (`internal/server` table runtime)

- Owns authoritative mutable state for table, trick, round, and totals.
- Enforces legal actions using `internal/game` and emits resulting events.
- Stays player-type agnostic (no human/bot branching in table rules).
- Is the only component allowed to commit game outcomes.

### Player automation layer (`internal/player/bot`)

- Defines bot strategies and bot runtime behavior.
- Reacts to table events and submits commands through transport clients.
- Ephemeral bot runtimes are spawned locally by CLI and join as normal players.
- Never bypasses table authority.

### Transport layer (`internal/transport/nats`)

- Owns NATS request/reply, subscriptions, event fan-out, and callback safety.
- Provides participant/table transport endpoints and codec helpers.
- Contains no Hearts rule logic or authoritative game state.

### Protocol layer (`internal/protocol`)

- Owns subject naming and wire contracts, split by lobby/table concerns and gameplay concerns.
- Defines serialization boundary between components.
- Contains no runtime game decisions.

### Domain layer (`internal/game`)

- Owns pure Hearts domain logic (cards, legal plays, trick winner, scoring).
- Owns game-level constants such as players-per-table.
- Reused by authoritative table logic and bot decision support.
- Contains no networking, goroutine orchestration, or transport concerns.

## Interaction flow

1. CLI starts or connects to a server.
2. CLI discovers available tables.
3. Player joins a table.
4. Optional bots are added by local CLI bot agents that join as normal players.
5. CLI starts the game only after the table has 4 players.
6. Players submit moves.
7. Table validates, applies, and emits outcomes.
8. CLI requests table stats at any time.

## Authority rules

- Only the table validates and commits game actions.
- Only the table allocates canonical `player_id` values.
- Players may pre-validate locally for UX, but table validation is final.
- Transport/protocol errors are handled outside domain logic boundaries.
