# hearts

Multiplayer Hearts in Go with a web-only architecture.

Architecture, boundaries, and concurrency model are documented in `architecture.md`.

## Quick start

1. Install project dependencies (Node + Go modules):

   ```bash
   mise run setup
   ```

2. Start the web server:

   ```bash
   mise run start
   ```

3. Open the lobby in your browser:

   `http://127.0.0.1:8080/`

4. In `/`:

   - choose player name
   - create a table or pick one from Open Tables
   - navigate to `/table/<table_id>`

5. In `/table/<table_id>`:

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

## Frontend styling workflow

- Build CSS once: `mise run css`
- Rebuild on changes: `mise run css-watch`

## Distroless image with ko

1. Install `ko`:

   ```bash
   go install github.com/google/ko@latest
   ```

2. Build a local distroless image (loaded into your Docker daemon):

   ```fish
   set -x IMAGE (ko build --local ./cmd/hearts)
   ```

3. Push an image to a registry:

   ```fish
   podman tag $IMAGE ghcr.io/julianknocke/hearts:latest
   podman push ghcr.io/julianknocke/hearts:latest
   ```

The ko config is in `.ko.yaml` and uses a distroless base image.
