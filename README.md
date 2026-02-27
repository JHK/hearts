# hearts

Multiplayer Hearts in Go using embedded NATS.

The CLI is an orchestration surface for running and joining tables.

Architecture, boundaries, and concurrency model are documented in `architecture.md`.

## Quick start

1. In one terminal, start a host (embedded NATS + table authority):

   ```bash
   go run ./cmd/hearts host --table default --host 127.0.0.1 --port 4222
   ```

2. In each player terminal, launch the interactive CLI:

   ```bash
	go run ./cmd/hearts cli --name Alice --server nats://127.0.0.1:4222
   ```

3. In the interactive CLI, discover/join:

   ```text
	discover
	join
   ```

   `discover` lists available tables and their IDs.
   `join` defaults to the `default` table. Use `join <table-id>` to join a specific table.

4. Fill seats, then start the round:

   ```text
	addbot
	addbot first-legal
   ```

   Add as many bots as needed until the table has 4 players.
   `addbot` accepts an optional `[strategy]` argument.
   Strategies: `random` (default), `first-legal`.

5. From any seated player, start the round:

   ```text
	start
   ```

6. Play cards:

   ```text
	play 2C
	play QS
   ```

## Web mode

You can also run a simple web server for gameplay:

```bash
go run ./cmd/hearts web --name Alice --server nats://127.0.0.1:4222 --table default --addr 127.0.0.1:8080
```

The web server exposes a minimal UI to reconnect/join, add bots, start, and play cards.

## Legacy in-CLI hosting

The interactive CLI still supports in-session hosting via:

```text
open [table-id] [port]
```

This is useful for quick local runs, but `hearts host` is the dedicated host mode.

## Command summary

```bash
go run ./cmd/hearts cli  --name Alice --server nats://127.0.0.1:4222
go run ./cmd/hearts host --table default --host 127.0.0.1 --port 4222
go run ./cmd/hearts web  --name Alice --server nats://127.0.0.1:4222 --table default --addr 127.0.0.1:8080
```

In `cli` mode, the command surface remains:

```text
open [table-id] [port]
discover
join [table-id]
connect <server>
addbot [strategy]
start
play <card>
hand
stats
status
help
quit
```

## Target interaction model

The intended CLI flow is: start server, connect, discover table, join table, add bot, start game, play moves, and inspect table stats.

For role ownership (CLI/server/table/player), interface abstractions, and agent-based concurrency details, see `architecture.md`.
