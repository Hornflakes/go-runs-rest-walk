# go-runs-rest-walk — project plan

## Repos

| Role                      | Path                                                                                       |
| ------------------------- | ------------------------------------------------------------------------------------------ |
| **Implement**             | `go-runs-rest-walk` — `github.com/hornflakes/go-runs-rest-walk`                            |
| **Reference (read only)** | `../tyrone-biggums` — peek for ideas only; do not copy wire names or winner format blindly |
| **Default port**          | 37373, path `/`                                                                            |

Attribution: inspired by tyrone-biggums; own implementation; cite in thesis.

---

## Current focus: TypeScript server

Port **this repo’s Go server** to TypeScript. Use **callback-style** I/O (event listeners + tick timer), not RxJS.

| Decision        | Choice                                                                                |
| --------------- | ------------------------------------------------------------------------------------- |
| Location        | `ts/` at repo root                                                                    |
| Runtime         | Node.js + [`ws`](https://github.com/websockets/ws)                                    |
| Source of truth | `go/internal/server`, `go/internal/gameloop`, `go/internal/stats`                     |
| Load harness    | **Go `cmd/load` only** — do **not** implement tyrone `game_player`                    |
| Tyrone TS       | Optional peek (`callback-server`, `game-callback`) — **not** the wire/protocol source |

---

## Go baseline (done)

Use as the reference implementation and comparison arm.

- Game server: pair → ready → ~60 Hz loop → GameOver + per-game histogram
- Tyrone-aligned simulation: `go/internal/gameloop/spec.go` (±2500 spawn, 100×100 players, 35×3 bullets, 180/300 ms fire, 16 ms tick target)
- `cmd/bot` — two-client smoke
- `cmd/load` — burst + stagger + winner stdout (`client.go`, `burst.go`, thin `main.go`)
- Benchmark pipeline: tee `.res` → `scripts/load-parse-results` → `*-rolling.csv` + `runs.csv` → `scripts/benchmark.py`

Example Go results (WSL, 1 core, burst, stagger 50, newspec): see `results-wsl-burst-1thread/runs.csv` (e.g. 10k → `bucket0_ratio_overall` ~0.85, `max_active_games` ~2655).

---

## Wire protocol (must match Go)

JSON: `{"type": n, "msg"?: "..."}`

| type | Name     | Notes                                                            |
| ---- | -------- | ---------------------------------------------------------------- |
| 0    | Hello    | Server → client; `msg`: `playerId=<id>`                          |
| 1    | Ready    | Handshake; client echoes Ready after parsing `enemyId=`          |
| 2    | GameOn   | Server → both when match starts                                  |
| 3    | Shoot    | Client → server while playing                                    |
| 4    | GameOver | Winner: `winner=<playerId> histogram=b0,...,b7 active_games=<N>` |
| 4    | GameOver | Loser: `loser` (load ignores for `games_over`)                   |

**Not** tyrone’s `ReadyUp` / `Play` / `Fire` / `winner(N)___…` — load and parser expect **your** Go format.

Constants: mirror `go/internal/gameloop/spec.go` and histogram buckets in `go/internal/stats/stats.go`.

---

## TypeScript implementation phases

Work through in order. Verify each phase before the next.

| Phase | Build                                                                              | Verify                                       |
| ----- | ---------------------------------------------------------------------------------- | -------------------------------------------- |
| **0** | `ts/package.json`, TypeScript, `ws`, entry script, listen `37373`                  | Process starts; port open                    |
| **1** | Message types + JSON parse/build (Go enum values)                                  | Unit test: `{"type":3}` → Shoot              |
| **2** | WebSocket wrapper: text frames only, read/write                                    | Minimal connect / echo if helpful            |
| **3** | Pairing server (mutex + waiting socket → `[s0,s1]` on channel)                     | Log or test: two clients paired              |
| **4** | Ready handshake (`WaitForReady` equivalent, timeout)                               | `go run ./cmd/bot` × 2 terminals             |
| **5** | Game loop: queue, physics, collisions, 16 ms tick pacing, `activeGames`, histogram | Winner line byte-for-byte compatible with Go |
| **6** | Endurance run with **`cmd/load`**                                                  | `ts-N10000-G20-run1-load.res` → parse → plot |

Map Go files → TS modules (suggested):

- `internal/server/message.go` → `ts/src/message.ts`
- `internal/server/socket.go` + `server.go` → `ts/src/server/`
- `internal/gameloop/*` → `ts/src/gameloop/`
- `internal/stats/*` → `ts/src/stats/`
- `cmd/server/main.go` → `ts/src/index.ts` (or `main.ts`)

---

## Benchmark harness (`cmd/load`)

**Always use Go load** against whichever server is running (Go or TS). Same flags for both languages.

### Commands

```bash
# Terminal 1 — TypeScript server (when ready)
cd ts && npm run start   # document script in package.json

# Terminal 2 — load (stdout → .res)
cd go && go run ./cmd/load/ \
  -connections 10000 -games 20 -stagger 50 -burst \
  2>&1 | tee ../results/<folder>/ts-N10000-G20-run1-load.res
```

Go comparison run (same flags, swap server):

```bash
cd go && taskset -c 0 go run ./cmd/server/
cd go && taskset -c 0 go run ./cmd/load/ \
  -connections 10000 -games 20 -stagger 50 -burst \
  2>&1 | tee ../results/<folder>/go-N10000-G20-run1-load.res
```

### Load flags

| Flag              | Default     | Meaning                                                      |
| ----------------- | ----------- | ------------------------------------------------------------ |
| `-connections`    | 10          | WebSocket clients (even N → ~N/2 concurrent games)           |
| `-games`          | 0           | Games per client; `0` = until Ctrl+C; formal runs use **20** |
| `-host`           | `127.0.0.1` | Server host                                                  |
| `-port`           | `37373`     | Server port                                                  |
| `-path`           | `/`         | WS path                                                      |
| `-stagger`        | `50`        | Ms delay before client `id` starts (`id * stagger`)          |
| `-fire`           | `200`       | Per-client shoot interval (ms); **ignored when `-burst`**    |
| `-burst`          | `true`      | Global fire loop (tyrone-like); use for formal runs          |
| `-burst-interval` | `5`         | Ms between shard ticks                                       |
| `-burst-shards`   | `40`        | Shard count; client `id % shards`                            |

Milestone line in `.res` (parsed into `runs.csv`):

```text
load | url=ws://127.0.0.1:37373/ connections=10000 games=20 stagger=50ms fire=burst-5ms-40shards
```

### Primary comparison run

| Setting                                       | Value                                                 |
| --------------------------------------------- | ----------------------------------------------------- |
| Connections                                   | **10000**                                             |
| Games per connection                          | **20**                                                |
| Stagger                                       | **50** ms                                             |
| Fire                                          | **burst** (5 ms, 40 shards)                           |
| Simulation                                    | Tyrone-aligned `spec.go` constants                    |
| CPU pinning (optional, document in lab notes) | `taskset -c 0` on server + load when comparing fairly |

**Metrics to record:** `bucket0_ratio_overall`, `max_active_games`, `mean_active_games`, `games_failed` (must be 0 for a valid run).

Winner lines on stdout only (load prints; server logs optional in terminal, no server tee).

```text
winner=<id> histogram=b0,...,b7 active_games=<N>
```

Approximate winner count: `connections × games_per_conn` (one line per completed game from winning client).

---

## Results layout

```text
results/
  go-N10000-G20-newspec-run1-load.res
  go-N10000-G20-newspec-run1-rolling.csv
  ts-N10000-G20-run1-load.res
  ts-N10000-G20-run1-rolling.csv
  runs.csv
  rolling-active-games.png    # benchmark.py --x active_games
  rolling-over-time.png       # benchmark.py --x window_end (default)
```

Naming: `{lang}-N{connections}-G{games}-run{n}-load.res`  
Parser sets `language` from prefix (`go` / `ts`).

### Parse

```bash
cd scripts && go run ./load-parse-results/main.go ../results/<folder>/ts-N10000-G20-run1-load.res
```

Outputs:

- `{run_id}-rolling.csv` — columns: `window_end`, `active_games`, `bucket0_ratio`
- append `runs.csv` — includes `max_active_games`, `mean_active_games`, etc.

### Plot

```bash
cd scripts && python3 benchmark.py ../results/<folder> --x active_games
python3 benchmark.py ../results/<folder> --x window_end
```

Overlay Go + TS rolling files in one folder for one chart.

---

## Environment notes (document per run)

Record in thesis / lab book, not in code:

- OS (WSL / Windows), `taskset` if used
- git commit
- `stagger`, `fire` / burst settings
- anomalies (frozen heartbeat bursts under saturation = backlog, not necessarily failure)

### Threats to mention when writing up

- Node event-loop timer vs Go sleep loop (tick pacing)
- Same-machine `cmd/load` competing for CPU with server
- `active_games` on X-axis for plots, not raw connection count

---

## Out of scope

- tyrone Rust `game_player` / TS `game_player` in `typescript/src/test/`
- RxJS server variant
- tyrone winner format `winner(N)___…`
- Browser client / `browser-index.ts`
- Server log capture to `.res` (watch server terminal only)
- `games.csv` (`.res` is the archive)
- Auto-sweep runner until manual runs are painful

### Deferred (after TS port)

- `pidstat` / CPU-RAM sampling (tyrone `measure` style)
- `perf` + FlameGraph
- Summary plot from `runs.csv` (connections vs median bucket0)

---

## Agent instructions

- Implement only in `go-runs-rest-walk`.
- Port from **Go**; peek `../tyrone-biggums` only when stuck on callback WS patterns.
- Do not change load winner-line format or parser regex without updating `scripts/load-parse-results`.
- Formal TS run: **`ts-N10000-G20-run1`** with load flags in table above.

---

## New chat starter

```
TypeScript server port in go-runs-rest-walk/ts/ (Node + ws, callback style).
Match Go wire protocol and spec.go; use go/cmd/load for benchmarks.
Reference: go/internal/* (source of truth), tyrone-biggums/typescript (peek only).
Parse/plot: scripts/load-parse-results, scripts/benchmark.py.
Target run: ts-N10000-G20, -connections 10000 -games 20 -stagger 50 -burst.
@.cursor/PLAN.md
```
