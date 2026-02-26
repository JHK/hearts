# Architecture

This project is a multiplayer Hearts system built around NATS, with each participant modeled as an independent agent.

## High-level design

- `internal/game` contains Hearts domain types and logic, reusable by both players and table agents.
- Players and table are separate agents that only communicate through NATS messages/events.
- Each agent owns its own state and processes messages in its own loop (actor-style boundary).
- `internal/protocol` defines subjects and wire message/event contracts between agents.
- Bots are players with automated decision-making, not a special protocol role.

## Agent roles

### Player agent

- Represents a human or bot participant.
- Provides a display name on join and receives a canonical player ID from the table.
- Subscribes to table/public events and player-private events.
- Maintains local state (hand, turn context, table info).
- Uses `internal/game` to validate/candidate-check local moves before publishing `play` commands.

### Table agent

- Owns authoritative table and round state.
- Handles join/start/play command subjects.
- Owns player identity allocation by assigning and returning canonical IDs on join.
- Validates all incoming plays using `internal/game`.
- Emits table-wide events (turn changes, cards played, trick/round completion).
- Emits private player events (for example hand updates and your-turn notifications).

### Bot player agent

- Uses the same subjects and event flow as any other player.
- Consumes the same player/private events and table/public events.
- Chooses moves automatically (for now random strategy).
- Uses `internal/game` rules locally to filter to valid moves before sending a play.

## Shared domain and protocol

### Game domain (`internal/game`)

- Card model and utilities (parse/format/deck/sort).
- Rules for legal leads/plays and first-trick constraints.
- Trick winner and points calculation.
- Round scoring including shoot-the-moon handling.

### Protocol (`internal/protocol`)

- Subject builders for discovery, commands, and events.
- Request/response types for commands.
- Event envelope and typed payloads.

## Message topology

- Discovery: `hearts.discovery` (request/reply)
- Commands:
  - `hearts.table.<id>.join`
  - `hearts.table.<id>.start`
  - `hearts.table.<id>.play`
- Public events: `hearts.table.<id>.events`
- Private events: `hearts.table.<id>.player.<player_id>.events`

## Round lifecycle

1. Player agents discover and join a table.
2. On join, table assigns canonical player IDs and returns them to joining players.
3. At round start, players execute the passing phase and receive updated hands.
4. Table agent sets first turn and distributes per-player hand updates for trick play.
5. Active player agent sends play command; table validates/applies and broadcasts outcomes.
6. After each trick, table computes winner/points and emits updates.
7. After 13 tricks, table publishes round totals and resets round state.
