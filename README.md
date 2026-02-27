# hearts

Multiplayer Hearts in Go using embedded NATS.

The CLI is an orchestration surface for running and joining tables.

Architecture, boundaries, and concurrency model are documented in `architecture.md`.

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

4. Fill seats, then start the round:

   ```text
   addbot
   addbot
   addbot
   ```

   Add as many bots as needed until the table has 4 players.
   Each `addbot` spawns an ephemeral local bot player in that CLI process.

5. From any seated player, start the round:

   ```text
   start
   ```

6. Play cards:

   ```text
   play 2C
   play QS
   ```

## Target interaction model

The intended CLI flow is: start server, connect, discover table, join table, add bot, start game, play moves, and inspect table stats.

For role ownership (CLI/server/table/player), interface abstractions, and agent-based concurrency details, see `architecture.md`.
