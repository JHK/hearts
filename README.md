# hearts

Multiplayer Hearts in Go using embedded NATS.

There is CLI binary. After launch, each player can either open a game locally or discover and join a game bus.

## Quick start

1. In each terminal, launch the same CLI:

   ```bash
   go run ./cmd/hearts -name Alice
   ```

2. In one terminal, open a table (this starts embedded NATS + table authority):

   ```text
   open demo
   ```

3. In other terminals, connect and discover/join:

   ```bash
   go run ./cmd/hearts -name Bob
   ```

   Then inside the CLI:

   ```text
   connect nats://127.0.0.1:4222
   discover
   join demo
   ```

4. From any player at the table, start the round:

   ```text
   start
   ```

   If fewer than 4 humans are seated, the table auto-fills with simple bots.

5. Play cards:

   ```text
   play 2C
   play QS
   ```

## Project structure

- `cmd/hearts`: unified CLI (open, discover, join, play)
- `internal/game`: card model, rule checks, trick/scoring logic
- `internal/protocol`: NATS subjects + wire messages
- `internal/table`: authoritative table state machine

## NATS subjects

- `hearts.discovery`: request/reply table discovery
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
- Missing seats are auto-filled with random-play bots on `start`

Passing cards and full game-to-100 flow are not implemented yet.
