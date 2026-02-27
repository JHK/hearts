# hearts

Multiplayer Hearts in Go with a web-only architecture.

Architecture, boundaries, and concurrency model are documented in `architecture.md`.

## Quick start

1. Start the web server:

   ```bash
   go run ./cmd/hearts
   ```

2. Open the lobby in your browser:

   `http://127.0.0.1:8080/`

3. In `/`:

   - choose player name
   - create a table or join an existing table
   - navigate to `/table/<table_id>`

4. In `/table/<table_id>`:

   - join as this browser player
   - add bots until 4 seats are filled
   - start a round
   - play cards when it is your turn

## Route model

- `/`: lobby for name + table selection/creation
- `/table/<table_id>`: gameplay page
- `/ws/table/<table_id>`: websocket channel for commands/events

## Interaction model

- Each browser instance is its own player.
- The table runtime is authoritative for all legal action checks and outcomes.
- Bots use the same command path as human players.

## State and persistence

- In-memory only.
- Server restart clears tables, players, and game progress.
