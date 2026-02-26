# hearts

Multiplayer Hearts in Go using embedded NATS.

This project uses a host-peer architecture: one player runs `heartsd`, which starts an embedded NATS server and the authoritative table engine. Other players run `hearts-cli` and connect to that host.

## Quick start

1. Start host (table authority + embedded NATS):

   ```bash
   go run ./cmd/heartsd -table demo -host 127.0.0.1 -port 4222
   ```

2. In four terminals, join as players:

   ```bash
   go run ./cmd/hearts-cli -url nats://127.0.0.1:4222 -table demo -name Alice
   go run ./cmd/hearts-cli -url nats://127.0.0.1:4222 -table demo -name Bob
   go run ./cmd/hearts-cli -url nats://127.0.0.1:4222 -table demo -name Carol
   go run ./cmd/hearts-cli -url nats://127.0.0.1:4222 -table demo -name Dave
   ```

3. From any client, start the round:

   ```text
   start
   ```

4. Play cards:

   ```text
   play 2C
   play QS
   ```

## Project structure

- `cmd/heartsd`: embedded NATS host + table service
- `cmd/hearts-cli`: terminal client
- `internal/game`: card model, rule checks, trick/scoring logic
- `internal/protocol`: NATS subjects + wire messages
- `internal/table`: authoritative table state machine

## NATS subjects

- `hearts.table.<id>.join`: request/reply join seat
- `hearts.table.<id>.start`: request/reply start round
- `hearts.table.<id>.play`: request/reply play card
- `hearts.table.<id>.events`: broadcast table events
- `hearts.table.<id>.player.<player_id>.events`: private player events (hand updates)

## Current rules coverage

- 4 players, 52-card deal, 13 tricks
- First trick must start with `2C`
- Follow suit is enforced
- Hearts cannot be led before broken (unless hand is all hearts)
- Penalty cards blocked on first trick unless no alternative
- Round scoring includes shoot-the-moon handling

Passing cards and full game-to-100 flow are not implemented yet.
