# go-runs-rest-walk — project plan

## Repos

| Role                      | Path                                                                                       |
| ------------------------- | ------------------------------------------------------------------------------------------ |
| **Implement**             | `go-runs-rest-walk` — `github.com/hornflakes/go-runs-rest-walk`                            |
| **Reference (read only)** | `../tyrone-biggums` — peek for ideas only; do not copy wire names or winner format blindly |
| **Default port**          | 37373, path `/`                                                                            |

Attribution: inspired by tyrone-biggums; own implementation; cite in thesis.

---

## Current focus: Java server

Port **this repo’s Go server** to **Java + Netty**: same game, same wire protocol, same load harness and metrics. Use **Java 25**, **Netty 4** (HTTP upgrade + WebSocket on `/`), and a **blocking ~60 Hz game loop** on virtual threads (same server phases as Go: pair → ready → queue → tick → GameOver).

| Decision        | Choice                                                                                                                            |
| --------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| Location        | `java/` at repo root                                                                                                              |
| Build           | **Gradle**                                                                                                                        |
| Language        | **Java 25** (virtual threads for pairing, ready, queue reader, game loop)                                                         |
| HTTP / WS       | **Netty 4** — `ServerBootstrap`, WebSocket upgrade, text frames only (gorilla / `ws` analogue)                                    |
| Concurrency     | Netty event loop for accept/I/O framing; virtual threads for pair dispatch, per-socket outbound drain, queue reader, `Game.run()` |
| WS idle timeout | **None** during stagger (match Go — clients may wait a long time before pairing)                                                  |
| Packages        | **`com.hornflakes.gorunsrestwalk.*`** — `.server`, `.gameloop`, `.stats`, `.logger` (mirrors Go `internal/*`)                     |
| Logging         | **Custom colored logger** — same levels/events as Go (`milestone`, `info`, `warn`, `softError`, `hardError`)                      |
| Source of truth | `go/internal/server`, `go/internal/gameloop`, `go/internal/stats`, `go/internal/logger`                                           |
| Load harness    | **Go `cmd/load` only** — do **not** implement tyrone `game_player`                                                                |
| Listen          | **Port 37373 only** (no explicit `0.0.0.0`; same as current TS)                                                                   |
| Tyrone          | Optional peek only — **not** wire/protocol source                                                                                 |

---

## Go baseline (done)

Reference implementation and primary comparison arm.

- Game server: pair → ready → ~60 Hz loop → GameOver + per-game histogram
- Tyrone-aligned simulation: `go/internal/gameloop/spec.go`
- `cmd/bot` — two-client smoke
- `cmd/load` — burst + stagger + winner stdout
- Benchmark pipeline: `mkdir -p` → tee `.res` → `scripts/load-parse-results` → `*-rolling.csv` + `runs.csv` → `scripts/benchmark.py`

Example Go results (WSL, 1 core, burst, stagger 50, newspec): `results-wsl-burst-1thread/runs.csv` (10k → `bucket0_ratio_overall` ~0.85, `max_active_games` ~2655).

---

## TypeScript baseline (done)

Callback-style Node server in `ts/` (mirrors Go package layout).

- `npm run start` → listen `:37373`
- Same wire protocol and `spec` constants as Go
- Formal run: **`ts-N10000-G20-run1`** (see `results-wsl-burst-1thread/runs.csv` — e.g. `bucket0_ratio_overall` ~0.57, `max_active_games` ~1473 on same harness)

Use Go + TS as comparison arms when Java is ready.

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

**Not** tyrone’s `ReadyUp` / `Play` / `Fire` / `winner(N)___…`.

Constants: mirror `go/internal/gameloop/spec.go` and histogram buckets in `go/internal/stats/stats.go`.

Winner stdout (load parses this):

```text
winner=<id> histogram=b0,...,b7 active_games=<N>
```

---

## Java implementation phases

Work through in order. Verify each phase before the next.

| Phase | Build                                                                                                                                   | Verify                                         |
| ----- | --------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------- |
| **0** | `java/` Gradle project, Java 25 toolchain, Netty 4, `com.hornflakes.gorunsrestwalk.logger` (Go/TS parity), entry `Main`, listen `37373` | Process starts; colored `server listening` log |
| **1** | Message types + JSON parse/build (Go enum values)                                                                                       | Unit test: `{"type":3}` → Shoot                |
| **2** | WebSocket channel handler: Netty WS, text frames only, send/receive + Go-equivalent socket logs                                         | Connect smoke if helpful                       |
| **3** | Pairing server (synchronized + waiting socket → pair callback)                                                                          | Log or test: two clients paired                |
| **4** | Ready handshake (`WaitForReady` equivalent, timeout)                                                                                    | `go run ./cmd/bot` × 2 terminals               |
| **5** | Game loop: queue, physics, collisions, 16 ms tick pacing, `activeGames`, histogram                                                      | Winner line byte-for-byte compatible with Go   |
| **6** | Endurance run with **`cmd/load`**                                                                                                       | `java-N10000-G20-run1-load.res` → parse → plot |

Map Go → Java (suggested):

| Go                                        | Java                                                                                                |
| ----------------------------------------- | --------------------------------------------------------------------------------------------------- |
| `cmd/server/main.go`                      | `com.hornflakes.gorunsrestwalk.Main`                                                                |
| `internal/server/message.go`              | `com.hornflakes.gorunsrestwalk.server.Message` (+ helpers)                                          |
| `internal/server/socket.go` + `server.go` | `com.hornflakes.gorunsrestwalk.server.Socket`, `com.hornflakes.gorunsrestwalk.server.PairingServer` |
| `internal/gameloop/*`                     | `com.hornflakes.gorunsrestwalk.gameloop.*`                                                          |
| `internal/stats/*`                        | `com.hornflakes.gorunsrestwalk.stats.*`                                                             |
| `internal/logger/*`                       | `com.hornflakes.gorunsrestwalk.logger.*`                                                            |

### Phase 0 Gradle sketch

- `settings.gradle.kts`: `rootProject.name = "go-runs-rest-walk"`
- Java 25 via toolchain in `build.gradle.kts`
- Dependencies (Netty 4): `io.netty:netty-all` (or modular `netty-handler`, `netty-codec-http`, `netty-codec-http2` as needed) — pin version in `build.gradle.kts`
- Run task: `./gradlew run` → `com.hornflakes.gorunsrestwalk.Main`
- **`com.hornflakes.gorunsrestwalk.logger`**: port `go/internal/logger` + `ts/src/logger` — static helpers (`info`, `warn`, `milestone`, `softError`, `hardError`, `player`, `playerWithAddr`, `pairPrefix`) and pair-scoped `Logger` class; ANSI colors; `event | detail` format
- **`Main`**: on bind, log `server listening` | `addr=:37373` via logger (smoke that phase 0 logger works)

### Logging parity (phase 2+)

Logger exists from phase 0; wire the events below into socket/server/game code as each phase lands.

Match Go event strings and levels in `com.hornflakes.gorunsrestwalk.server` socket I/O:

- `websocket connected` (info)
- `websocket upgrade failed` (hardError)
- `websocket read ended` (warn)
- `websocket message unmarshal failed` (softError)
- `websocket message marshal failed` (softError)
- `websocket message write failed` (hardError)

Pair/game (from `cmd/server` + `gameloop`):

- `websockets paired`, `websockets ready handshake failed`, `websockets ready handshake ok`
- `game on`, `player shot` (verbose), `game over`
- `server listening` | `addr=:37373`

---

## Benchmark harness (`cmd/load`)

**Always use Go load** against whichever server is running (Go, TS, or Java). Same flags for all languages.

### Commands

```bash
# Create results folder first (tee does not mkdir)
mkdir -p ../results/<folder>

# Terminal 1 — Java server (when ready)
cd java && ./gradlew run

# Terminal 2 — load (stdout → .res)
cd go && go run ./cmd/load/ \
  -connections 10000 -games 20 -stagger 50 -burst \
  2>&1 | tee ../results/<folder>/java-N10000-G20-run1-load.res
```

Comparison runs (same flags, swap server):

```bash
cd go && taskset -c 0 go run ./cmd/server/
cd go && taskset -c 0 go run ./cmd/load/ \
  -connections 10000 -games 20 -stagger 50 -burst \
  2>&1 | tee ../results/<folder>/go-N10000-G20-newspec-run1-load.res

cd ts && npm run start
cd go && taskset -c 0 go run ./cmd/load/ \
  -connections 10000 -games 20 -stagger 50 -burst \
  2>&1 | tee ../results/<folder>/ts-N10000-G20-run1-load.res
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

Milestone line in `.res`:

```text
load | url=ws://127.0.0.1:37373/ connections=10000 games=20 stagger=50ms fire=burst-5ms-40shards
```

### Primary comparison run

| Setting                                       | Value                                                 |
| --------------------------------------------- | ----------------------------------------------------- |
| Run id                                        | **`java-N10000-G20-run1`**                            |
| Connections                                   | **10000**                                             |
| Games per connection                          | **20**                                                |
| Stagger                                       | **50** ms                                             |
| Fire                                          | **burst** (5 ms, 40 shards)                           |
| Simulation                                    | Tyrone-aligned `spec.go` constants                    |
| CPU pinning (optional, document in lab notes) | `taskset -c 0` on server + load when comparing fairly |

**Metrics:** `bucket0_ratio_overall`, `max_active_games`, `mean_active_games`, `games_failed` (0 for a valid run).

Parser `language` column: prefix `java` (regex `[a-z]+` — no parser change).

---

## Results layout

```text
results/<folder>/
  go-N10000-G20-newspec-run1-load.res
  go-N10000-G20-newspec-run1-rolling.csv
  ts-N10000-G20-run1-load.res
  ts-N10000-G20-run1-rolling.csv
  java-N10000-G20-run1-load.res
  java-N10000-G20-run1-rolling.csv
  runs.csv
  rolling-active-games.png    # benchmark.py --x active_games
  rolling-over-time.png       # benchmark.py --x window_end
```

Naming: `{lang}-N{connections}-G{games}[-tag]-run{n}-load.res`

### Parse

```bash
mkdir -p ../results/<folder>
cd scripts && go run ./load-parse-results/main.go ../results/<folder>/java-N10000-G20-run1-load.res
```

### Plot

```bash
cd scripts && python3 benchmark.py ../results/<folder> --x active_games
python3 benchmark.py ../results/<folder> --x window_end
```

Overlay Go + TS + Java rolling files in one folder.

---

## Environment notes (document per run)

Record in thesis / lab book:

- OS (WSL / Windows / Linux), Java version, Netty version
- `taskset` if used
- git commit
- load flags (`stagger`, burst settings)
- anomalies (`games_failed`, heartbeat bursts under saturation)

### Threats to mention when writing up

- Netty event-loop I/O + virtual-thread game scheduling vs Go goroutines + gorilla
- Netty outbound writes must respect channel event-loop thread confinement
- `Thread.sleep` / nanosleep tick pacing vs Go `time.Sleep`
- Same-machine `cmd/load` competing for CPU with server
- `active_games` on X-axis for plots, not raw connection count

---

## Out of scope

- tyrone Rust / TS `game_player`
- RxJS server variant
- tyrone winner format `winner(N)___…`
- Spring, WebFlux, other HTTP stacks
- Browser client
- Server log capture to `.res`
- `games.csv`
- Auto-sweep runner until manual runs are painful

### Deferred (after Java port)

- `pidstat` / CPU-RAM sampling (tyrone `measure` style)
- `perf` + FlameGraph (Java + Go)
- Summary plot from `runs.csv` (connections vs median bucket0)
- Full 3× N ladder (5000 / 10000 / …) across all languages

---

## Agent instructions

- Implement only in `go-runs-rest-walk`.
- Port from **Go** (`go/internal/*`); peek `../tyrone-biggums` only when stuck.
- Do not change load winner-line format or parser regex without updating `scripts/load-parse-results`.
- Formal Java run: **`java-N10000-G20-run1`** with load flags in table above.
- **`mkdir -p`** before `tee` when capturing `.res`.

---

## New chat starter

```
Java server port in go-runs-rest-walk/java/ (Java 25, Gradle, Netty 4 WebSocket, virtual threads).
Same game and wire protocol as Go; port server phases from go/internal/*.
Use go/cmd/load for benchmarks.
Packages: com.hornflakes.gorunsrestwalk.server, .gameloop, .stats, .logger.
Reference: go/internal/* (source of truth), ts/ (sibling port), tyrone-biggums (peek only).
Parse/plot: scripts/load-parse-results, scripts/benchmark.py.
Target run: java-N10000-G20-run1, -connections 10000 -games 20 -stagger 50 -burst.
@.cursor/PLAN.md
```
