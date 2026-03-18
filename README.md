# hearts

Multiplayer Hearts in Go with a web-only architecture.

Architecture, boundaries, and concurrency model are documented in `architecture.md`.

## Quick start

Install project dependencies and start the server:
```bash
mise dev
```

Then open `http://127.0.0.1:8080/` in your browser.

## Container image

Builds use [ko](https://ko.build) with a distroless base image (see `.ko.yaml`). `ko` is managed by mise — no separate install needed.

Build and load into the local podman daemon:

```bash
mise run image-build
```

Build and push to `docker.io/julianknocke/hearts`:

```bash
mise run image-push
```
