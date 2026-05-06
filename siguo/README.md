# 四国军棋

A self-contained browser 四国军棋 server. The current build is a playable dependency-free vertical slice: room codes, reconnect tokens, hidden-info views, setup randomization, click-to-move play, combat adjudication, chat, and a single embedded web UI.

## Run

```sh
/opt/homebrew/bin/go run ./cmd/siguo
```

Open `http://localhost:8080`, create a room, and share the six-character room code with three other players.

## Build And Test

```sh
make test
make build
./bin/siguo
```

The binary serves the embedded frontend and API on one port.

## Notes

- v1 uses in-memory rooms by default.
- `--persist`, `--db-path`, and `--metrics` are reserved CLI flags for the next persistence/ops pass.
- The frontend is dependency-free for now because the local project does not include `pnpm`; the backend package boundaries still leave room for replacing it with Svelte later.
