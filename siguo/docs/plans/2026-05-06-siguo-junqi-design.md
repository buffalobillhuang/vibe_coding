# 四国军棋 (siguo) — Design

**Date:** 2026-05-06
**Status:** Approved

A LAN-or-cloud-hostable 4-player 四国军棋 game, two teams of two (diagonal partners), played in a web browser.

---

## 1. Decisions

| Area | Decision |
|---|---|
| Rules | 暗棋 (hidden pieces) with server-side auto-referee, standard Chinese ruleset |
| Teammate visibility | Strict — teammates cannot see each other's pieces |
| Players | Exactly 4, fixed 2v2 diagonal teams (North+South vs East+West) |
| Client | Web app, single URL, no install |
| Backend | Go, single static binary, embedded frontend via `//go:embed` |
| Frontend | Svelte 5 (runes) + Vite + TypeScript + Tailwind |
| Deployment | Same binary for LAN host or cloud VPS; Docker image published per release |
| Persistence | None by default; optional SQLite (`--persist`) for replay log + reconnect tokens |
| Lobby | 6-character room codes; no public list; no mDNS |
| Setup UX | Drag-drop placement + randomize button; server validates legality |
| Time control | Per-room configurable; default 3 min setup + 30 s/move with 5 s increment |
| Disconnect | Game pauses; 90 s reconnect window via session token; team can vote-abandon early |
| Chat | All-table + team channels; 200-char text + emote presets; rate-limited; team-chat disable-able |
| Sound | Per-player client-side toggle, defaults off, persisted in localStorage |

## 2. Non-goals (v1)

AI bots, spectators, voice chat, user accounts, ELO/match history, public room lists, matchmaking queues, saved setup templates, mDNS LAN discovery, profanity filtering, English UI, rule-variant toggles (flying 司令 etc.), mobile-native apps, tournaments.

The boundaries between `game/`, `room/`, `ws/`, and `web/` are kept clean so each of the above can be added later without rearchitecting.

## 3. Architecture

Single Go binary. One HTTP port serves three things:

1. Static frontend (Svelte build output, embedded).
2. REST endpoints for pre-game handshake (`POST /api/rooms`, `POST /api/rooms/:code/join`, `GET /api/rooms/:code`).
3. WebSocket endpoint (`GET /ws?room=...&token=...`) for all in-game messages.

Process model:

```
       browsers (4)
            │  WebSocket (one per player)
            ▼
     ┌────────────┐
     │   Hub      │  ──► route to room
     └────────────┘
            │
            ▼
     ┌────────────┐
     │ Room actor │  ──► owns GameState, Clock, Timers
     │ (1 per     │  ──► applies validated commands
     │  game)     │  ──► broadcasts filtered views
     └────────────┘
```

A `Hub` goroutine owns the map of active rooms. Each `Room` is its own goroutine with an inbox channel — an actor. Player WebSocket reader/writer goroutines forward to/from the room actor. The room actor is the **only** thing that mutates game state; this gives serialized access without locks and makes the hidden-info contract enforceable structurally.

Storage is in-process. Optional SQLite is a single file sidecar to the binary, gated behind `--persist`.

## 4. Data model & hidden-info enforcement

This is the safety-critical part of the project — the entire 暗棋 promise rests on it.

**Board:** 17×17 grid. Four 6×5 home zones, central cross of railroads, 4 行营 (camps) per zone, 大本营 (HQ) cells. Cell types are static and table-encoded: `Normal`, `Camp`, `Railroad`, `HQ`, `Frontline`.

**Piece:**

```go
type Piece struct {
    ID    PieceID  // server-generated, never sent to opponents
    Owner Seat     // North | East | South | West
    Rank  Rank     // 司令 ... 工兵, plus 地雷, 炸弹, 军旗
    Alive bool
}
```

**GameState (server-only — never serialized to a client as-is):**

```go
type GameState struct {
    Phase    Phase             // Setup | Playing | Ended
    Turn     Seat
    Pieces   map[PieceID]Piece
    BoardIdx [17][17]PieceID
    Clocks   map[Seat]Clock
    Revealed map[PieceID]bool
    History  []Event
}
```

### View filter — the safety primitive

```go
func (g *GameState) ViewFor(viewer Seat) ClientView
```

Rules:

- The viewer's own pieces → full identity always.
- Opponent or teammate pieces → identity only if `Revealed[id] == true` (e.g., piece died in combat, or 军旗 after the owning side's 司令 was killed). Otherwise client sees `{owner, rank: Unknown, alive: true}`.
- Combat events are filtered: combat that resolves reveals both ranks; events report only what the viewer is permitted to see.
- During setup, opponent boards always show `Unknown` everywhere.

**Structural guarantee:** clients never receive `GameState`. Outgoing-message construction lives in exactly one builder function that takes a `Seat` parameter, calls `ViewFor(seat)`, and only then serializes. Broadcast-to-all is impossible by construction. A unit test enumerates every event type × every viewer seat across a 100-move random game and asserts no piece-rank string appears for any non-revealed piece — this is the test that defends the 暗棋 contract.

## 5. Chat

**Channels:**

- **All-table** — all 4 players see it.
- **Team** — only you and your diagonal teammate. Critical channel given strict visibility — it's how teammates coordinate without sharing piece-level information.

**Mechanics:**

- Free text, UTF-8, 200-char cap, 5 messages / 10 seconds rate limit (server-enforced).
- Server stamps sender + channel + monotonic sequence; chat events are recorded in `History` so replays include the conversation.
- Predefined emote/quick-message set sent as enum codes (low-toxicity floor, language-agnostic).
- Per-player client-side mute toggle (mute opponent / mute team / mute all) — purely local.
- Room creator can disable team chat via room option for stricter games.

Out of scope: voice, private 1-on-1 with opponents, image/sticker uploads, profanity filter.

## 6. Network protocol

JSON over WebSocket, text frames. 20 s heartbeat ping; client considered dropped after 30 s silence.

**Envelope:** `{ "type": "<kind>", "seq": <int>, ...payload }`. `seq` is monotonic per direction and is used for idempotency on reconnect.

**Type sharing:** Go structs in `internal/protocol/` are the source of truth. `tygo` (or equivalent) generates `web/src/lib/protocol.ts` via `go generate`. CI fails if the generated file is stale. One protocol package, two consumers.

### Client → Server

| `type` | Phase | Payload |
|---|---|---|
| `room.join` | pre | `{ name, sessionToken? }` |
| `room.leave` | any | `{}` |
| `room.config` | lobby | host-only: `{ timeControl, allowTeamChat }` |
| `room.start` | lobby | host-only: `{}` once 4 seated |
| `setup.place` | setup | `{ pieceId, row, col }` (idempotent) |
| `setup.randomize` | setup | `{}` — server returns valid layout |
| `setup.submit` | setup | `{}` — locks player's setup |
| `move` | playing | `{ from, to, path? }` (path required for railroad multi-step) |
| `concede` | playing | `{}` |
| `vote.abandon` | paused | `{}` |
| `chat.send` | any | `{ channel, text? , emote? }` |

### Server → Client

| `type` | Payload |
|---|---|
| `room.state` | full lobby/room snapshot |
| `view` | filtered `ClientView` per recipient — sent on every state change |
| `event.move` | `{ from, to, mover }` (rank shown only if revealed) |
| `event.combat` | `{ at, attacker, defender, outcome, revealed }` |
| `event.flagCaptured` | `{ loser }` |
| `event.teamEliminated` | `{ losers, winners }` |
| `clock.tick` | `{ seat, msRemaining }` |
| `chat.msg` | `{ from, name, channel, text\|emote, ts }` (team msgs delivered only to teammate + sender) |
| `paused` / `resumed` | pause control |
| `error` | `{ code, message, refSeq }` — never crashes the connection |

**Sequencing & idempotency:** server tracks last applied client `seq` per (player, room); replays are no-ops returning the original result. Reconnect-and-resend is safe.

**Optimistic UI:** `setup.place` may render immediately and reconcile on `view`. Moves wait for server confirmation since combat outcomes need server adjudication.

**Auth:** no accounts. `room.join` issues a 32-byte random `sessionToken`, returned to client and stored in `localStorage`. Reconnection presents `(roomCode, sessionToken)` to recover the seat. Tokens expire on game end.

## 7. Referee / judge logic

Lives in `internal/game/`. Pure functions. Zero networking.

### Combat resolution

```
炸弹 vs anything           → both die
工兵 vs 地雷                → 地雷 dies, 工兵 lives
anything-else vs 地雷       → attacker dies, 地雷 lives
军旗 captured by anything   → 军旗 dies, attacker lives, game-end trigger
otherwise                  → higher rank wins, lower dies; equal ranks both die
```

Table-driven. Truth-table test covers every (attacker, defender) pair.

### Movement

- **地雷, 军旗:** never move.
- **炸弹:** 1 step orthogonally, or full railroad straight-line; cannot enter 大本营.
- **司令 through 排长:** 1 step orthogonally on normal cells; full straight-line on railroads (no corners).
- **工兵:** as above on normal cells; on railroads, BFS over the entire connected railroad graph (corners allowed) — full railroad mobility.

`LegalMoves(state, pieceId) []Move` enumerates legal destinations using a precomputed railroad-adjacency table.

### Cell rules

- **Camps (行营):** a piece standing on a 行营 cannot be attacked. Two pieces never share a camp.
- **HQ (大本营):** a piece in HQ cannot move out. Moving into an opponent HQ that contains the flag captures it.
- **Frontline:** the central row/column is shared territory; standard pass-through.

### Move application

`ApplyMove(state, move) (newState, []Event)`:

1. Re-validates legality server-side (defense in depth).
2. Empty destination → reposition; emit `event.move`.
3. Enemy destination → `Resolve`; mark dead pieces `Alive=false`; mark dead pieces and surviving combatant as `Revealed=true`; emit `event.combat`.
4. If 司令 dies → mark its owner's flag as `Revealed=true`. Opponents now see the flag's identity.
5. If 军旗 captured → owner is eliminated; their remaining pieces are removed from the board; their teammate plays alone. If both teammates lose flags, opposing team wins.

### Turn order

Anti-clockwise: North → West → South → East. Eliminated player's turn is skipped. If both members of one team are eliminated, the other team wins immediately.

### Win conditions

1. Both flags of one team captured → other team wins.
2. Both members of one team concede → other team wins.
3. Both members of one team time-forfeit (after reconnect window) → other team wins.
4. **Stalemate:** 50 consecutive moves without combat or flag capture → draw. Threshold is a per-room setting.

### Determinism

`ApplyMove` is pure. `setup.randomize` uses a seeded RNG; the seed is logged to `History`. Replaying `History` from an empty state must reproduce the final state byte-for-byte. A `TestReplayRoundTrip` integration test enforces this for every recorded test game.

### Testing strategy

- **Unit:** combat truth table; movement generators per piece × cell type; camp/HQ edge cases; railroad pathfinder including 工兵 cycle traversal.
- **Property:** fuzz random legal sequences; assert invariants (piece count never increases, flags stationary, eliminated pieces stay eliminated, view filter never reveals non-revealed ranks).
- **Scripted-game integration:** YAML scripts of moves with expected events. Bug reports are captured as new scripts.
- **Hidden-info leak test:** the regression net for the 暗棋 contract — see Section 4.

## 8. Frontend

### Rendering — DOM/SVG hybrid

Board is a 17×17 CSS grid of cell `<div>`s for layout/interaction. Pieces are SVG components absolutely positioned over the grid. Get CSS animations, keyboard a11y, and crisp scaling for free; canvas would be overkill for a turn-based 17×17 board.

### Component tree

```
App
├── LobbyScreen        — create/join, name entry, host config
├── SeatPicker         — pick North/East/South/West
├── SetupScreen
│   ├── Board (read-only opponent zones during setup)
│   ├── PiecePalette
│   ├── HomeZone       — drop target with live legality
│   ├── RandomizeBtn
│   └── SubmitBtn
├── PlayScreen
│   ├── Board
│   ├── ClockBar       — 4 clocks, current player highlighted
│   ├── EventLog       — combat log
│   ├── ChatPanel      — 公屏 / 队伍 tabs + emote bar
│   └── PausedOverlay
└── EndScreen          — winner, replay download (.json), play again
```

A single `gameStore` holds the current `ClientView` and is updated atomically on every `view`. Components subscribe via Svelte 5 runes (`$derived`); only changed cells re-render.

### Piece visuals

Junqi web apps usually look generic. Concrete plan:

- **Shape:** rounded-rectangle "chip" silhouette in SVG, ~52×40 px at 1×, proportions of a physical wooden 军棋 tile. Subtle bevel via SVG filter (inner shadow + light highlight).
- **Color:** red and blue per team. Each player on a team gets the same base color with a small directional accent stripe (north tile has top-edge accent, east has right-edge, etc.) so allied pieces at a glance.
- **Typography:** rank glyph in Song/Ming-style serif (Noto Serif SC, weight 700 or commissioned). Engraved-feel: dark glyph with subtle inset shadow filter so the character looks pressed into the tile.
- **Hidden pieces:** opponents see a uniform team-colored tile with a stylized 五角星 / military emblem center — no rank glyph, no rank-derived hints.
- **Revealed-by-combat:** dying pieces flip card-style to reveal their rank glyph before fading out. This is the dramatic moment in junqi and gets the polish.
- **Camps / HQ / railroads:** distinctive board art — 行营 as small octagon outlines, 大本营 as star, railroads as parallel-line tracks with sleeper marks.
- **Asset pipeline:** SVG component files, swappable. Start with a hand-tuned SVG + Noto Serif SC pass; if not striking enough, commission 12 ranks × 2 color variants = 24 SVGs. UI architecture is independent of the asset choice.

### Animations

- **Move:** 200 ms ease-out transform on the piece wrapper, FLIP-anchored by piece ID.
- **Combat:** ~600 ms — converge, brief shake, dying pieces flip-reveal, fade.
- **Flag capture:** screen-shake + centered "军旗已被夺取!" toast + flag glyph zoom before the end-game transition.
- **Clock low-time:** active clock pulses red under 10 s.
- Respect `prefers-reduced-motion` — fall back to instant transitions.

### Interaction

- Drag-drop via `@neodrag/svelte` for setup and moves.
- Click-to-select + click-to-move supported in parallel for touch and keyboard.
- Touch handlers covered by the same code path.

### Accessibility & i18n

- ARIA labels on cells, filtered through visibility rules — opponent's hidden piece announced as "敌方棋子", not its rank.
- Keyboard nav: arrow keys move focus, Enter selects, Enter again confirms.
- Strings in `lang/zh-CN.json`. Structure ready for English in v2.

### Sound

Optional. Per-player client-side toggle, defaults **off**, persisted in localStorage. Each of the 4 players controls their own sound independently. Subtle wood-tile click on move, soft drum on combat, low chime on flag capture.

## 9. Disconnect, reconnect, time control

### Disconnect

When a player's socket drops:

1. Room actor moves to `Paused` state, broadcasts `paused` with the deadline (90 s from now).
2. Clocks halt for everyone.
3. Dropped player has 90 s to reconnect by presenting `(roomCode, sessionToken)`. On success, server replays missed events from `History` since the player's last `seq`, and resumes.
4. If the deadline expires, the player's team forfeits and the game ends.
5. The other 3 players can `vote.abandon` to end the pause early; unanimous → forfeit immediately.

### Time control

Per-room configuration, set by host before start:

- **Setup phase clock:** total seconds for all setup actions (default 180).
- **Per-move clock:** seconds per move (default 30).
- **Increment:** seconds added back on move completion (default 5).
- **Bank:** optional total game-clock cap (default unlimited).

Clocks are server-authoritative. Client only displays `clock.tick` events. Time expiry during play = auto-forfeit for that player (treated as `concede`).

## 10. Repo layout

```
siguo/
├── cmd/siguo/                 — main, flag parsing, server boot
├── internal/
│   ├── protocol/              — Go structs + tygo config (source of truth)
│   ├── game/                  — pure game logic
│   ├── room/                  — Room actor
│   ├── hub/                   — Hub, lifecycle, room codes
│   ├── ws/                    — WebSocket adapter, heartbeat, reconnect
│   └── persist/               — optional SQLite layer
├── web/                       — Svelte app
│   ├── src/
│   │   ├── lib/protocol.ts    — generated, do not edit
│   │   ├── lib/store.ts
│   │   ├── components/
│   │   └── routes/
│   ├── package.json
│   └── vite.config.ts
├── assets/                    — SVG piece art, sounds, fonts
├── docs/plans/
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── go.mod
```

## 11. Build, deploy, ops

### Build pipeline

- `make codegen` — regenerates `web/src/lib/protocol.ts` from Go structs.
- `make web` — `pnpm install && pnpm build` → `web/dist/`.
- `make build` — `make codegen web`, then `go build -o bin/siguo ./cmd/siguo` (embeds `web/dist/`).
- `make docker` — multi-stage Dockerfile: `node:lts-alpine` → web build, `golang:alpine` → go build, `gcr.io/distroless/static` → runtime. Single `COPY` of the binary. Image size target <20 MB. `EXPOSE 8080`. No shell.

CI runs `make codegen` and fails if the generated TS file differs.

### Configuration

| Flag | Env | Default | Purpose |
|---|---|---|---|
| `--addr` | `SIGUO_ADDR` | `:8080` | listen address |
| `--persist` | `SIGUO_PERSIST` | `false` | enable SQLite |
| `--db-path` | `SIGUO_DB_PATH` | `./siguo.db` | SQLite location |
| `--max-rooms` | `SIGUO_MAX_ROOMS` | `100` | hard cap |
| `--log-level` | `SIGUO_LOG_LEVEL` | `info` | slog levels |
| `--allow-origin` | `SIGUO_ALLOW_ORIGIN` | (same-origin only) | CORS for dev |
| `--public-url` | `SIGUO_PUBLIC_URL` | autodetected | share-link generation |
| `--metrics` | `SIGUO_METRICS` | `false` | enable `/metrics` |

LAN: `./siguo`, share `http://<host>:8080`. Cloud: `docker run -p 80:8080 siguo:latest --persist`.

### Operations

- **Logging:** structured `slog` (JSON in cloud, text in LAN). Per-room context fields. **No piece identities in logs ever** — same filter discipline as the wire protocol.
- **Metrics:** Prometheus at `/metrics` when `--metrics`. Counters: rooms created, games completed, disconnects, reconnects, time-forfeits. Histograms: game duration, setup duration, move latency.
- **Health:** `/healthz`, `/readyz`.
- **Graceful shutdown:** SIGTERM stops new rooms, broadcasts shutdown notice, persists in-flight rooms if `--persist`, exits.
- **Resource bounds:** 1000-event cap per `History` (paginate), 200-char chat, 5/10s chat rate, 90 s reconnect window, `--max-rooms` cap.

### Security (cloud mode)

- TLS at reverse proxy (nginx/Caddy/Traefik). siguo speaks plaintext HTTP; cloud users put it behind a proxy. `docker-compose.yml` ships a Caddy sidecar example.
- Origin check on WebSocket upgrade when `--allow-origin` is set.
- Session tokens 32 random bytes, never logged, expire on game end.
- Frame size limits: 64 KB max WebSocket frame.
- SQLite parameterized queries only.
- No user-uploaded content. Chat text rendered as text, never HTML.

## 12. Risks

| Risk | Mitigation |
|---|---|
| Hidden-info leak via misrouted message | Single `ViewFor` choke point; per-recipient builder; CI leak-test |
| Tygo codegen drift between Go and TS | CI assertion that generated file matches |
| Endless games, especially low-piece endgames | 50-move no-progress draw rule, configurable |
| Race on "both flags fall same turn" | Atomic event resolution inside actor; turn-end sweep checks all teams |
| Reconnect fails on flaky network | 90 s window + idempotent commands by `seq` |
| SVG art looks generic | Plan for commissioned art if placeholder pass falls short; UI doesn't depend on asset choice |
| Server clock drift affecting fairness | Clocks server-authoritative; client only displays |

## 13. Implementation order

1. `internal/game/` — referee, movement, combat — with full test coverage. **No networking yet.**
2. `internal/protocol/` + tygo codegen.
3. `internal/room/` actor with in-memory state, exercised by tests.
4. `internal/hub/` + WebSocket + REST — first end-to-end happy path with a CLI test client.
5. Svelte scaffold, lobby, board rendering with placeholder SVG pieces.
6. Setup phase end-to-end.
7. Play phase end-to-end with combat animations and chat.
8. Disconnect/reconnect + clocks.
9. Polished piece art pass + sound.
10. Docker, CI, release.
